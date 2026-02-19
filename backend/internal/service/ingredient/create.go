package ingredientService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"

	"github.com/LouYuanbo1/go-webservice/gormx/options"
)

func (is *ingredientService) Create(ctx context.Context, req *dto.CreateIngredientRequest) error {
	// 处理图片
	imageURLs := make([]*model.IngredientImage, 0, len(req.Images))
	for i, fileHeader := range req.Images {
		url, err := is.processImage(fileHeader, req.IngredientCode)
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

	err := is.repoFactory.Tx().Exec(ctx, func(ctx context.Context) error {
		// 第一步：创建产品
		if err := is.repoFactory.Ingredient().Create(ctx, &model.Ingredient{
			IngredientCode: req.IngredientCode,
			Name:           req.Name,
			Description:    req.Description,
		},
			options.OnConstraintColumns("ingredient_code"),
			options.UpdateAllOption(),
		); err != nil {
			return fmt.Errorf("创建产品失败: %w", err)
		}
		// 第二步：创建产品图片关系
		if err := is.repoFactory.IngredientImage().CreateInBatches(ctx, imageURLs, 10,
			options.OnConstraintColumns("ingredient_code", "sort_order"),
			options.UpdateAllOption(),
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
