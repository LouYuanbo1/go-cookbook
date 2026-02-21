package dishService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"
	"go-cookbook/internal/utils"
	"mime/multipart"

	"github.com/LouYuanbo1/go-webservice/gormx/options"
)

/*
e.g.

	{
	  "dishCode": "D001",
	  "name": "宫保鸡丁",
	  "description": "经典川菜",
	  "images": [
	    {"type": "existing", "id": 10, "order": 0},
	    {"type": "new", "id": 0, "order": 1},      // 对应 newImages[0]
	    {"type": "existing", "id": 15, "order": 2},
	    {"type": "deleted", "id": 20, "order": -1} // 要删除的图片
	  ]
	}
*/

// 逻辑复杂,可能有问题,需要测试
func (ds *dishService) Update(ctx context.Context, req *dto.UpdateDishRequest) error {
	tempIDToFileHeader := make(map[string]*multipart.FileHeader) // tempID -> fileHeader
	for _, img := range req.NewImages {
		tempIDToFileHeader[img.TempID] = img.File
	}

	deletedImages := make([]uint64, 0)
	upsertImages := make([]*model.DishImage, 0) // id -> true
	for _, img := range req.Images {
		switch img.Type {
		case "deleted":
			deletedImages = append(deletedImages, img.ID)
		case "existing":
			upsertImages = append(upsertImages, &model.DishImage{
				ID:       img.ID,
				DishCode: req.DishCode,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
			ds.repoFactory.DishImage().UpdateByStructFilter(ctx, &model.DishImage{
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
			if err := ds.repoFactory.DishIngredient().DeleteByStructFilter(
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

	err := ds.repoFactory.Tx().Exec(ctx, func(ctx context.Context) error {
		// 第一步:更新菜品基本信息
		err := ds.repoFactory.Dish().UpdateByStructFilter(ctx, &model.Dish{
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
		if err := ds.repoFactory.DishImage().DeleteByIDs(ctx, deletedImages); err != nil {
			return fmt.Errorf("删除删除图片失败: %w", err)
		}

		// 第三步:创建或更新图片
		if err := ds.repoFactory.DishImage().CreateInBatches(
			ctx,
			upsertImages,
			10,
			options.OnConstraintColumns("id"),
			options.UpdateColumnsOption("sort_order"),
		); err != nil {
			return fmt.Errorf("创建新图片失败: %w", err)
		}

		// 第四步:创建或更新食材
		if err := ds.repoFactory.DishIngredient().CreateInBatches(
			ctx,
			upsertIngredients,
			10,
			options.OnConstraintColumns("dish_code", "ingredient_code"),
			options.UpdateColumnsOption("quantity", "note"),
		); err != nil {
			return fmt.Errorf("创建新食材失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("更新菜品失败: %w", err)
	}
	return nil
}
