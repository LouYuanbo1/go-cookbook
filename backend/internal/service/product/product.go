package productService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"
	"go-cookbook/internal/utils"
	"go-cookbook/internal/utils/imgutil"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"github.com/LouYuanbo1/go-webservice/gormc"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ProductService interface {
	//创建和更新产品需要原子性索引使用事务和更复杂的结构体，
	Create(ctx context.Context, req *dto.CreateProductRequest) error
	//查询对一致性要求不高,可以使用两次分别查询产品和对应菜品,减少复杂性
	//使用Get获取基础信息,之后使用Find获取对应菜品,考虑到食材可能有极多对应菜品，e.g"盐",所以这里使用游标
	GetByCode(ctx context.Context, code string) (*dto.ViewProductResponse, error)
	FindDishesByProductCodeAndCursor(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewDishCardListWithCursor, error)
	//更新产品需要原子性索引使用事务和更复杂的结构体，
	Update(ctx context.Context, req *dto.UpdateProductRequest) error
	Delete(ctx context.Context, code string) error
	Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error
	Export(gctx *gin.Context, batchSize int) error
}

type productService struct {
	db      *gormc.CacheDB
	imgUtil imgutil.ImgUtil
}

func NewProductService(db *gormc.CacheDB, imgUtil imgutil.ImgUtil) ProductService {
	return &productService{db: db, imgUtil: imgUtil}
}

func (ps *productService) Create(ctx context.Context, req *dto.CreateProductRequest) error {
	// 处理图片
	imageURLs := make([]*model.ProductImage, 0, len(req.Images))
	for i, fileHeader := range req.Images {
		url, err := utils.ProcessImageFileHeader(
			ps.imgUtil,
			fileHeader,
			[]string{"uploads", "products"},
			req.ProductCode,
			i,
		)
		if err != nil {
			return fmt.Errorf("处理图片失败: %w", err)
		}
		imageURLs = append(imageURLs, &model.ProductImage{
			ProductCode: req.ProductCode,
			// 排序,用于显示顺序
			SortOrder: i,
			ImageURL:  url,
		})
	}

	err := ps.db.Transaction(ctx, func(ctx context.Context, db *gormx.DB) error {
		// 第一步：创建产品
		if err := db.Create(ctx, &model.Product{
			ProductCode:    req.ProductCode,
			IngredientCode: req.IngredientCode,
			Name:           req.Name,
			Amount:         req.Amount,
			Unit:           req.Unit,
			Description:    req.Description,
			Price:          req.Price,
			AllergenType:   req.AllergenType,
		},
			gormx.OnConstraintColumns("product_code"),
			gormx.UpdateAll(),
		); err != nil {
			return fmt.Errorf("创建产品失败: %w", err)
		}
		// 第二步：创建产品图片关系
		if err := db.CreateInBatches(ctx, imageURLs, 10,
			gormx.OnConstraintColumns("product_code", "sort_order"),
			gormx.UpdateAll(),
		); err != nil {
			return fmt.Errorf("创建产品图片关系失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("创建菜品和食材关系失败: %w", err)
	}
	return nil
}

func (ps *productService) GetByCode(ctx context.Context, code string) (*dto.ViewProductResponse, error) {
	var product model.Product
	key := fmt.Sprintf("product:%s", code)
	err := ps.db.Query(ctx, key, &product, func(ctx context.Context, db *gormx.DB, val *model.Product) error {
		return db.GetByStructFilter(ctx, val, &model.Product{ProductCode: code})
	})
	if err != nil {
		return nil, fmt.Errorf("查询产品基本信息失败: %w", err)
	}

	productResp := dto.ViewProductResponse{
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
	err = ps.db.GetXDB().FindByStructFilter(
		ctx, &images,
		&model.ProductImage{ProductCode: code},
		gormx.WithAsc("sort_order"),
	)
	if err != nil {
		return nil, fmt.Errorf("查询产品图片关系失败: %w", err)
	}

	productResp.Images = make([]dto.ImageResponse, 0, len(images))
	for _, img := range images {
		productResp.Images = append(productResp.Images, dto.ImageResponse{
			ID: img.ID,
			// 排序,用于显示顺序
			SortOrder: img.SortOrder,
			ImageURL:  img.ImageURL,
		})
	}

	return &productResp, nil
}

func (ps *productService) FindDishesByProductCodeAndCursor(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewDishCardListWithCursor, error) {
	var product model.Product
	err := ps.db.GetXDB().GetDBWithContext(ctx).First(&product, "product_code = ?", code).Error
	if err != nil {
		return nil, fmt.Errorf("查询产品失败: %w", err)
	}

	var dishIngredients []model.DishIngredient
	err = ps.db.GetXDB().GetDBWithContext(ctx).
		Where("ingredient_code = ?", product.IngredientCode).
		Find(&dishIngredients).Error
	if err != nil {
		return nil, fmt.Errorf("查询菜品食材关系失败: %w", err)
	}

	dishCodes := make([]string, 0, len(dishIngredients))
	dishCodeMap := make(map[string]bool)
	for _, di := range dishIngredients {
		if !dishCodeMap[di.DishCode] {
			dishCodes = append(dishCodes, di.DishCode)
			dishCodeMap[di.DishCode] = true
		}
	}

	var dishes []model.Dish
	err = ps.db.GetXDB().GetDBWithContext(ctx).
		Where("id > ?", cursor).
		Where("dish_code IN (?)", dishCodes).
		Order("id ASC").
		Limit(limit + 1).
		Find(&dishes).Error
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

	var dishImages []model.DishImage
	if len(innerDishCodes) > 0 {
		err := ps.db.GetXDB().FindByStructFilter(ctx, &dishImages, &model.DishImage{}, gormx.WithAsc("sort_order"))
		if err != nil {
			return nil, fmt.Errorf("查询菜品图片失败: %w", err)
		}
	}

	dishImageMap := make(map[string]*model.DishImage)
	for i := range dishImages {
		img := &dishImages[i]
		if _, exists := dishImageMap[img.DishCode]; !exists {
			dishImageMap[img.DishCode] = img
		}
	}

	resultDishes := make([]*dto.ViewDishCard, 0, len(dishes))
	for _, dish := range dishes {
		resultDish := &dto.ViewDishCard{
			ID:       dish.ID,
			DishCode: dish.DishCode,
			Name:     dish.Name,
		}
		if img, exists := dishImageMap[dish.DishCode]; exists {
			resultDish.Image = dto.ImageResponse{
				ID:        img.ID,
				SortOrder: img.SortOrder,
				ImageURL:  img.ImageURL,
			}
		}
		resultDishes = append(resultDishes, resultDish)
	}

	return &dto.ViewDishCardListWithCursor{
		Dishes:  resultDishes,
		Cursor:  newCursor,
		HasMore: hasMore,
	}, nil
}

func (ps *productService) Update(ctx context.Context, req *dto.UpdateProductRequest) error {

	fmt.Printf("UpdateProductRequest: %v\n", req)

	tempIDToFileHeader := make(map[string]*multipart.FileHeader) // tempID -> fileHeader
	for _, img := range req.NewImages {
		tempIDToFileHeader[img.TempID] = img.File
	}

	deletedImages := make([]uint64, 0)
	deletedImageURLs := make([]string, 0)
	upsertImages := make([]*model.ProductImage, 0)

	// 删除标记为deleted的图片
	for _, img := range req.Images {
		switch img.Type {
		case "deleted":
			deletedImages = append(deletedImages, img.ID)
			var img model.ProductImage
			err := ps.db.GetXDB().GetByID(ctx, &img, img.ID)
			if err != nil {
				return fmt.Errorf("查询删除图片失败: %w", err)
			}
			deletedImageURLs = append(deletedImageURLs, img.ImageURL)
		case "existing":
			upsertImages = append(upsertImages, &model.ProductImage{
				ID:          img.ID,
				ProductCode: req.ProductCode,
				SortOrder:   img.SortOrder,
			})
		case "new":
			fmt.Println("New Image:", img.TempID)
			url, err := utils.ProcessImageFileHeader(
				ps.imgUtil,
				tempIDToFileHeader[img.TempID],
				[]string{"uploads", "products"},
				req.ProductCode,
				img.SortOrder,
			)
			if err != nil {
				return fmt.Errorf("处理新图片失败: %w", err)
			}
			fmt.Printf("TempID: %s, URL: %s", img.TempID, url)
			upsertImages = append(upsertImages, &model.ProductImage{
				ProductCode: req.ProductCode,
				ImageURL:    url,
				SortOrder:   img.SortOrder,
			})
		default:
			return fmt.Errorf("未知图片操作类型: %s", img.Type)
		}
	}

	err := ps.db.Transaction(ctx, func(ctx context.Context, tx *gormx.DB) error {

		//第一步:更新产品基本信息
		err := tx.GetDBWithContext(ctx).Model(&model.Product{}).
			Where("product_code = ?", req.ProductCode).
			Updates(&model.Product{
				IngredientCode: req.IngredientCode,
				Name:           req.Name,
				Amount:         req.Amount,
				Unit:           req.Unit,
				Description:    req.Description,
				Price:          req.Price,
				AllergenType:   req.AllergenType,
			}).Error
		if err != nil {
			return fmt.Errorf("更新产品基本信息失败: %w", err)
		}

		fmt.Printf("DeletedImages: %v\n", deletedImages)

		// 第二步：删除图片
		if len(deletedImages) > 0 {
			if err := tx.DeleteByIDs[model.ProductImage](ctx, deletedImages...); err != nil {
				return fmt.Errorf("删除产品图片关系失败: %w", err)
			}
		}

		fmt.Printf("UpsertImages: %v\n", upsertImages)

		// 第三步：更新或插入图片关系
		if len(upsertImages) > 0 {
			if err := tx.CreateInBatches(
				ctx,
				upsertImages,
				10,
				gormx.OnConstraintColumns("id"),
				gormx.UpdateColumns("sort_order"),
			); err != nil {
				return fmt.Errorf("创建新图片失败: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("更新菜品失败: %w", err)
	}

	// 第五步:删除图片文件
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	for _, url := range deletedImageURLs {
		filePath := filepath.Join(wd, url)
		fmt.Printf("删除图片路径: %s\n", filePath)
		if err := ps.imgUtil.Delete(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
}

func (ps *productService) Delete(ctx context.Context, code string) error {
	var images []model.ProductImage
	var err error
	err = ps.db.Transaction(ctx, func(ctx context.Context, tx *gormx.DB) error {
		// 第一步：查询所有关联的图片
		err = tx.FindByStructFilter(ctx, &images, &model.ProductImage{ProductCode: code})
		if err != nil {
			return fmt.Errorf("查询产品图片关系失败: %w", err)
		}

		// 第二步：删除产品图片关系
		if err := tx.DeleteByStructFilter(ctx, &model.ProductImage{ProductCode: code}); err != nil {
			return fmt.Errorf("删除产品图片关系失败: %w", err)
		}

		// 第三步：删除产品
		if err := tx.DeleteByStructFilter(ctx, &model.Product{ProductCode: code}); err != nil {
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

		if err := ps.imgUtil.Delete(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
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

	var products []*model.Product
	var imageURLs []*model.ProductImage

	for rowNum := 2; rowNum <= len(rows); rowNum++ {

		fmt.Printf("\n处理第 %d 行,行长度 %d...\n", rowNum, len(rows[rowNum-1]))

		if rowNum%batchSize == 0 && len(products) > 0 {

			fmt.Printf("已处理 %d 行，准备提交事务...\n", rowNum)

			if err := ps.db.GetXDB().CreateInBatches(ctx, products, batchSize,
				gormx.OnConstraintColumns("product_code"),
				gormx.UpdateAll(),
			); err != nil {
				return fmt.Errorf("创建产品失败: %w", err)
			}

			products = products[:0]
		}

		for i, val := range rows[rowNum-1] {
			switch i {
			case 0:
				products = append(products, &model.Product{
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
			if err := ps.db.GetXDB().CreateInBatches(ctx, imageURLs, batchSize,
				gormx.OnConstraintColumns("product_code", "sort_order"),
				gormx.UpdateAll(),
			); err != nil {
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

			url, err := utils.ProcessExcelPicture(
				ps.imgUtil,
				picture[0],
				[]string{"uploads", "products"},
				productCode,
				sortOrder,
			)
			if err != nil {
				continue
			}
			imageURLs = append(imageURLs, &model.ProductImage{
				ProductCode: productCode,
				// 排序,用于显示顺序
				SortOrder: sortOrder,
				ImageURL:  url,
			})
		}
	}

	if len(products) > 0 || len(imageURLs) > 0 {
		if err := ps.db.Transaction(ctx, func(ctx context.Context, db *gormx.DB) error {

			// 第一步：创建产品
			if len(products) > 0 {
				if err := db.CreateInBatches(ctx, products, batchSize,
					gormx.OnConstraintColumns("product_code"),
					gormx.UpdateAll(),
				); err != nil {
					return fmt.Errorf("创建产品失败: %w", err)
				}
			}
			// 第二步：创建产品图片关系
			if len(imageURLs) > 0 {

				if err := db.CreateInBatches(ctx, imageURLs, batchSize,
					gormx.OnConstraintColumns("product_code", "sort_order"),
					gormx.UpdateAll(),
				); err != nil {
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
		err := ps.db.GetXDB().GetDBWithContext(ctx).
			Offset(offset).
			Limit(batchSize).
			Find(&products).Error
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
