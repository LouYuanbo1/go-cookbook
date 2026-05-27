package dishService

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

type DishService interface {
	//创建和更新菜品需要原子性索引使用事务和更复杂的结构体，
	Create(ctx context.Context, req *dto.CreateDishRequest) error
	//查询对一致性要求不高,可以使用两次分别查询菜品和对应食材,减少复杂性
	GetByCode(ctx context.Context, code string) (*dto.ViewDishResponse, error)
	FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.ViewDishCardListWithCursor, error)
	//查询对一致性要求不高,可以使用两次分别查询菜品和对应食材,减少复杂性
	FindIngredientsByDishCode(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewDishIngredientCardListWithCursor, error)
	//创建和更新菜品需要原子性索引使用事务和更复杂的结构体，
	Update(ctx context.Context, req *dto.UpdateDishRequest) error
	Delete(ctx context.Context, code string) error
	Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error
	Export(gctx *gin.Context, batchSize int) error
}

type dishService struct {
	db      *gormc.CacheDB
	imgUtil imgutil.ImgUtil
}

func NewDishService(db *gormc.CacheDB, imgUtil imgutil.ImgUtil) DishService {
	return &dishService{db: db, imgUtil: imgUtil}
}

func (ds *dishService) Create(ctx context.Context, req *dto.CreateDishRequest) error {
	// 处理图片
	imageURLs := make([]*model.DishImage, 0, len(req.Images))
	for i, fileHeader := range req.Images {
		url, err := utils.ProcessImageFileHeader(
			ds.imgUtil,
			fileHeader,
			[]string{"uploads", "dishes"},
			req.DishCode,
			i,
		)
		if err != nil {
			return fmt.Errorf("处理图片失败: %w", err)
		}
		imageURLs = append(imageURLs, &model.DishImage{
			DishCode:  req.DishCode,
			SortOrder: i,
			ImageURL:  url,
		})
	}

	err := ds.db.Transaction(ctx, func(ctx context.Context, db *gormx.DB) error {
		// 第一步：创建菜品
		if err := db.Create(ctx, &model.Dish{
			DishCode:    req.DishCode,
			Name:        req.Name,
			Description: req.Description,
			Recipe:      req.Recipe,
		},
			gormx.OnConstraintColumns("dish_code"),
			gormx.UpdateAll(),
		); err != nil {
			return fmt.Errorf("创建菜品失败: %w", err)
		}
		// 第二步：创建菜品图片
		if err := db.CreateInBatches(ctx, imageURLs, 10,
			gormx.OnConstraintColumns("dish_code", "sort_order"),
			gormx.UpdateAll(),
		); err != nil {
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

		if err := db.CreateInBatches(ctx, ingredients, 10,
			gormx.OnConstraintColumns("dish_code", "ingredient_code"),
			gormx.UpdateAll(),
		); err != nil {
			return fmt.Errorf("创建菜品食材关系失败: %w", err)
		}

		return nil
	})
	return err
}

func (ds *dishService) GetByCode(ctx context.Context, code string) (*dto.ViewDishResponse, error) {
	var dish model.Dish
	key := fmt.Sprintf("dish:%s", code)
	err := ds.db.Query(ctx, key, &dish, func(ctx context.Context, db *gormx.DB, val *model.Dish) error {
		return db.GetByStructFilter(ctx, val, &model.Dish{DishCode: code})
	})
	if err != nil {
		return nil, fmt.Errorf("查询菜品基本信息失败: %w", err)
	}

	dishResp := dto.ViewDishResponse{
		DishCode:    dish.DishCode,
		Name:        dish.Name,
		Description: dish.Description,
		Recipe:      dish.Recipe,
	}

	var images []model.DishImage
	err = ds.db.GetXDB().
		FindByStructFilter(
			ctx,
			&images,
			&model.DishImage{DishCode: code},
			gormx.WithAsc("sort_order"))
	if err != nil {
		return nil, fmt.Errorf("查询菜品图片关系失败: %w", err)
	}

	dishResp.Images = make([]dto.ImageResponse, 0, len(images))
	for _, img := range images {
		dishResp.Images = append(dishResp.Images, dto.ImageResponse{
			ID: img.ID,
			// 排序,用于显示顺序
			SortOrder: img.SortOrder,
			ImageURL:  img.ImageURL,
		})
	}

	return &dishResp, nil
}

func (ds *dishService) FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.ViewDishCardListWithCursor, error) {
	var dishes []model.Dish
	res := ds.db.GetXDB().GetDBWithContext(ctx).
		Where("id > ?", cursor).
		Order("id ASC").
		Limit(limit + 1).
		Find(&dishes)

	if res.Error != nil {
		return nil, fmt.Errorf("查询菜品失败: %w", res.Error)
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
	if len(dishCodes) > 0 {
		err := ds.db.GetXDB().FindByStructFilter(ctx, &dishImages, &model.DishImage{DishCode: dishCodes[0]}, gormx.WithAsc("sort_order"))
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

func (ds *dishService) FindIngredientsByDishCode(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewDishIngredientCardListWithCursor, error) {
	var dishIngredients []model.DishIngredient
	err := ds.db.GetXDB().GetDBWithContext(ctx).
		Table("dish_ingredients di").
		Joins("JOIN ingredients i ON di.ingredient_code = i.ingredient_code").
		Where("di.dish_code = ?", code).
		Where("i.id > ?", cursor).
		Order("i.id ASC").
		Limit(limit + 1).
		Find(&dishIngredients).Error
	if err != nil {
		return nil, fmt.Errorf("查询菜品食材关系失败: %w", err)
	}

	hasMore := len(dishIngredients) > limit
	if hasMore {
		dishIngredients = dishIngredients[:limit]
	}
	newCursor := cursor
	if len(dishIngredients) > 0 {
		newCursor = dishIngredients[len(dishIngredients)-1].ID
	}

	ingredientCodes := make([]string, 0, len(dishIngredients))
	for _, di := range dishIngredients {
		ingredientCodes = append(ingredientCodes, di.IngredientCode)
	}

	var ingredients []model.Ingredient
	if len(ingredientCodes) > 0 {
		err := ds.db.GetXDB().GetDBWithContext(ctx).
			Where("ingredient_code IN (?)", ingredientCodes).
			Find(&ingredients).Error
		if err != nil {
			return nil, fmt.Errorf("查询食材信息失败: %w", err)
		}
	}

	ingredientMap := make(map[string]*model.Ingredient)
	for i := range ingredients {
		ing := &ingredients[i]
		ingredientMap[ing.IngredientCode] = ing
	}

	var ingredientImages []model.IngredientImage
	if len(ingredientCodes) > 0 {
		err := ds.db.GetXDB().FindByStructFilter(ctx, &ingredientImages, &model.IngredientImage{}, gormx.WithAsc("sort_order"))
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

	resultIngredients := make([]*dto.ViewDishIngredientCard, 0, len(dishIngredients))
	for _, di := range dishIngredients {
		ing := ingredientMap[di.IngredientCode]
		if ing == nil {
			continue
		}
		resultIngredient := &dto.ViewDishIngredientCard{
			ID:             ing.ID,
			IngredientCode: ing.IngredientCode,
			Name:           ing.Name,
			Quantity:       di.Quantity,
			Note:           di.Note,
		}
		if img, exists := ingredientImageMap[ing.IngredientCode]; exists {
			resultIngredient.Image = dto.ImageResponse{
				ID:        img.ID,
				SortOrder: img.SortOrder,
				ImageURL:  img.ImageURL,
			}
		}
		resultIngredients = append(resultIngredients, resultIngredient)
	}

	return &dto.ViewDishIngredientCardListWithCursor{
		DishIngredients: resultIngredients,
		Cursor:          newCursor,
		HasMore:         hasMore,
	}, nil
}

// 逻辑复杂,可能有问题,需要测试
func (ds *dishService) Update(ctx context.Context, req *dto.UpdateDishRequest) error {
	tempIDToFileHeader := make(map[string]*multipart.FileHeader) // tempID -> fileHeader
	for _, img := range req.NewImages {
		tempIDToFileHeader[img.TempID] = img.File
	}

	deletedImages := make([]uint64, 0)
	deletedImageURLs := make([]string, 0)
	upsertImages := make([]*model.DishImage, 0) // id -> true
	for _, img := range req.Images {
		switch img.Type {
		case "deleted":
			deletedImages = append(deletedImages, img.ID)
			var img model.DishImage
			err := ds.db.GetXDB().GetByID(ctx, &img, img.ID)
			if err != nil {
				return fmt.Errorf("查询删除图片失败: %w", err)
			}
			deletedImageURLs = append(deletedImageURLs, img.ImageURL)
		case "existing":
			upsertImages = append(upsertImages, &model.DishImage{
				ID:       img.ID,
				DishCode: req.DishCode,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
			ds.db.GetXDB().UpdatesByStructFilter(ctx, &model.DishImage{
				ID: img.ID,
			}, &model.DishImage{
				//使用负顺序,避免唯一约束影响调换顺序
				SortOrder: -img.SortOrder,
			})
		case "new":
			url, err := utils.ProcessImageFileHeader(
				ds.imgUtil,
				tempIDToFileHeader[img.TempID],
				[]string{"uploads", "dishes"},
				req.DishCode,
				img.SortOrder,
			)
			if err != nil {
				return fmt.Errorf("处理新图片失败: %w", err)
			}
			upsertImages = append(upsertImages, &model.DishImage{
				DishCode: req.DishCode,
				ImageURL: url,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
		default:
			return fmt.Errorf("未知的图片操作类型: %s", img.Type)
		}
	}

	upsertIngredients := make([]*model.DishIngredient, 0)

	for _, ingredient := range req.Ingredients {
		switch ingredient.Type {
		case "deleted":
			if err := ds.db.GetXDB().DeleteByStructFilter(
				ctx,
				&model.DishIngredient{
					DishCode:       req.DishCode,
					IngredientCode: ingredient.IngredientCode,
				}); err != nil {
				return fmt.Errorf("删除删除食材失败: %w", err)
			}
		case "existing":
			upsertIngredients = append(upsertIngredients, &model.DishIngredient{
				DishCode:       req.DishCode,
				IngredientCode: ingredient.IngredientCode,
				Quantity:       ingredient.Quantity,
				Note:           ingredient.Note,
			})
		case "new":
			upsertIngredients = append(upsertIngredients, &model.DishIngredient{
				DishCode:       req.DishCode,
				IngredientCode: ingredient.IngredientCode,
				Quantity:       ingredient.Quantity,
				Note:           ingredient.Note,
			})
		default:
			return fmt.Errorf("未知的食材操作类型: %s", ingredient.Type)
		}
	}

	err := ds.db.Transaction(ctx, func(ctx context.Context, db *gormx.DB) error {

		// 第一步:更新菜品基本信息
		err := db.UpdatesByStructFilter(ctx, &model.Dish{
			DishCode: req.DishCode,
		}, &model.Dish{
			Name:        req.Name,
			Description: req.Description,
			Recipe:      req.Recipe,
		})
		if err != nil {
			return fmt.Errorf("更新菜品失败: %w", err)
		}

		// 第二步:删除图片
		if len(deletedImages) > 0 {
			if err := db.DeleteByIDs[model.DishImage](ctx, deletedImages...); err != nil {
				return fmt.Errorf("删除图片失败: %w", err)
			}
		}

		// 第三步:创建或更新图片
		if err := db.CreateInBatches(
			ctx,
			upsertImages,
			10,
			gormx.OnConstraintColumns("id"),
			gormx.UpdateColumns("sort_order"),
		); err != nil {
			return fmt.Errorf("创建新图片失败: %w", err)
		}

		// 第四步:创建或更新食材
		if err := db.CreateInBatches(
			ctx,
			upsertIngredients,
			10,
			gormx.OnConstraintColumns("dish_code", "ingredient_code"),
			gormx.UpdateColumns("quantity", "note"),
		); err != nil {
			return fmt.Errorf("创建新食材失败: %w", err)
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
		if err := ds.imgUtil.Delete(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}

	return nil
}

func (ds *dishService) Delete(ctx context.Context, code string) error {
	var images []model.DishImage
	var err error
	err = ds.db.Transaction(ctx, func(ctx context.Context, db *gormx.DB) error {

		// 第一步：查询所有关联的图片
		err = db.FindByStructFilter(ctx, &images, &model.DishImage{DishCode: code})
		if err != nil {
			return fmt.Errorf("查询菜品图片关系失败: %w", err)
		}

		// 第二步：删除图片关系
		if err := db.DeleteByStructFilter(ctx, &model.DishImage{DishCode: code}); err != nil {
			return fmt.Errorf("删除菜品图片失败: %w", err)
		}

		// 第三步：删除菜品食材关系
		if err := db.DeleteByStructFilter(ctx, &model.DishIngredient{DishCode: code}); err != nil {
			return fmt.Errorf("删除菜品食材关系失败: %w", err)
		}

		// 第四步：删除菜品
		if err := db.DeleteByStructFilter(ctx, &model.Dish{DishCode: code}); err != nil {
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
		if err := ds.imgUtil.Delete(filePath); err != nil {
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

	var dishes []*model.Dish
	var dishIngredients []*model.DishIngredient
	var imageURLs []*model.DishImage

	for rowNum := 2; rowNum <= len(rows); rowNum++ {

		fmt.Printf("\n处理第 %d 行,行长度 %d...\n", rowNum, len(rows[rowNum-1]))

		if rowNum%batchSize == 0 && (len(dishes) > 0 || len(dishIngredients) > 0) {

			fmt.Printf("已处理 %d 行，准备提交事务...\n", rowNum)

			if err := ds.db.Transaction(ctx, func(ctx context.Context, db *gormx.DB) error {
				// 第一步：创建菜品
				if len(dishes) > 0 {
					if err := db.CreateInBatches(ctx, dishes, batchSize,
						gormx.OnConstraintColumns("dish_code"),
						gormx.UpdateAll(),
					); err != nil {
						return fmt.Errorf("创建产品失败: %w", err)
					}
				}
				// 第二步：创建菜品食材关系
				if len(dishIngredients) > 0 {
					if err := db.CreateInBatches(ctx, dishIngredients, batchSize,
						gormx.OnConstraintColumns("dish_code", "ingredient_code"),
						gormx.UpdateAll(),
					); err != nil {
						return fmt.Errorf("创建产品失败: %w", err)
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
					dishes = append(dishes, &model.Dish{
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
					dishIngredients = append(dishIngredients, &model.DishIngredient{
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

				if err := ds.db.GetXDB().CreateInBatches(ctx, imageURLs, batchSize,
					gormx.OnConstraintColumns("dish_code", "sort_order"),
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

				url, err := utils.ProcessExcelPicture(
					ds.imgUtil,
					picture[0],
					[]string{"uploads", "dishes"},
					dishCode,
					sortOrder,
				)
				if err != nil {
					continue
				}

				imageURLs = append(imageURLs, &model.DishImage{
					DishCode: dishCode,
					// 排序,用于显示顺序
					SortOrder: sortOrder,
					ImageURL:  url,
				})
			}
		}

		if len(dishes) > 0 || len(dishIngredients) > 0 || len(imageURLs) > 0 {

			if err := ds.db.Transaction(ctx, func(ctx context.Context, db *gormx.DB) error {
				// 第一步：创建菜品
				if len(dishes) > 0 {
					if err := db.CreateInBatches(ctx, dishes, batchSize,
						gormx.OnConstraintColumns("dish_code"),
						gormx.UpdateAll(),
					); err != nil {
						return fmt.Errorf("创建产品失败: %w", err)
					}
				}
				// 第二步：创建菜品食材关系
				if len(dishIngredients) > 0 {
					if err := db.CreateInBatches(ctx, dishIngredients, batchSize,
						gormx.OnConstraintColumns("dish_code", "ingredient_code"),
						gormx.UpdateAll(),
					); err != nil {
						return fmt.Errorf("创建产品失败: %w", err)
					}
				}
				// 第三步：创建菜品图片关系
				if len(imageURLs) > 0 {

					if err := db.CreateInBatches(ctx, imageURLs, batchSize,
						gormx.OnConstraintColumns("dish_code", "sort_order"),
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
		FindInBatches(ctx, batchSize, func(ctx context.Context, tx *gormx.DB, batch int, models *[]model.Dish) error {
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
