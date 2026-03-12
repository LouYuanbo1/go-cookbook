package ingredientService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"
	"go-cookbook/internal/utils"

	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/LouYuanbo1/go-webservice/gormx/gen"
)

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

	err := is.repoFactory.Tx().Exec(ctx, func(ctx context.Context, sf *gen.SessionFactory) error {
		ingredientSession := gen.NewSessionFromFactory[model.Ingredient, uint64](sf)
		ingredientImageSession := gen.NewSessionFromFactory[model.IngredientImage, uint64](sf)

		// 第一步：创建产品
		if err := ingredientSession.Create(ctx, &model.Ingredient{
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
		if err := ingredientImageSession.CreateInBatches(ctx, imageURLs, 10,
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
