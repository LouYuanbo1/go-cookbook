package dish

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

	"github.com/LouYuanbo1/go-webservice/gormc"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DishService interface {
	//创建和更新菜品需要原子性索引使用事务和更复杂的结构体，
	Create(ctx context.Context, req *dto.CreateDishReq) error
	//查询对一致性要求不高,可以使用两次分别查询菜品和对应食材,减少复杂性
	FirstByCode(ctx context.Context, code string) (*dto.DishResp, error)
	FindByCursor(
		ctx context.Context,
		cursor uint64,
		limit int,
	) (*dto.CursorResp[dto.DishCardResp, uint64], error)
	//创建和更新菜品需要原子性索引使用事务和更复杂的结构体，
	Update(ctx context.Context, req *dto.UpdateDishReq) error
	Delete(ctx context.Context, code string) error

	//查询对一致性要求不高,可以使用两次分别查询菜品和对应食材,减少复杂性
	FindDishIngredientsByCode(
		ctx context.Context,
		dishCode string,
		cursor uint64,
		limit int,
	) (*dto.CursorResp[dto.DishIngredientCardResp, uint64], error)
	Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error
	Export(gctx *gin.Context, batchSize int) error
}

type dishService struct {
	db        *gormc.CacheDB
	imgUtil   img.ImgUtil
	tempfs    tempfs.TempFs
	tempDir   string
	uploadDir string
}

func NewDishService(
	db *gormc.CacheDB,
	imgUtil img.ImgUtil,
	tempfs tempfs.TempFs,
	tempDir string,
	uploadDir string,
) DishService {
	return &dishService{
		db:        db,
		imgUtil:   imgUtil,
		tempfs:    tempfs,
		tempDir:   tempDir,
		uploadDir: uploadDir,
	}
}

func (ds *dishService) registerImages(
	imageFileHeaders []*multipart.FileHeader,
	dishCode string,
) (images []model.DishImage, mapIdImageName map[string]string, err error) {
	images = make([]model.DishImage, 0, len(imageFileHeaders))
	mapIdImageName = make(map[string]string)

	for i, fileHeader := range imageFileHeaders {

		ext := filepath.Ext(fileHeader.Filename)
		imageName := fmt.Sprintf("%s_%d%s", dishCode, i, ext)

		err := ds.imgUtil.ProcessFileHeader(
			fileHeader,
			ds.tempDir,
			imageName,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("处理图片失败: %w", err)
		}

		/*
			// 从 URL 中提取实际文件名（带扩展名）
			// URL 格式: /tempDir/DISH001_0.jpg
			actualFilename := path.Base(url)

			id, err := ds.tempfs.RegisterTempFile(filepath.Join(ds.tempDir, actualFilename))
			if err != nil {
				return nil, nil, fmt.Errorf("注册临时文件失败: %w", err)
			}
			mapIdImageName[id] = actualFilename
		*/

		id, err := ds.tempfs.RegisterTempFile(filepath.Join(ds.tempDir, imageName))
		if err != nil {
			return nil, nil, fmt.Errorf("注册临时文件失败: %w", err)
		}
		mapIdImageName[id] = imageName

		images = append(images, model.DishImage{
			DishCode: dishCode,
			// 排序,用于显示顺序
			SortOrder: i,
			ImageURL:  fmt.Sprintf("/%s", filepath.ToSlash(filepath.Join(ds.uploadDir, imageName))),
		})
	}
	return images, mapIdImageName, nil
}

func (ds *dishService) Create(ctx context.Context, req *dto.CreateDishReq) error {
	// 处理图片
	images, mapIdImageName, err := ds.registerImages(req.Images, req.DishCode)
	if err != nil {
		return fmt.Errorf("注册图片失败: %w", err)
	}

	err = ds.db.Transaction(ctx, func(tx *gormx.Executor) error {
		// 第一步：创建菜品
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "dish_code"},
			},
			UpdateAll: true,
		}).Create(ctx, &model.Dish{
			DishCode:    req.DishCode,
			Name:        req.Name,
			Description: req.Description,
			Recipe:      req.Recipe,
		},
		); err != nil {
			return fmt.Errorf("创建菜品失败: %w", err)
		}
		// 第二步：创建菜品图片
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "dish_code"},
				{Name: "sort_order"},
			},
			UpdateAll: true,
		}).CreateInBatches(ctx, &images, 10); err != nil {
			return fmt.Errorf("创建菜品图片关系失败: %w", err)
		}

		// 第三步：创建菜品食材关系
		ingredients := make([]*model.DishIngredient, 0, len(req.Ingredients))
		for _, ingredient := range req.Ingredients {
			ingredients = append(ingredients, &model.DishIngredient{
				DishCode:       req.DishCode,
				IngredientCode: ingredient.IngredientCode,
				Quantity:       ingredient.Quantity,
				Note:           ingredient.Note,
			})
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "dish_code"},
				{Name: "ingredient_code"},
			},
			UpdateAll: true,
		}).CreateInBatches(ctx, &ingredients, 10); err != nil {
			return fmt.Errorf("创建菜品食材关系失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("创建菜品和图片关系失败: %w", err)
	}

	// 第三步：将临时文件移动到正式目录
	for id, imageName := range mapIdImageName {
		if err := ds.tempfs.PromoteTempFile(id, ds.uploadDir, imageName); err != nil {
			return fmt.Errorf("移动临时文件失败: %w", err)
		}
	}

	return err
}

func (ds *dishService) FirstByCode(ctx context.Context, code string) (*dto.DishResp, error) {
	var dish model.Dish
	key := fmt.Sprintf("dish:%s", code)
	err := ds.db.Query(ctx, key, &dish, func(ctx context.Context, db *gormx.DB, val *model.Dish) error {
		return db.StructFilter(&model.Dish{DishCode: code}).First(ctx, val)
	})
	if err != nil {
		return nil, fmt.Errorf("查询菜品基本信息失败: %w", err)
	}

	dishResp := dto.DishResp{
		DishCode:    dish.DishCode,
		Name:        dish.Name,
		Description: dish.Description,
		Recipe:      dish.Recipe,
	}

	var images []model.DishImage
	err = ds.db.GetXDB().
		StructFilter(&model.DishImage{DishCode: code}).
		OrderByColumn(clause.OrderByColumn{Column: clause.Column{Name: "sort_order"}, Desc: false}).
		Find(ctx,
			&images)
	if err != nil {
		return nil, fmt.Errorf("查询菜品图片关系失败: %w", err)
	}

	dishResp.Images = make([]dto.ImageResp, 0, len(images))
	for _, img := range images {
		dishResp.Images = append(dishResp.Images, dto.ImageResp{
			ID: img.ID,
			// 排序,用于显示顺序
			SortOrder: img.SortOrder,
			ImageURL:  img.ImageURL,
		})
	}

	return &dishResp, nil
}

func (ds *dishService) FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.CursorResp[dto.DishCardResp, uint64], error) {
	var dishes []model.Dish
	err := ds.db.GetXDB().
		Where("id > ?", cursor).
		Order("id ASC").
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

	dishCodes := make([]string, 0, len(dishes))
	for _, dish := range dishes {
		dishCodes = append(dishCodes, dish.DishCode)
	}

	var dishImages []model.DishImage

	/*
		if len(dishCodes) > 0 {
			err := ds.db.GetXDB().
				StructFilter(&model.DishImage{DishCode: dishCodes[0]}).
				OrderByColumn(clause.OrderByColumn{Column: clause.Column{Name: "sort_order"}, Desc: false}).
				Find(ctx, &dishImages)
			if err != nil {
				return nil, fmt.Errorf("查询菜品图片失败: %w", err)
			}
		}
	*/

	err = ds.db.GetXDB().Build(func(tx *gorm.DB) *gorm.DB {
		subQuery := tx.Model(&model.DishImage{}).
			Select("*, ROW_NUMBER() OVER (PARTITION BY dish_code ORDER BY sort_order) as row_number").
			Where("dish_code IN ?", dishCodes)
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

// 逻辑复杂,可能有问题,需要测试
func (ds *dishService) Update(ctx context.Context, req *dto.UpdateDishReq) error {
	// 处理新图片注册
	upsertImages, mapIdImageName, err := ds.registerImages(req.NewImages, req.DishCode)
	if err != nil {
		return fmt.Errorf("注册图片失败: %w", err)
	}

	// 处理更新图片
	if len(req.UpdatedImages) > 0 {
		for _, img := range req.UpdatedImages {
			upsertImages = append(upsertImages, model.DishImage{
				ID:       img.ID,
				DishCode: req.DishCode,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
		}
	}

	// 处理删除图片
	deletedImageURLs := make([]string, 0)
	if len(req.DeletedImageIDs) > 0 {
		err := ds.db.GetXDB().Model(&model.DishImage{}).Where("id IN ?", req.DeletedImageIDs).Pluck(ctx, "image_url", &deletedImageURLs)
		if err != nil {
			return fmt.Errorf("查询删除图片失败: %w", err)
		}
	}

	upsertIngredients := make([]model.DishIngredient, 0)

	for _, ingredient := range req.Ingredients {
		switch ingredient.Type {
		case "deleted":
			if err := ds.db.GetXDB().Delete(
				ctx,
				&model.DishIngredient{
					DishCode:       req.DishCode,
					IngredientCode: ingredient.IngredientCode,
				}); err != nil {
				return fmt.Errorf("删除删除食材失败: %w", err)
			}
		case "existing", "new":
			upsertIngredients = append(upsertIngredients, model.DishIngredient{
				DishCode:       req.DishCode,
				IngredientCode: ingredient.IngredientCode,
				Quantity:       ingredient.Quantity,
				Note:           ingredient.Note,
			})
		default:
			return fmt.Errorf("未知的食材操作类型: %s", ingredient.Type)
		}
	}

	err = ds.db.Transaction(ctx, func(tx *gormx.Executor) error {

		// 第一步:更新菜品基本信息
		err := tx.StructFilter(&model.Dish{
			DishCode: req.DishCode,
		}).Updates(ctx, &model.Dish{
			Name:        req.Name,
			Description: req.Description,
			Recipe:      req.Recipe,
		})
		if err != nil {
			return fmt.Errorf("更新菜品失败: %w", err)
		}

		// 第二步:删除图片
		if len(req.DeletedImageIDs) > 0 {
			if err := tx.Delete(ctx, &model.DishImage{}, req.DeletedImageIDs); err != nil {
				return fmt.Errorf("删除图片失败: %w", err)
			}
		}

		// 第三步:创建或更新图片
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "id"},
			},
			UpdateAll: true,
		}).CreateInBatches(
			ctx,
			&upsertImages,
			10,
		); err != nil {
			return fmt.Errorf("创建新图片失败: %w", err)
		}

		// 第四步:创建或更新食材
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "dish_code"},
				{Name: "ingredient_code"},
			},
			UpdateAll: true,
		}).CreateInBatches(
			ctx,
			&upsertIngredients,
			10,
		); err != nil {
			return fmt.Errorf("创建新食材失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("更新菜品失败: %w", err)
	}

	// 持久化新图片
	for id, imageName := range mapIdImageName {
		if err := ds.tempfs.PromoteTempFile(id, ds.uploadDir, imageName); err != nil {
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

func (ds *dishService) Delete(ctx context.Context, code string) error {
	var images []model.DishImage
	var err error
	err = ds.db.Transaction(ctx, func(tx *gormx.Executor) error {

		// 第一步：查询所有关联的图片
		err = tx.StructFilter(&model.DishImage{DishCode: code}).Find(ctx, &images)
		if err != nil {
			return fmt.Errorf("查询菜品图片关系失败: %w", err)
		}

		// 第二步：删除图片关系
		if err := tx.StructFilter(&model.DishImage{DishCode: code}).Delete(ctx, &model.DishImage{}); err != nil {
			return fmt.Errorf("删除菜品图片失败: %w", err)
		}

		// 第三步：删除菜品食材关系
		if err := tx.StructFilter(&model.DishIngredient{DishCode: code}).Delete(ctx, &model.DishIngredient{}); err != nil {
			return fmt.Errorf("删除菜品食材关系失败: %w", err)
		}

		// 第四步：删除菜品
		if err := tx.StructFilter(&model.Dish{DishCode: code}).Delete(ctx, &model.Dish{}); err != nil {
			return fmt.Errorf("删除菜品失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("删除菜品失败: %w", err)
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

/*
[
  {
    IngredientCode: "ING001",
    Name: "面粉",
    Quantity: "200g",
    Note: "高筋面粉",
    ImageID: 1,
    ImageOrder: 1,
    ImageURL: "flour1.jpg"
  },
  {
    IngredientCode: "ING001",
    Name: "面粉",
    Quantity: "200g",
    Note: "高筋面粉",
    ImageID: 2,
    ImageOrder: 2,
    ImageURL: "flour2.jpg"
  },
  {
    IngredientCode: "ING002",
    Name: "鸡蛋",
    Quantity: "2个",
    Note: "新鲜鸡蛋",
    ImageID: 3,
    ImageOrder: 1,
    ImageURL: "egg1.jpg"
  }
]
*/

func (ds *dishService) FindDishIngredientsByCode(
	ctx context.Context,
	dishCode string,
	cursor uint64,
	limit int,
) (*dto.CursorResp[dto.DishIngredientCardResp, uint64], error) {
	// 第一步：查询该菜品关联的所有食材编码
	var dishIngredientCodes []string
	err := ds.db.GetXDB().Model(&model.DishIngredient{}).
		Where("dish_code = ?", dishCode).
		Pluck(ctx, "ingredient_code", &dishIngredientCodes)
	if err != nil {
		return nil, fmt.Errorf("查询菜品食材关系失败: %w", err)
	}

	if len(dishIngredientCodes) == 0 {
		return &dto.CursorResp[dto.DishIngredientCardResp, uint64]{
			Items:   []dto.DishIngredientCardResp{},
			Cursor:  0,
			HasMore: false,
		}, nil
	}

	// 第二步：按 ingredients.id 分页查询食材
	var ingredients []model.Ingredient
	err = ds.db.GetXDB().Model(&model.Ingredient{}).
		Where("ingredient_code IN ?", dishIngredientCodes).
		Where("id > ?", cursor).
		Order("id ASC").
		Limit(limit+1).
		Find(ctx, &ingredients)
	if err != nil {
		return nil, fmt.Errorf("查询食材信息失败: %w", err)
	}

	hasMore := len(ingredients) > limit
	if hasMore {
		ingredients = ingredients[:limit]
	}
	newCursor := cursor
	if len(ingredients) > 0 {
		newCursor = ingredients[len(ingredients)-1].ID
	}

	// 第三步：构建 ingredient_code -> quantity/note 的映射
	ingredientNoteMap := make(map[string]*model.DishIngredient)
	var dishIngredients []model.DishIngredient
	err = ds.db.GetXDB().Model(&model.DishIngredient{}).
		Where("dish_code = ?", dishCode).
		Where("ingredient_code IN ?", dishIngredientCodes).
		Find(ctx, &dishIngredients)
	if err != nil {
		return nil, fmt.Errorf("查询菜品食材详情失败: %w", err)
	}
	for i := range dishIngredients {
		di := &dishIngredients[i]
		ingredientNoteMap[di.IngredientCode] = di
	}

	// 第四步：查询食材图片（每个食材取排序第一的图片）
	ingredientCodes := make([]string, 0, len(ingredients))
	for _, ing := range ingredients {
		ingredientCodes = append(ingredientCodes, ing.IngredientCode)
	}

	var ingredientImages []model.IngredientImage
	if len(ingredientCodes) > 0 {
		err := ds.db.GetXDB().Build(func(tx *gorm.DB) *gorm.DB {
			subQuery := tx.Model(&model.IngredientImage{}).
				Select("*, ROW_NUMBER() OVER (PARTITION BY ingredient_code ORDER BY sort_order) as row_number").
				Where("ingredient_code IN ?", ingredientCodes)
			return tx.Table("(?) as sub", subQuery).
				Where("row_number = ?", 1)
		}).Find(ctx, &ingredientImages)
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

	// 第五步：组装结果
	resultIngredients := make([]dto.DishIngredientCardResp, 0, len(ingredients))
	for _, ing := range ingredients {
		di := ingredientNoteMap[ing.IngredientCode]
		if di == nil {
			continue
		}
		resultIngredient := dto.DishIngredientCardResp{
			ID:             ing.ID,
			IngredientCode: ing.IngredientCode,
			Name:           ing.Name,
			Quantity:       di.Quantity,
			Note:           di.Note,
		}
		if img, exists := ingredientImageMap[ing.IngredientCode]; exists {
			resultIngredient.Image = dto.ImageResp{
				ID:        img.ID,
				SortOrder: img.SortOrder,
				ImageURL:  img.ImageURL,
			}
		}
		resultIngredients = append(resultIngredients, resultIngredient)
	}

	return &dto.CursorResp[dto.DishIngredientCardResp, uint64]{
		Items:   resultIngredients,
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

func (ds *dishService) Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error {
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

	var dishes []model.Dish
	var dishIngredients []model.DishIngredient
	var imageURLs []model.DishImage

	for rowNum := 2; rowNum <= len(rows); rowNum++ {

		fmt.Printf("\n处理第 %d 行,行长度 %d...\n", rowNum, len(rows[rowNum-1]))

		if rowNum%batchSize == 0 && (len(dishes) > 0 || len(dishIngredients) > 0) {

			fmt.Printf("已处理 %d 行，准备提交事务...\n", rowNum)

			if err := ds.db.Transaction(ctx, func(tx *gormx.Executor) error {
				// 第一步：创建菜品
				if len(dishes) > 0 {
					if err := tx.Clauses(
						clause.OnConflict{
							Columns:   []clause.Column{{Name: "dish_code"}},
							UpdateAll: true,
						},
					).
						CreateInBatches(ctx, &dishes, batchSize); err != nil {
						return fmt.Errorf("创建产品失败: %w", err)
					}
				}
				// 第二步：创建菜品食材关系
				if len(dishIngredients) > 0 {
					if err := tx.Clauses(
						clause.OnConflict{
							Columns:   []clause.Column{{Name: "dish_code"}, {Name: "ingredient_code"}},
							UpdateAll: true,
						},
					).
						CreateInBatches(ctx, &dishIngredients, batchSize); err != nil {
						return fmt.Errorf("创建菜品食材关系失败: %w", err)
					}
				}

				return nil
			}); err != nil {
				return fmt.Errorf("提交事务失败: %w", err)
			}

			dishes = dishes[:0]
			dishIngredients = dishIngredients[:0]
		}

		for i, val := range rows[rowNum-1] {
			if i < 4 {
				switch i {
				case 0:
					dishes = append(dishes, model.Dish{
						DishCode: val,
					})
				case 1:
					dishes[len(dishes)-1].Name = val
				case 2:
					dishes[len(dishes)-1].Description = val
				case 3:
					dishes[len(dishes)-1].Recipe = val
				}
			} else {
				switch i%3 - 1 {
				case 0:
					dishIngredients = append(dishIngredients, model.DishIngredient{
						DishCode:       dishes[len(dishes)-1].DishCode,
						IngredientCode: val,
					})
				case 1:
					dishIngredients[len(dishIngredients)-1].Quantity = val
				case 2:
					dishIngredients[len(dishIngredients)-1].Note = val
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

				if err := ds.db.GetXDB().Clauses(
					clause.OnConflict{
						Columns:   []clause.Column{{Name: "dish_code"}, {Name: "sort_order"}},
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

			cellDishCode, err := excelize.CoordinatesToCellName(1, row)
			if err != nil {
				continue
			}

			dishCode, err := f.GetCellValue(sheetName, cellDishCode)
			if err != nil {
				continue
			}

			if len(picture) > 0 {
				sortOrder := col - len(rows[row-1]) - 1

				url, err := ds.imgUtil.ProcessExcelPicture(
					picture[0],
					ds.uploadDir,
					dishCode,
					sortOrder,
				)
				if err != nil {
					continue
				}

				imageURLs = append(imageURLs, model.DishImage{
					DishCode: dishCode,
					// 排序,用于显示顺序
					SortOrder: sortOrder,
					ImageURL:  url,
				})
			}
		}

		if len(dishes) > 0 || len(dishIngredients) > 0 || len(imageURLs) > 0 {

			if err := ds.db.Transaction(ctx, func(tx *gormx.Executor) error {
				// 第一步：创建菜品
				if len(dishes) > 0 {
					if err := tx.Clauses(
						clause.OnConflict{
							Columns:   []clause.Column{{Name: "dish_code"}},
							UpdateAll: true,
						},
					).
						CreateInBatches(ctx, &dishes, batchSize); err != nil {
						return fmt.Errorf("创建产品失败: %w", err)
					}
				}
				// 第二步：创建菜品食材关系
				if len(dishIngredients) > 0 {
					if err := tx.Clauses(
						clause.OnConflict{
							Columns:   []clause.Column{{Name: "dish_code"}, {Name: "ingredient_code"}},
							UpdateAll: true,
						},
					).
						CreateInBatches(ctx, &dishIngredients, batchSize); err != nil {
						return fmt.Errorf("创建产品食材关系失败: %w", err)
					}
				}
				// 第三步：创建菜品图片关系
				if len(imageURLs) > 0 {

					if err := tx.Clauses(
						clause.OnConflict{
							Columns:   []clause.Column{{Name: "dish_code"}, {Name: "sort_order"}},
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
	}
	return nil
}

func (ds *dishService) Export(gctx *gin.Context, batchSize int) error {

	gctx.Writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	gctx.Writer.Header().Set("Content-Disposition", "attachment; filename=dishes.xlsx")

	f := excelize.NewFile()
	// 创建一个工作表（默认已有 Sheet1，这里直接使用）

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return fmt.Errorf("创建工作表失败: %w", err)
	}

	defer f.Close()

	// 可选：设置表头
	headers := []any{"菜品编码", "菜品名称", "菜品描述", "菜品做法"}
	if err := sw.SetRow("A1", headers); err != nil {
		return fmt.Errorf("设置表头失败: %w", err)
	}

	currentRow := 2

	ctx := gctx.Request.Context()

	err = ds.db.
		GetXDB().
		FindInBatches(ctx, batchSize, func(tx *gormx.Executor, batch int, models *[]model.Dish) error {
			for _, dish := range *models {
				// 构造一行数据（必须与表头列数一致）
				row := []any{dish.DishCode, dish.Name, dish.Description, dish.Recipe}
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
			log.Printf("已处理批次 %d,写入 %d 行", batch, len(*models))
			return nil
		})

	// 7. 直接写入 ResponseWriter (不保存到本地磁盘)
	// 注意：此时 excelize 会在内存中构建 ZIP 结构，然后一次性或分块写入 w
	_, err = f.WriteTo(gctx.Writer)
	if err != nil {
		return fmt.Errorf("写入响应失败: %w", err)
	}
	return nil
}
