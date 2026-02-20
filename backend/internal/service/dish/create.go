package dishService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"
	"go-cookbook/internal/utils"

	"github.com/LouYuanbo1/go-webservice/gormx/options"
)

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

	err := ds.repoFactory.Tx().Exec(ctx, func(ctx context.Context) error {
		// 第一步：创建菜品
		if err := ds.repoFactory.Dish().Create(ctx, &model.Dish{
			DishCode:    req.DishCode,
			Name:        req.Name,
			Description: req.Description,
			Recipe:      req.Recipe,
		},
			options.OnConstraintColumns("dish_code"),
			options.UpdateAllOption(),
		); err != nil {
			return fmt.Errorf("创建菜品失败: %w", err)
		}
		// 第二步：创建菜品图片
		if err := ds.repoFactory.DishImage().CreateInBatches(ctx, imageURLs, 10,
			options.OnConstraintColumns("dish_code", "sort_order"),
			options.UpdateAllOption(),
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

		if err := ds.repoFactory.DishIngredient().CreateInBatches(ctx, ingredients, 10,
			options.OnConstraintColumns("dish_code", "ingredient_code"),
			options.UpdateAllOption(),
		); err != nil {
			return fmt.Errorf("创建菜品食材关系失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("创建菜品和食材关系失败: %w", err)
	}
	return nil
}
