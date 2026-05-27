package ingredientService

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

	"github.com/LouYuanbo1/go-webservice/gormc"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type IngredientService interface {
	Create(ctx context.Context, req *dto.CreateIngredientRequest) error
	GetByCode(ctx context.Context, code string) (*dto.ViewIngredientResponse, error)
	FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.ViewIngredientCardListWithCursor, error)
	FindProductsByIngredientCodeAndCursor(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewProductCardListWithCursor, error)
	Update(ctx context.Context, req *dto.UpdateIngredientRequest) error
	Delete(ctx context.Context, code string) error
	Import(ctx context.Context, file *multipart.FileHeader, batchSize int) error
	Export(gctx *gin.Context, batchSize int) error
}

type ingredientService struct {
	db      *gormc.CacheDB
	imgUtil imgutil.ImgUtil
}

func NewIngredientService(db *gormc.CacheDB, imgUtil imgutil.ImgUtil) IngredientService {
	return &ingredientService{db: db, imgUtil: imgUtil}
}

func (is *ingredientService) Create(ctx context.Context, req *dto.CreateIngredientRequest) error {
	// 处理图片
	imageURLs := make([]*model.IngredientImage, 0, len(req.Images))
	for i, fileHeader := range req.Images {
		url, err := utils.ProcessImageFileHeader(
			is.imgUtil,
			fileHeader,
			[]string{"uploads", "ingredients"},
			req.IngredientCode,
			i,
		)
		if err != nil {
			return fmt.Errorf("处理图片失败: %w", err)
		}
		imageURLs = append(imageURLs, &model.IngredientImage{
			IngredientCode: req.IngredientCode,
			// 排序,用于显示顺序
			SortOrder: i,
			ImageURL:  url,
		})
	}

	err := is.db.Transaction(ctx, func(ctx context.Context, db *gormx.DB) error {

		// 第一步：创建产品
		if err := db.Create(ctx, &model.Ingredient{
			IngredientCode: req.IngredientCode,
			Name:           req.Name,
			Description:    req.Description,
		},
			gormx.OnConstraintColumns("ingredient_code"),
			gormx.UpdateAll(),
		); err != nil {
			return fmt.Errorf("创建产品失败: %w", err)
		}
		// 第二步：创建产品图片关系
		if err := db.CreateInBatches(ctx, imageURLs, 10,
			gormx.OnConstraintColumns("ingredient_code", "sort_order"),
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

func (is *ingredientService) GetByCode(ctx context.Context, code string) (*dto.ViewIngredientResponse, error) {
	var ingredient model.Ingredient
	key := fmt.Sprintf("ingredient:%s", code)
	err := is.db.Query(ctx, key, &ingredient, func(ctx context.Context, db *gormx.DB, val *model.Ingredient) error {
		return db.GetByStructFilter(ctx, val, &model.Ingredient{IngredientCode: code})
	})
	if err != nil {
		return nil, fmt.Errorf("查询食材基本信息失败: %w", err)
	}

	ingredientResp := dto.ViewIngredientResponse{
		IngredientCode: ingredient.IngredientCode,
		Name:           ingredient.Name,
		Description:    ingredient.Description,
	}

	var images []model.IngredientImage
	err = is.db.GetXDB().FindByStructFilter(
		ctx, &images,
		&model.IngredientImage{IngredientCode: code},
		gormx.WithAsc("sort_order"),
	)
	if err != nil {
		return nil, fmt.Errorf("查询食材图片关系失败: %w", err)
	}

	ingredientResp.Images = make([]dto.ImageResponse, 0, len(images))
	for _, img := range images {
		ingredientResp.Images = append(ingredientResp.Images, dto.ImageResponse{
			ID: img.ID,
			// 排序,用于显示顺序
			SortOrder: img.SortOrder,
			ImageURL:  img.ImageURL,
		})
	}

	return &ingredientResp, nil
}

func (is *ingredientService) FindProductsByIngredientCodeAndCursor(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewProductCardListWithCursor, error) {
	var products []model.Product
	result := is.db.GetXDB().GetDBWithContext(ctx).
		Where("ingredient_code = ?", code).
		Where("id > ?", cursor).
		Order("id ASC").
		Limit(limit + 1).
		Find(&products)

	if result.Error != nil {
		return nil, fmt.Errorf("查询产品失败: %w", result.Error)
	}

	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit]
	}
	newCursor := cursor
	if len(products) > 0 {
		newCursor = products[len(products)-1].ID
	}

	productCodes := make([]string, 0, len(products))
	for _, p := range products {
		productCodes = append(productCodes, p.ProductCode)
	}

	var productImages []model.ProductImage
	if len(productCodes) > 0 {
		err := is.db.GetXDB().FindByStructFilter(ctx, &productImages, &model.ProductImage{}, gormx.WithAsc("sort_order"))
		if err != nil {
			return nil, fmt.Errorf("查询产品图片失败: %w", err)
		}
	}

	productImageMap := make(map[string]*model.ProductImage)
	for i := range productImages {
		img := &productImages[i]
		if _, exists := productImageMap[img.ProductCode]; !exists {
			productImageMap[img.ProductCode] = img
		}
	}

	resultProducts := make([]*dto.ViewProductCard, 0, len(products))
	for _, p := range products {
		resultProduct := &dto.ViewProductCard{
			ID:           p.ID,
			ProductCode:  p.ProductCode,
			Name:         p.Name,
			Amount:       p.Amount,
			Unit:         p.Unit,
			Price:        p.Price,
			AllergenType: p.AllergenType,
		}
		if img, exists := productImageMap[p.ProductCode]; exists {
			resultProduct.Image = dto.ImageResponse{
				ID:        img.ID,
				SortOrder: img.SortOrder,
				ImageURL:  img.ImageURL,
			}
		}
		resultProducts = append(resultProducts, resultProduct)
	}

	return &dto.ViewProductCardListWithCursor{
		Products: resultProducts,
		Cursor:   newCursor,
		HasMore:  hasMore,
	}, nil
}

/*
慢SQL,平均查询时间300ms,可能的性能瓶颈,需要测试
1.使用LATERAL:
SELECT

	i.id,
	i.ingredient_code,
	i.name,
	i.description,
	ii.id AS img_id,
	ii.sort_order AS img_order,
	ii.image_url AS img_url

FROM ingredients i
LEFT JOIN LATERAL (

	SELECT id, sort_order, image_url
	FROM ingredient_images
	WHERE ingredient_code = i.ingredient_code
	ORDER BY sort_order ASC
	LIMIT 1

) ii ON true
WHERE i.id > 0
ORDER BY i.id ASC
LIMIT 16;

2.索引覆盖:
CREATE INDEX idx_images_code_sort ON ingredient_images (ingredient_code, sort_order)
INCLUDE (id, image_url);
*/
func (is *ingredientService) FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.ViewIngredientCardListWithCursor, error) {
	var ingredients []model.Ingredient
	result := is.db.GetXDB().GetDBWithContext(ctx).
		Where("id > ?", cursor).
		Order("id ASC").
		Limit(limit + 1).
		Find(&ingredients)

	if result.Error != nil {
		return nil, fmt.Errorf("查询食材失败: %w", result.Error)
	}

	hasMore := len(ingredients) > limit
	if hasMore {
		ingredients = ingredients[:limit]
	}
	newCursor := cursor
	if len(ingredients) > 0 {
		newCursor = ingredients[len(ingredients)-1].ID
	}

	ingredientCodes := make([]string, 0, len(ingredients))
	for _, ing := range ingredients {
		ingredientCodes = append(ingredientCodes, ing.IngredientCode)
	}

	var ingredientImages []model.IngredientImage
	if len(ingredientCodes) > 0 {
		err := is.db.GetXDB().FindByStructFilter(ctx, &ingredientImages, &model.IngredientImage{}, gormx.WithAsc("sort_order"))
		if err != nil {
			return nil, fmt.Errorf("查询食材图片失败: %w", err)
		}
	}

	ingredientImageMap := make(map[string]*model.IngredientImage)
	for i := range ingredientImages {
		img := &ingredientImages[i]
		if _, exists := ingredientImageMap[img.IngredientCode]; !exists {
			ingredientImageMap[img.IngredientCode] = img
		}
	}

	ingredientCards := make([]*dto.ViewIngredientCard, 0, len(ingredients))
	for _, ing := range ingredients {
		ingredient := &dto.ViewIngredientCard{
			IngredientCode: ing.IngredientCode,
			Name:           ing.Name,
			Description:    ing.Description,
		}
		if img, exists := ingredientImageMap[ing.IngredientCode]; exists {
			ingredient.Image = dto.ImageResponse{
				ID:        img.ID,
				SortOrder: img.SortOrder,
				ImageURL:  img.ImageURL,
			}
		}
		ingredientCards = append(ingredientCards, ingredient)
	}

	return &dto.ViewIngredientCardListWithCursor{
		Ingredients: ingredientCards,
		Cursor:      newCursor,
		HasMore:     hasMore,
	}, nil
}

func (is *ingredientService) Update(ctx context.Context, req *dto.UpdateIngredientRequest) error {
	tempIDToFileHeader := make(map[string]*multipart.FileHeader) // tempID -> fileHeader
	for _, img := range req.NewImages {
		tempIDToFileHeader[img.TempID] = img.File
	}

	deletedImages := make([]uint64, 0)
	deletedImageURLs := make([]string, 0)
	upsertImages := make([]*model.IngredientImage, 0)
	// 删除标记为deleted的图片
	for _, img := range req.Images {
		switch img.Type {
		case "deleted":
			deletedImages = append(deletedImages, img.ID)
			var img model.IngredientImage
			err := is.db.GetXDB().GetByID(ctx, &img, img.ID)
			if err != nil {
				return fmt.Errorf("查询删除图片失败: %w", err)
			}
			deletedImageURLs = append(deletedImageURLs, img.ImageURL)
		case "existing":
			upsertImages = append(upsertImages, &model.IngredientImage{
				ID:             img.ID,
				IngredientCode: req.IngredientCode,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
		case "new":
			fmt.Println("New Image:", img.TempID)
			url, err := utils.ProcessImageFileHeader(
				is.imgUtil,
				tempIDToFileHeader[img.TempID],
				[]string{"uploads", "ingredients"},
				req.IngredientCode,
				img.SortOrder,
			)
			if err != nil {
				return fmt.Errorf("处理新图片失败: %w", err)
			}
			fmt.Printf("TempID: %s, URL: %s", img.TempID, url)
			upsertImages = append(upsertImages, &model.IngredientImage{
				IngredientCode: req.IngredientCode,
				ImageURL:       url,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
		default:
			return fmt.Errorf("未知图片操作类型: %s", img.Type)
		}
	}

	err := is.db.Transaction(ctx, func(ctx context.Context, tx *gormx.DB) error {
		//第一步:更新食材基本信息
		err := tx.GetDBWithContext(ctx).Model(&model.Ingredient{}).
			Where("ingredient_code = ?", req.IngredientCode).
			Updates(&model.Ingredient{
				Name:        req.Name,
				Description: req.Description,
			}).Error
		if err != nil {
			return fmt.Errorf("更新食材基本信息失败: %w", err)
		}

		fmt.Printf("deletedImages: %v\n", deletedImages)

		if len(deletedImages) > 0 {
			// 第二步：删除图片
			if err := tx.DeleteByIDs[model.IngredientImage](ctx, deletedImages...); err != nil {
				return fmt.Errorf("删除食材图片关系失败: %w", err)
			}
		}

		fmt.Printf("upsertImages: %v\n", upsertImages)

		if len(upsertImages) > 0 {
			// 第三步：更新或插入图片关系
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
		return fmt.Errorf("更新食材失败: %w", err)
	}

	// 第五步:删除图片文件
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	for _, url := range deletedImageURLs {
		filePath := filepath.Join(wd, url)
		fmt.Printf("删除图片路径: %s\n", filePath)
		if err := is.imgUtil.Delete(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
}

func (is *ingredientService) Delete(ctx context.Context, code string) error {
	var images []model.IngredientImage
	var err error
	err = is.db.Transaction(ctx, func(ctx context.Context, tx *gormx.DB) error {
		// 第一步：查询所有关联的图片
		err = tx.FindByStructFilter(ctx, &images, &model.IngredientImage{IngredientCode: code})
		if err != nil {
			return fmt.Errorf("查询食材图片关系失败: %w", err)
		}

		// 第二步：删除食材图片关系
		if err := tx.DeleteByStructFilter(ctx, &model.IngredientImage{IngredientCode: code}); err != nil {
			return fmt.Errorf("删除食材图片关系失败: %w", err)
		}

		// 第三步：删除食材
		if err := tx.DeleteByStructFilter(ctx, &model.Ingredient{IngredientCode: code}); err != nil {
			return fmt.Errorf("删除食材失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("删除食材失败: %w", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	//第二步:删除图片
	for _, img := range images {
		filePath := filepath.Join(wd, img.ImageURL)

		fmt.Printf("删除图片路径: %s\n", filePath)
		if err := is.imgUtil.Delete(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
}

// 非常恐怖的面条代码,强烈与excel和数据库结构耦合并且暂时无法优化,需要重点关注
/*
为了减小文件体积，Excel默认会对插入的图片进行压缩。当你手动将图片拖拽缩小到单元格大小时，Excel可能会在保存文件时，
根据其内部的压缩设置，丢弃图片的像素数据以“适应”这个较小的显示尺寸。一旦原始像素丢失，即使后续在代码中尝试高清读取，也无法挽回。
修改Excel的全局设置：在 Excel 中，点击「文件」->「选项」->「高级」，在“图像大小和质量”部分，勾选 “不压缩文件中的图像”，
并将“默认分辨率”设置为 “高保真” 或最高ppi值。注意：这个设置需要在插入图片前就配置好，才能保证原始图片数据被完整保留。
*/

func (is *ingredientService) Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error {
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

	var ingredients []*model.Ingredient
	var imageURLs []*model.IngredientImage

	for rowNum := 2; rowNum <= len(rows); rowNum++ {

		fmt.Printf("\n处理第 %d 行,行长度 %d...\n", rowNum, len(rows[rowNum-1]))

		if rowNum%batchSize == 0 && len(ingredients) > 0 {

			fmt.Printf("已处理 %d 行，准备提交事务...\n", rowNum)

			if err := is.db.GetXDB().CreateInBatches(ctx, ingredients, batchSize,
				gormx.OnConstraintColumns("ingredient_code"),
				gormx.UpdateAll(),
			); err != nil {
				return fmt.Errorf("创建产品失败: %w", err)
			}

			ingredients = ingredients[:0]
		}

		for i, val := range rows[rowNum-1] {
			switch i {
			case 0:
				ingredients = append(ingredients, &model.Ingredient{
					IngredientCode: val,
				})
			case 1:
				ingredients[len(ingredients)-1].Name = val
			case 2:
				ingredients[len(ingredients)-1].Description = val
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

			if err := is.db.GetXDB().CreateInBatches(ctx, imageURLs, batchSize,
				gormx.OnConstraintColumns("ingredient_code", "sort_order"),
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

		cellIngredientCode, err := excelize.CoordinatesToCellName(1, row)
		if err != nil {
			continue
		}

		ingredientCode, err := f.GetCellValue(sheetName, cellIngredientCode)
		if err != nil {
			continue
		}

		if len(picture) > 0 {
			sortOrder := col - len(rows[row-1]) - 1

			url, err := utils.ProcessExcelPicture(
				is.imgUtil,
				picture[0],
				[]string{"uploads", "ingredients"},
				ingredientCode,
				sortOrder,
			)
			if err != nil {
				continue
			}

			imageURLs = append(imageURLs, &model.IngredientImage{
				IngredientCode: ingredientCode,
				// 排序,用于显示顺序
				SortOrder: sortOrder,
				ImageURL:  url,
			})
		}
	}

	if len(ingredients) > 0 || len(imageURLs) > 0 {

		if err := is.db.Transaction(ctx, func(ctx context.Context, db *gormx.DB) error {
			// 第一步：创建产品
			if len(ingredients) > 0 {
				if err := db.CreateInBatches(ctx, ingredients, batchSize,
					gormx.OnConstraintColumns("ingredient_code"),
					gormx.UpdateAll(),
				); err != nil {
					return fmt.Errorf("创建产品失败: %w", err)
				}
			}
			// 第二步：创建产品图片关系
			if len(imageURLs) > 0 {

				if err := db.CreateInBatches(ctx, imageURLs, batchSize,
					gormx.OnConstraintColumns("ingredient_code", "sort_order"),
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

func (is *ingredientService) Export(gctx *gin.Context, batchSize int) error {

	gctx.Writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	gctx.Writer.Header().Set("Content-Disposition", "attachment; filename=ingredients.xlsx")

	f := excelize.NewFile()
	// 创建一个工作表（默认已有 Sheet1，这里直接使用）

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return fmt.Errorf("创建工作表失败: %w", err)
	}

	defer f.Close()

	// 可选：设置表头
	headers := []any{"食材编码", "名称", "描述"}
	if err := sw.SetRow("A1", headers); err != nil {
		return fmt.Errorf("设置表头失败: %w", err)
	}

	currentRow := 2
	batch := 0

	ctx := gctx.Request.Context()

	offset := 0
	for {
		var ingredients []model.Ingredient
		err := is.db.GetXDB().GetDBWithContext(ctx).
			Offset(offset).
			Limit(batchSize).
			Find(&ingredients).Error
		if err != nil {
			return fmt.Errorf("查询食材失败: %w", err)
		}

		if len(ingredients) == 0 {
			break
		}

		batch++
		for _, ingredient := range ingredients {
			// 构造一行数据（必须与表头列数一致）
			row := []any{ingredient.IngredientCode, ingredient.Name, ingredient.Description}
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
		log.Printf("已处理批次 %d,写入 %d 行", batch, len(ingredients))

		offset += batchSize
		if len(ingredients) < batchSize {
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
