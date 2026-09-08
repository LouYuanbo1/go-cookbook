package ingredient

import (
	"context"
	"fmt"
	"go-cookbook/internal/common/dto"
	"go-cookbook/internal/common/utils/img"
	"go-cookbook/internal/model"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	"go-cookbook/internal/common/utils/tempfs"

	"github.com/LouYuanbo1/go-webservice/gormc"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IngredientService interface {
	Create(ctx context.Context, req *dto.CreateIngredientReq) error
	FirstByCode(ctx context.Context, code string) (*dto.IngredientResp, error)
	FindByCursor(
		ctx context.Context,
		req *dto.CursorReq[uint64],
	) (*dto.CursorResp[dto.IngredientCardResp, uint64], error)
	Update(ctx context.Context, req *dto.UpdateIngredientReq) error
	Delete(ctx context.Context, code string) error

	FindProductsByCode(
		ctx context.Context,
		ingredientCode string,
		cursor uint64,
		limit int,
	) (*dto.CursorResp[dto.ProductCardResp, uint64], error)
	Import(ctx context.Context, file *multipart.FileHeader, batchSize int) error
	Export(gctx *gin.Context, batchSize int) error
}

type ingredientService struct {
	db        *gormc.CacheDB
	imgUtil   img.ImgUtil
	tempfs    tempfs.TempFs
	tempDir   string
	uploadDir string
}

func NewIngredientService(
	db *gormc.CacheDB,
	imgUtil img.ImgUtil,
	tempfs tempfs.TempFs,
	tempDir string,
	uploadDir string,
) IngredientService {
	return &ingredientService{
		db:        db,
		imgUtil:   imgUtil,
		tempfs:    tempfs,
		tempDir:   tempDir,
		uploadDir: uploadDir,
	}
}

func (is *ingredientService) registerImages(
	imageFileHeaders []*multipart.FileHeader,
	ingredientCode string,
) (images []model.IngredientImage, mapIdImageName map[string]string, err error) {
	images = make([]model.IngredientImage, 0, len(imageFileHeaders))
	mapIdImageName = make(map[string]string)

	for i, fileHeader := range imageFileHeaders {

		ext := filepath.Ext(fileHeader.Filename)
		imageName := fmt.Sprintf("%s_%d%s", ingredientCode, i, ext)

		err := is.imgUtil.ProcessFileHeader(
			fileHeader,
			is.tempDir,
			imageName,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("处理图片失败: %w", err)
		}

		/*
			// 从 URL 中提取实际文件名（带扩展名）
			// URL 格式: /tempDir/ING001_0.jpg
			actualFilename := path.Base(url)

			id, err := is.tempfs.RegisterTempFile(filepath.Join(is.tempDir, actualFilename))
			if err != nil {
				return nil, nil, fmt.Errorf("注册临时文件失败: %w", err)
			}
			mapIdImageName[id] = actualFilename
		*/

		id, err := is.tempfs.RegisterTempFile(filepath.Join(is.tempDir, imageName))
		if err != nil {
			return nil, nil, fmt.Errorf("注册临时文件失败: %w", err)
		}
		mapIdImageName[id] = imageName

		images = append(images, model.IngredientImage{
			IngredientCode: ingredientCode,
			// 排序,用于显示顺序
			SortOrder: i,
			ImageURL:  fmt.Sprintf("/%s", filepath.ToSlash(filepath.Join(is.uploadDir, imageName))),
		})
	}
	return images, mapIdImageName, nil
}

func (is *ingredientService) Create(ctx context.Context, req *dto.CreateIngredientReq) error {
	// 处理图片
	images, mapIdImageName, err := is.registerImages(req.Images, req.IngredientCode)
	if err != nil {
		return fmt.Errorf("注册图片失败: %w", err)
	}

	err = is.db.Transaction(ctx, func(tx *gormx.Executor) error {

		// 第一步：创建产品
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "ingredient_code"},
			},
			UpdateAll: true,
		}).Create(ctx, &model.Ingredient{
			IngredientCode: req.IngredientCode,
			Name:           req.Name,
			Description:    req.Description,
		},
		); err != nil {
			return fmt.Errorf("创建产品失败: %w", err)
		}
		// 第二步：创建产品图片关系
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "ingredient_code"},
				{Name: "sort_order"},
			},
			UpdateAll: true,
		}).CreateInBatches(ctx, &images, 10); err != nil {
			return fmt.Errorf("创建产品图片关系失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("创建食材和图片关系失败: %w", err)
	}

	// 第三步：将临时文件移动到正式目录
	for id, imageName := range mapIdImageName {
		if err := is.tempfs.PromoteTempFile(id, is.uploadDir, imageName); err != nil {
			return fmt.Errorf("移动临时文件失败: %w", err)
		}
	}

	return nil
}

func (is *ingredientService) FirstByCode(ctx context.Context, code string) (*dto.IngredientResp, error) {
	var ingredient model.Ingredient
	key := fmt.Sprintf("ingredient:%s", code)
	err := is.db.Query(ctx, key, &ingredient, func(ctx context.Context, db *gormx.DB, val *model.Ingredient) error {
		return db.StructFilter(&model.Ingredient{IngredientCode: code}).First(ctx, val)
	})
	if err != nil {
		return nil, fmt.Errorf("查询食材基本信息失败: %w", err)
	}

	ingredientResp := dto.IngredientResp{
		IngredientCode: ingredient.IngredientCode,
		Name:           ingredient.Name,
		Description:    ingredient.Description,
	}

	var images []model.IngredientImage
	err = is.db.GetXDB().StructFilter(&model.IngredientImage{IngredientCode: code}).
		Order("sort_order ASC").
		Find(ctx, &images)
	if err != nil {
		return nil, fmt.Errorf("查询食材图片关系失败: %w", err)
	}

	ingredientResp.Images = make([]dto.ImageResp, 0, len(images))
	for _, img := range images {
		ingredientResp.Images = append(ingredientResp.Images, dto.ImageResp{
			ID: img.ID,
			// 排序,用于显示顺序
			SortOrder: img.SortOrder,
			ImageURL:  img.ImageURL,
		})
	}

	return &ingredientResp, nil
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

/*
查询所有食材卡片列表,支持分页
*/
func (is *ingredientService) FindByCursor(
	ctx context.Context,
	req *dto.CursorReq[uint64],
) (*dto.CursorResp[dto.IngredientCardResp, uint64], error) {
	var ingredients []model.Ingredient
	result := is.db.GetXDB().Model(&model.Ingredient{}).Where("id > ?", req.Cursor).
		Order("id ASC").
		Limit(req.Limit+1).
		Find(ctx, &ingredients)

	if result != nil {
		return nil, fmt.Errorf("查询食材失败: %w", result)
	}

	hasMore := len(ingredients) > req.Limit
	if hasMore {
		ingredients = ingredients[:req.Limit]
	}
	newCursor := req.Cursor
	if len(ingredients) > 0 {
		newCursor = ingredients[len(ingredients)-1].ID
	}

	ingredientCodes := make([]string, 0, len(ingredients))
	for _, ing := range ingredients {
		ingredientCodes = append(ingredientCodes, ing.IngredientCode)
	}

	// 查询食材图片，按排序顺序查询,每个食材最多查询1条图片，貌似有些问题，需要修改
	var ingredientImages []model.IngredientImage
	/*
		if len(ingredientCodes) > 0 {
			err := is.db.GetXDB().Model(&model.IngredientImage{}).Where("ingredient_code IN ?", ingredientCodes).
				Order("sort_order ASC").
				Find(ctx, &ingredientImages)
			if err != nil {
				return nil, fmt.Errorf("查询食材图片失败: %w", err)
			}
		}
	*/

	err := is.db.GetXDB().Build(func(tx *gorm.DB) *gorm.DB {
		subQuery := tx.Model(&model.IngredientImage{}).
			Select("*, ROW_NUMBER() OVER (PARTITION BY ingredient_code ORDER BY sort_order) as row_number").
			Where("ingredient_code IN ?", ingredientCodes)
		return tx.Table("(?) as sub", subQuery).
			Where("row_number = ?", 1)
	}).Find(ctx, &ingredientImages)
	if err != nil {
		return nil, fmt.Errorf("查询食材图片失败: %w", err)
	}

	ingredientImageMap := make(map[string]*model.IngredientImage)
	for i := range ingredientImages {
		img := &ingredientImages[i]
		if _, exists := ingredientImageMap[img.IngredientCode]; !exists {
			ingredientImageMap[img.IngredientCode] = img
		}
	}

	ingredientCards := make([]dto.IngredientCardResp, 0, len(ingredients))
	for _, ing := range ingredients {
		ingredient := dto.IngredientCardResp{
			IngredientCode: ing.IngredientCode,
			Name:           ing.Name,
		}
		if img, exists := ingredientImageMap[ing.IngredientCode]; exists {
			ingredient.Image = dto.ImageResp{
				ID:        img.ID,
				SortOrder: img.SortOrder,
				ImageURL:  img.ImageURL,
			}
		}
		ingredientCards = append(ingredientCards, ingredient)
	}

	return &dto.CursorResp[dto.IngredientCardResp, uint64]{
		Items:   ingredientCards,
		Cursor:  newCursor,
		HasMore: hasMore,
	}, nil
}

/*
func (is *ingredientService) Update(ctx context.Context, req *UpdateIngredientReq) error {
	tempIDToFileHeader := make(map[string]*multipart.FileHeader) // tempID -> fileHeader
	for _, img := range req.NewImages {
		tempIDToFileHeader[img.TempID] = img.File
	}

	deletedImages := make([]uint64, 0)
	deletedImageURLs := make([]string, 0)
	upsertImages := make([]model.IngredientImage, 0)
	// 删除标记为deleted的图片
	for _, img := range req.Images {
		switch img.Type {
		case "deleted":
			deletedImages = append(deletedImages, img.ID)
		case "existing":
			upsertImages = append(upsertImages, model.IngredientImage{
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
				[]string{"uploads", "ingredients", req.IngredientCode},
				req.IngredientCode,
				true,
			)
			if err != nil {
				return fmt.Errorf("处理新图片失败: %w", err)
			}
			fmt.Printf("TempID: %s, URL: %s", img.TempID, url)
			upsertImages = append(upsertImages, model.IngredientImage{
				IngredientCode: req.IngredientCode,
				ImageURL:       url,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
		default:
			return fmt.Errorf("未知图片操作类型: %s", img.Type)
		}
	}

	if len(deletedImages) > 0 {
		err := is.db.GetXDB().Model(&model.IngredientImage{}).Where("id IN ?", deletedImages).Pluck(ctx, "image_url", &deletedImageURLs)
		if err != nil {
			return fmt.Errorf("查询删除图片失败: %w", err)
		}
	}

	err := is.db.Transaction(ctx, func(tx *gormx.Executor) error {
		//第一步:更新食材基本信息
		err := tx.Model(&model.Ingredient{}).
			StructFilter(&model.Ingredient{
				IngredientCode: req.IngredientCode,
			}).
			Updates(ctx, &model.Ingredient{
				Name:        req.Name,
				Description: req.Description,
			})
		if err != nil {
			return fmt.Errorf("更新食材基本信息失败: %w", err)
		}

		fmt.Printf("deletedImages: %v\n", deletedImages)

		if len(deletedImages) > 0 {
			// 第二步：删除图片
			if err := tx.Delete(ctx, &model.IngredientImage{}, deletedImages); err != nil {
				return fmt.Errorf("删除食材图片关系失败: %w", err)
			}
		}

		fmt.Printf("upsertImages: %v\n", upsertImages)

		if len(upsertImages) > 0 {
			// 第三步：更新或插入图片关系
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"sort_order"}),
			}).CreateInBatches(
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
*/

func (is *ingredientService) Update(ctx context.Context, req *dto.UpdateIngredientReq) error {
	// 处理新图片注册
	upsertImages, mapIdImageName, err := is.registerImages(req.NewImages, req.IngredientCode)
	if err != nil {
		return fmt.Errorf("注册图片失败: %w", err)
	}

	// 处理更新图片
	if len(req.UpdatedImages) > 0 {
		for _, img := range req.UpdatedImages {
			upsertImages = append(upsertImages, model.IngredientImage{
				ID:             img.ID,
				IngredientCode: req.IngredientCode,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
		}
	}

	// 处理删除图片
	deletedImageURLs := make([]string, 0)
	if len(req.DeletedImageIDs) > 0 {
		err := is.db.GetXDB().Model(&model.IngredientImage{}).Where("id IN ?", req.DeletedImageIDs).Pluck(ctx, "image_url", &deletedImageURLs)
		if err != nil {
			return fmt.Errorf("查询删除图片失败: %w", err)
		}
	}

	err = is.db.Transaction(ctx, func(tx *gormx.Executor) error {
		//第一步:更新食材基本信息
		err := tx.Model(&model.Ingredient{}).
			StructFilter(&model.Ingredient{
				IngredientCode: req.IngredientCode,
			}).
			Updates(ctx, &model.Ingredient{
				Name:        req.Name,
				Description: req.Description,
			})
		if err != nil {
			return fmt.Errorf("更新食材基本信息失败: %w", err)
		}

		if len(req.DeletedImageIDs) > 0 {
			// 第二步：删除图片
			if err := tx.Delete(ctx, &model.IngredientImage{}, req.DeletedImageIDs); err != nil {
				return fmt.Errorf("删除食材图片关系失败: %w", err)
			}
		}

		if len(upsertImages) > 0 {
			// 第三步：更新或插入图片关系
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"sort_order"}),
			}).CreateInBatches(
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
		return fmt.Errorf("更新食材失败: %w", err)
	}

	// 持久化新图片
	for id, imageName := range mapIdImageName {
		if err := is.tempfs.PromoteTempFile(id, is.uploadDir, imageName); err != nil {
			return fmt.Errorf("移动临时文件失败: %w", err)
		}
	}

	// 第五步:删除图片文件,这里的逻辑可能修改
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

func (is *ingredientService) Delete(ctx context.Context, code string) error {
	var images []model.IngredientImage
	err := is.db.Transaction(ctx, func(tx *gormx.Executor) error {
		// 第一步：查询所有关联的图片
		err := tx.Find(ctx, &images, &model.IngredientImage{
			IngredientCode: code})
		if err != nil {
			return fmt.Errorf("查询食材图片关系失败: %w", err)
		}

		// 第二步：删除食材图片关系
		if err := tx.StructFilter(&model.IngredientImage{IngredientCode: code}).Delete(ctx, &model.IngredientImage{}); err != nil {
			return fmt.Errorf("删除食材图片关系失败: %w", err)
		}

		// 第三步：删除食材
		if err := tx.StructFilter(&model.Ingredient{IngredientCode: code}).Delete(ctx, &model.Ingredient{}); err != nil {
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
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
}

func (is *ingredientService) FindProductsByCode(
	ctx context.Context,
	ingredientCode string,
	cursor uint64,
	limit int,
) (*dto.CursorResp[dto.ProductCardResp, uint64], error) {
	var products []model.Product
	result := is.db.GetXDB().Model(&model.Product{}).Where("ingredient_code = ?", ingredientCode).
		Where("id > ?", cursor).
		Order("id ASC").
		Limit(limit+1).
		Find(ctx, &products)

	if result != nil {
		return nil, fmt.Errorf("查询产品失败: %w", result)
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
	/*
		if len(productCodes) > 0 {
			err := is.db.GetXDB().Model(&model.ProductImage{}).Where("product_code IN ?", productCodes).
				Order("sort_order ASC").
				Find(ctx, &productImages)
			if err != nil {
				return nil, fmt.Errorf("查询产品图片失败: %w", err)
			}
		}
	*/

	err := is.db.GetXDB().Build(func(tx *gorm.DB) *gorm.DB {
		subQuery := tx.Model(&model.ProductImage{}).
			Select("*, ROW_NUMBER() OVER (PARTITION BY product_code ORDER BY sort_order) as row_number").
			Where("product_code IN ?", productCodes)
		return tx.Table("(?) as sub", subQuery).
			Where("row_number = ?", 1)
	}).Find(ctx, &productImages)
	if err != nil {
		return nil, fmt.Errorf("查询产品图片失败: %w", err)
	}

	productImageMap := make(map[string]*model.ProductImage)
	for i := range productImages {
		img := &productImages[i]
		productImageMap[img.ProductCode] = img
	}

	resultProducts := make([]dto.ProductCardResp, 0, len(products))
	for _, p := range products {
		resultProduct := dto.ProductCardResp{
			ID:           p.ID,
			ProductCode:  p.ProductCode,
			Name:         p.Name,
			Amount:       p.Amount,
			Unit:         p.Unit,
			Price:        p.Price,
			AllergenType: p.AllergenType,
		}
		if img, exists := productImageMap[p.ProductCode]; exists {
			resultProduct.Image = dto.ImageResp{
				ID:        img.ID,
				SortOrder: img.SortOrder,
				ImageURL:  img.ImageURL,
			}
		}
		resultProducts = append(resultProducts, resultProduct)
	}

	return &dto.CursorResp[dto.ProductCardResp, uint64]{
		Items:   resultProducts,
		Cursor:  newCursor,
		HasMore: hasMore,
	}, nil
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

	var ingredients []model.Ingredient
	var imageURLs []model.IngredientImage

	for rowNum := 2; rowNum <= len(rows); rowNum++ {

		fmt.Printf("\n处理第 %d 行,行长度 %d...\n", rowNum, len(rows[rowNum-1]))

		if rowNum%batchSize == 0 && len(ingredients) > 0 {

			fmt.Printf("已处理 %d 行，准备提交事务...\n", rowNum)

			if err := is.db.GetXDB().Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "ingredient_code"},
				},
				UpdateAll: true,
			}).CreateInBatches(ctx, &ingredients, batchSize); err != nil {
				return fmt.Errorf("创建产品失败: %w", err)
			}

			ingredients = ingredients[:0]
		}

		for i, val := range rows[rowNum-1] {
			switch i {
			case 0:
				ingredients = append(ingredients, model.Ingredient{
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

			if err := is.db.GetXDB().Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "ingredient_code"},
					{Name: "sort_order"},
				},
				UpdateAll: true,
			}).CreateInBatches(ctx, &imageURLs, batchSize); err != nil {
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

			url, err := is.imgUtil.ProcessExcelPicture(
				picture[0],
				is.uploadDir,
				ingredientCode,
				sortOrder,
			)
			if err != nil {
				continue
			}

			imageURLs = append(imageURLs, model.IngredientImage{
				IngredientCode: ingredientCode,
				// 排序,用于显示顺序
				SortOrder: sortOrder,
				ImageURL:  url,
			})
		}
	}

	if len(ingredients) > 0 || len(imageURLs) > 0 {

		if err := is.db.Transaction(ctx, func(tx *gormx.Executor) error {
			// 第一步：创建产品
			if len(ingredients) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns: []clause.Column{
						{Name: "ingredient_code"},
					},
					UpdateAll: true,
				}).CreateInBatches(ctx, &ingredients, batchSize); err != nil {
					return fmt.Errorf("创建产品失败: %w", err)
				}
			}
			// 第二步：创建产品图片关系
			if len(imageURLs) > 0 {

				if err := tx.Clauses(clause.OnConflict{
					Columns: []clause.Column{
						{Name: "ingredient_code"},
						{Name: "sort_order"},
					},
					UpdateAll: true,
				}).CreateInBatches(ctx, &imageURLs, batchSize); err != nil {
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
		err := is.db.GetXDB().Offset(offset).
			Limit(batchSize).
			Find(ctx, &ingredients)
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
