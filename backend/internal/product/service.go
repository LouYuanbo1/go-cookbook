package product

import (
	"context"
	"fmt"
	"go-cookbook/internal/common/dto"
	"go-cookbook/internal/common/utils/img"
	"go-cookbook/internal/common/utils/tempfs"
	"go-cookbook/internal/model"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"github.com/LouYuanbo1/go-webservice/gormc"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductService interface {
	//创建和更新产品需要原子性索引使用事务和更复杂的结构体，
	Create(ctx context.Context, req *dto.CreateProductReq) error
	//查询对一致性要求不高,可以使用两次分别查询产品和对应菜品,减少复杂性
	//使用First获取基础信息,之后使用Find获取对应菜品,考虑到食材可能有极多对应菜品，e.g"盐",所以这里使用游标
	FirstByCode(ctx context.Context, code string) (*dto.ProductResp, error)
	//更新产品需要原子性索引使用事务和更复杂的结构体，
	Update(ctx context.Context, req *dto.UpdateProductReq) error
	Delete(ctx context.Context, code string) error

	FindDishesByCode(
		ctx context.Context,
		productCode string,
		cursor uint64,
		limit int,
	) (*dto.CursorResp[dto.DishCardResp, uint64], error)
	Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error
	Export(gctx *gin.Context, batchSize int) error
}

type productService struct {
	db        *gormc.CacheDB
	imgUtil   img.ImgUtil
	tempfs    tempfs.TempFs
	tempDir   string
	uploadDir string
}

func NewProductService(
	db *gormc.CacheDB,
	imgUtil img.ImgUtil,
	tempfs tempfs.TempFs,
	tempDir string,
	uploadDir string,
) ProductService {
	return &productService{
		db:        db,
		imgUtil:   imgUtil,
		tempfs:    tempfs,
		tempDir:   tempDir,
		uploadDir: uploadDir,
	}
}

func (ps *productService) registerImages(
	imageFileHeaders []*multipart.FileHeader,
	productCode string,
) (images []model.ProductImage, mapIdImageName map[string]string, err error) {
	images = make([]model.ProductImage, 0, len(imageFileHeaders))
	mapIdImageName = make(map[string]string)

	for i, fileHeader := range imageFileHeaders {

		ext := filepath.Ext(fileHeader.Filename)
		imageName := fmt.Sprintf("%s_%d%s", productCode, i, ext)

		err := ps.imgUtil.ProcessFileHeader(
			fileHeader,
			ps.tempDir,
			imageName,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("处理图片失败: %w", err)
		}

		/*
			// 从 URL 中提取实际文件名（带扩展名）
			// URL 格式: /tempDir/PROD001_0.jpg
			actualFilename := path.Base(url)

			id, err := ps.tempfs.RegisterTempFile(filepath.Join(ps.tempDir, actualFilename))
			if err != nil {
				return nil, nil, fmt.Errorf("注册临时文件失败: %w", err)
			}
			mapIdImageName[id] = actualFilename
		*/

		id, err := ps.tempfs.RegisterTempFile(filepath.Join(ps.tempDir, imageName))
		if err != nil {
			return nil, nil, fmt.Errorf("注册临时文件失败: %w", err)
		}
		mapIdImageName[id] = imageName

		images = append(images, model.ProductImage{
			ProductCode: productCode,
			// 排序,用于显示顺序
			SortOrder: i,
			ImageURL:  fmt.Sprintf("/%s", filepath.ToSlash(filepath.Join(ps.uploadDir, imageName))),
		})
	}
	return images, mapIdImageName, nil
}

func (ps *productService) Create(ctx context.Context, req *dto.CreateProductReq) error {
	// 处理图片
	images, mapIdImageName, err := ps.registerImages(req.Images, req.ProductCode)
	if err != nil {
		return fmt.Errorf("注册图片失败: %w", err)
	}

	err = ps.db.Transaction(ctx, func(tx *gormx.Executor) error {
		// 第一步：创建产品
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "product_code"},
			},
			UpdateAll: true,
		}).Create(ctx, &model.Product{
			ProductCode:    req.ProductCode,
			IngredientCode: req.IngredientCode,
			Name:           req.Name,
			Amount:         req.Amount,
			Unit:           req.Unit,
			Description:    req.Description,
			Price:          req.Price,
			AllergenType:   req.AllergenType,
		}); err != nil {
			return fmt.Errorf("创建产品失败: %w", err)
		}
		// 第二步：创建产品图片关系
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "product_code"},
				{Name: "sort_order"},
			},
			UpdateAll: true,
		}).CreateInBatches(ctx, &images, 10); err != nil {
			return fmt.Errorf("创建产品图片关系失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("创建产品失败: %w", err)
	}

	// 第三步：将临时文件移动到正式目录
	for id, imageName := range mapIdImageName {
		if err := ps.tempfs.PromoteTempFile(id, ps.uploadDir, imageName); err != nil {
			return fmt.Errorf("移动临时文件失败: %w", err)
		}
	}

	return nil
}

func (ps *productService) FirstByCode(ctx context.Context, code string) (*dto.ProductResp, error) {
	var product model.Product
	key := fmt.Sprintf("product:%s", code)
	err := ps.db.Query(ctx, key, &product, func(ctx context.Context, db *gormx.DB, val *model.Product) error {
		return db.First(ctx, val, &model.Product{ProductCode: code})
	})
	if err != nil {
		return nil, fmt.Errorf("查询产品基本信息失败: %w", err)
	}

	productResp := dto.ProductResp{
		ProductCode:    product.ProductCode,
		IngredientCode: product.IngredientCode,
		Name:           product.Name,
		Amount:         product.Amount,
		Unit:           product.Unit,
		Description:    product.Description,
		Price:          product.Price,
		AllergenType:   product.AllergenType,
	}

	var images []model.ProductImage
	err = ps.db.GetXDB().Order("sort_order").Find(
		ctx, &images,
		&model.ProductImage{ProductCode: code},
	)
	if err != nil {
		return nil, fmt.Errorf("查询产品图片关系失败: %w", err)
	}

	productResp.Images = make([]dto.ImageResp, 0, len(images))
	for _, img := range images {
		productResp.Images = append(productResp.Images, dto.ImageResp{
			ID: img.ID,
			// 排序,用于显示顺序
			SortOrder: img.SortOrder,
			ImageURL:  img.ImageURL,
		})
	}

	return &productResp, nil
}

func (ps *productService) Update(ctx context.Context, req *dto.UpdateProductReq) error {
	// 处理新图片注册
	upsertImages, mapIdImageName, err := ps.registerImages(req.NewImages, req.ProductCode)
	if err != nil {
		return fmt.Errorf("注册图片失败: %w", err)
	}

	// 处理更新图片
	if len(req.UpdatedImages) > 0 {
		for _, img := range req.UpdatedImages {
			upsertImages = append(upsertImages, model.ProductImage{
				ID:          img.ID,
				ProductCode: req.ProductCode,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
		}
	}

	// 处理删除图片
	deletedImageURLs := make([]string, 0)
	if len(req.DeletedImageIDs) > 0 {
		err := ps.db.GetXDB().Model(&model.ProductImage{}).Where("id IN ?", req.DeletedImageIDs).Pluck(ctx, "image_url", &deletedImageURLs)
		if err != nil {
			return fmt.Errorf("查询删除图片失败: %w", err)
		}
	}

	err = ps.db.Transaction(ctx, func(tx *gormx.Executor) error {

		//第一步:更新产品基本信息
		err := tx.StructFilter(&model.Product{ProductCode: req.ProductCode}).
			Updates(ctx, &model.Product{
				IngredientCode: req.IngredientCode,
				Name:           req.Name,
				Amount:         req.Amount,
				Unit:           req.Unit,
				Description:    req.Description,
				Price:          req.Price,
				AllergenType:   req.AllergenType,
			})
		if err != nil {
			return fmt.Errorf("更新产品基本信息失败: %w", err)
		}

		// 第二步：删除图片
		if len(req.DeletedImageIDs) > 0 {
			if err := tx.Delete(ctx, &model.ProductImage{}, req.DeletedImageIDs); err != nil {
				return fmt.Errorf("删除产品图片关系失败: %w", err)
			}
		}

		// 第三步：更新或插入图片关系
		if len(upsertImages) > 0 {
			if err := tx.Clauses(
				clause.OnConflict{
					Columns: []clause.Column{
						{Name: "id"},
					},
					DoUpdates: clause.AssignmentColumns([]string{"sort_order"}),
				},
			).CreateInBatches(
				ctx,
				&upsertImages,
				10,
			); err != nil {
				return fmt.Errorf("创建新图片失败: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("更新菜品失败: %w", err)
	}

	// 持久化新图片
	for id, imageName := range mapIdImageName {
		if err := ps.tempfs.PromoteTempFile(id, ps.uploadDir, imageName); err != nil {
			return fmt.Errorf("移动临时文件失败: %w", err)
		}
	}

	// 第五步:删除图片文件
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	for _, url := range deletedImageURLs {
		filePath := filepath.Join(wd, url)
		fmt.Printf("删除图片路径: %s\n", filePath)
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
}

func (ps *productService) Delete(ctx context.Context, code string) error {
	var images []model.ProductImage
	err := ps.db.Transaction(ctx, func(tx *gormx.Executor) error {
		// 第一步：查询所有关联的图片
		err := tx.Find(ctx, &images, &model.ProductImage{ProductCode: code})
		if err != nil {
			return fmt.Errorf("查询产品图片关系失败: %w", err)
		}

		// 第二步：删除产品图片关系
		if err := tx.StructFilter(&model.ProductImage{ProductCode: code}).Delete(ctx, &model.ProductImage{}); err != nil {
			return fmt.Errorf("删除产品图片关系失败: %w", err)
		}

		// 第三步：删除产品
		if err := tx.StructFilter(&model.Product{ProductCode: code}).Delete(ctx, &model.Product{}); err != nil {
			return fmt.Errorf("删除产品失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("删除产品失败: %w", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	//第二步:删除图片
	for _, img := range images {
		filePath := filepath.Join(wd, img.ImageURL)

		fmt.Printf("删除图片路径: %s\n", filePath)

		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
}

// 尚未检查，可能有问题
func (ps *productService) FindDishesByCode(
	ctx context.Context,
	productCode string,
	cursor uint64,
	limit int,
) (*dto.CursorResp[dto.DishCardResp, uint64], error) {
	var product model.Product
	key := fmt.Sprintf("product:%s", productCode)
	err := ps.db.Query(ctx, key, &product, func(ctx context.Context, db *gormx.DB, val *model.Product) error {
		return db.StructFilter(&model.Product{ProductCode: productCode}).First(ctx, val)
	})
	if err != nil {
		return nil, fmt.Errorf("查询产品失败: %w", err)
	}
	ingredientCode := product.IngredientCode

	/*
		dishCodes := make([]string, 0)
		err = ps.db.GetXDB().Find(
			ctx, &dishIngredients,
			&model.DishIngredient{IngredientCode: ingredientCode},
		)
		if err != nil {
			return nil, fmt.Errorf("查询菜品食材关系失败: %w", err)
		}

		dishCodeMap := make(map[string]bool)
		for _, di := range dishIngredients {
			if !dishCodeMap[di.DishCode] {
				dishCodes = append(dishCodes, di.DishCode)
				dishCodeMap[di.DishCode] = true
			}
		}

		var dishes []model.Dish
		err = ps.db.GetXDB().
			Where("id > ?", cursor).
			Where("dish_code IN (?)", dishCodes).
			Order("id ASC").
			Limit(limit+1).
			Find(ctx, &dishes)
		if err != nil {
			return nil, fmt.Errorf("查询菜品失败: %w", err)
		}
	*/

	var dishes []model.Dish

	err = ps.db.GetXDB().Table("dishes as d").
		Joins("JOIN dish_ingredients as di on d.dish_code = di.dish_code").
		Where("di.ingredient_code = ?", ingredientCode).
		Where("d.id > ?", cursor).
		Order("d.id ASC").
		Limit(limit+1).
		Find(ctx, &dishes)
	if err != nil {
		return nil, fmt.Errorf("查询菜品失败: %w", err)
	}

	hasMore := len(dishes) > limit
	if hasMore {
		dishes = dishes[:limit]
	}
	newCursor := cursor
	if len(dishes) > 0 {
		newCursor = dishes[len(dishes)-1].ID
	}

	innerDishCodes := make([]string, 0, len(dishes))
	for _, dish := range dishes {
		innerDishCodes = append(innerDishCodes, dish.DishCode)
	}

	//获取第一个菜品的图片作为封面
	var dishImages []model.DishImage

	/*
		if len(innerDishCodes) > 0 {
			err := ps.db.GetXDB().
				StructFilter(&model.DishImage{DishCode: innerDishCodes[0]}).
				Order("sort_order ASC").
				Find(ctx, &dishImages)
			if err != nil {
				return nil, fmt.Errorf("查询菜品图片失败: %w", err)
			}
		}
	*/

	err = ps.db.GetXDB().Build(func(tx *gorm.DB) *gorm.DB {
		subQuery := tx.Model(&model.DishImage{}).
			Select("*, ROW_NUMBER() OVER (PARTITION BY dish_code ORDER BY sort_order) as row_number").
			Where("dish_code IN ?", innerDishCodes)
		return tx.Table("(?) as sub", subQuery).
			Where("row_number = ?", 1)
	}).Find(ctx, &dishImages)
	if err != nil {
		return nil, fmt.Errorf("查询菜品图片失败: %w", err)
	}

	dishImageMap := make(map[string]*model.DishImage)
	for i := range dishImages {
		img := &dishImages[i]
		if _, exists := dishImageMap[img.DishCode]; !exists {
			dishImageMap[img.DishCode] = img
		}
	}

	resultDishes := make([]dto.DishCardResp, 0, len(dishes))
	for _, dish := range dishes {
		resultDish := dto.DishCardResp{
			ID:       dish.ID,
			DishCode: dish.DishCode,
			Name:     dish.Name,
		}
		if img, exists := dishImageMap[dish.DishCode]; exists {
			resultDish.Image = dto.ImageResp{
				ID:        img.ID,
				SortOrder: img.SortOrder,
				ImageURL:  img.ImageURL,
			}
		}
		resultDishes = append(resultDishes, resultDish)
	}

	return &dto.CursorResp[dto.DishCardResp, uint64]{
		Items:   resultDishes,
		Cursor:  newCursor,
		HasMore: hasMore,
	}, nil
}

func (ps *productService) Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error {
	// 解析 Excel 文件
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("打开 Excel 文件失败: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("获取工作表行失败: %w", err)
	}

	var products []model.Product
	var imageURLs []model.ProductImage

	for rowNum := 2; rowNum <= len(rows); rowNum++ {

		fmt.Printf("\n处理第 %d 行,行长度 %d...\n", rowNum, len(rows[rowNum-1]))

		if rowNum%batchSize == 0 && len(products) > 0 {

			fmt.Printf("已处理 %d 行，准备提交事务...\n", rowNum)

			if err := ps.db.GetXDB().
				Clauses(
					clause.OnConflict{
						Columns: []clause.Column{
							{Name: "product_code"},
						},
						UpdateAll: true,
					},
				).
				CreateInBatches(ctx, &products, batchSize); err != nil {
				return fmt.Errorf("创建产品失败: %w", err)
			}

			products = products[:0]
		}

		for i, val := range rows[rowNum-1] {
			switch i {
			case 0:
				products = append(products, model.Product{
					ProductCode: val,
				})
			case 1:
				products[len(products)-1].IngredientCode = val
			case 2:
				products[len(products)-1].Name = val
			case 3:
				products[len(products)-1].Amount, err = strconv.ParseFloat(val, 64)
				if err != nil {
					return fmt.Errorf("解析数量失败: %w", err)
				}
			case 4:
				products[len(products)-1].Unit = model.UnitType(val)
			case 5:
				products[len(products)-1].Description = val
			case 6:
				products[len(products)-1].Price, err = strconv.ParseFloat(val, 64)
				if err != nil {
					return fmt.Errorf("解析金额失败: %w", err)
				}
			case 7:
				products[len(products)-1].AllergenType = model.AllergenType(val)
			}
		}
	}

	pictureCells, err := f.GetPictureCells(sheetName)
	if err != nil {
		return fmt.Errorf("获取图片单元格失败: %w", err)
	}

	for i, cell := range pictureCells {
		if i%batchSize == 0 && len(imageURLs) > 0 {
			fmt.Printf("已处理 %d 图片，准备提交事务...\n", i)
			if err := ps.db.GetXDB().
				Clauses(
					clause.OnConflict{
						Columns: []clause.Column{
							{Name: "product_code"},
							{Name: "sort_order"},
						},
						UpdateAll: true,
					},
				).
				CreateInBatches(ctx, &imageURLs, batchSize); err != nil {
				return fmt.Errorf("创建产品图片关系失败: %w", err)
			}
			imageURLs = imageURLs[:0]
		}

		picture, err := f.GetPictures(sheetName, cell)
		if err != nil {
			continue
		}

		fmt.Printf("图片 %s 处理中...\n", cell)
		col, row, err := excelize.CellNameToCoordinates(cell)
		if err != nil {
			continue
		}

		cellProductCode, err := excelize.CoordinatesToCellName(1, row)
		if err != nil {
			continue
		}

		productCode, err := f.GetCellValue(sheetName, cellProductCode)
		if err != nil {
			continue
		}

		if len(picture) > 0 {
			sortOrder := col - len(rows[row-1]) - 1

			url, err := ps.imgUtil.ProcessExcelPicture(
				picture[0],
				ps.uploadDir,
				productCode,
				sortOrder,
			)
			if err != nil {
				continue
			}
			imageURLs = append(imageURLs, model.ProductImage{
				ProductCode: productCode,
				// 排序,用于显示顺序
				SortOrder: sortOrder,
				ImageURL:  url,
			})
		}
	}

	if len(products) > 0 || len(imageURLs) > 0 {
		if err := ps.db.Transaction(ctx, func(tx *gormx.Executor) error {

			// 第一步：创建产品
			if len(products) > 0 {
				if err := tx.
					Clauses(
						clause.OnConflict{
							Columns: []clause.Column{
								{Name: "product_code"},
							},
							UpdateAll: true,
						},
					).
					CreateInBatches(ctx, &products, batchSize); err != nil {
					return fmt.Errorf("创建产品失败: %w", err)
				}
			}
			// 第二步：创建产品图片关系
			if len(imageURLs) > 0 {

				if err := tx.
					Clauses(
						clause.OnConflict{
							Columns: []clause.Column{
								{Name: "product_code"},
								{Name: "sort_order"},
							},
							UpdateAll: true,
						},
					).
					CreateInBatches(ctx, &imageURLs, batchSize); err != nil {
					return fmt.Errorf("创建产品图片关系失败: %w", err)
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("提交事务失败: %w", err)
		}
	}
	return nil
}

func (ps *productService) Export(gctx *gin.Context, batchSize int) error {

	gctx.Writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	gctx.Writer.Header().Set("Content-Disposition", "attachment; filename=products.xlsx")

	f := excelize.NewFile()
	// 创建一个工作表（默认已有 Sheet1，这里直接使用）

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return fmt.Errorf("创建工作表失败: %w", err)
	}

	defer f.Close()

	// 可选：设置表头
	headers := []any{"商品编码", "食材编码", "商品名称", "商品量值", "商品单位", "商品描述", "商品价格", "过敏原类型"}
	if err := sw.SetRow("A1", headers); err != nil {
		return fmt.Errorf("设置表头失败: %w", err)
	}

	currentRow := 2
	batch := 0

	ctx := gctx.Request.Context()

	offset := 0
	for {
		var products []model.Product
		err := ps.db.GetXDB().
			Offset(offset).
			Limit(batchSize).
			Find(ctx, &products)
		if err != nil {
			return fmt.Errorf("查询产品失败: %w", err)
		}

		if len(products) == 0 {
			break
		}

		batch++
		for _, product := range products {
			// 构造一行数据（必须与表头列数一致）
			row := []any{
				product.ProductCode,
				product.IngredientCode,
				product.Name,
				product.Amount,
				product.Unit,
				product.Description,
				product.Price,
				product.AllergenType,
			}
			// 计算单元格坐标，例如 A2, B2, ... 然后下一行 A3...
			cell, err := excelize.CoordinatesToCellName(1, currentRow) // 1 表示第 A 列
			if err != nil {
				return fmt.Errorf("计算单元格坐标失败: %w", err)
			}
			if err := sw.SetRow(cell, row); err != nil {
				return fmt.Errorf("写入行 %d 失败: %v", currentRow, err)
			}
			currentRow++
		}
		// 完成流式写入
		if err := sw.Flush(); err != nil {
			return fmt.Errorf("写入行 %d 失败: %w", currentRow, err)
		}
		// 可选：打印进度
		log.Printf("已处理批次 %d,写入 %d 行", batch, len(products))

		offset += batchSize
		if len(products) < batchSize {
			break
		}
	}

	// 7. 直接写入 ResponseWriter (不保存到本地磁盘)
	// 注意：此时 excelize 会在内存中构建 ZIP 结构，然后一次性或分块写入 w
	_, err = f.WriteTo(gctx.Writer)
	if err != nil {
		return fmt.Errorf("写入响应失败: %w", err)
	}
	return nil
}
