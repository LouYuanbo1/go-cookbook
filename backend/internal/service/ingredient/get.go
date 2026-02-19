package ingredientService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"

	"github.com/LouYuanbo1/go-webservice/gormx/options"
)

func (is *ingredientService) GetByCode(ctx context.Context, code string) (*dto.ViewIngredientResponse, error) {
	ingredient, err := is.repoFactory.Ingredient().GetByStructFilter(ctx, &model.Ingredient{IngredientCode: code})
	if err != nil {
		return nil, fmt.Errorf("查询产品基本信息失败: %w", err)
	}

	ingredientResp := dto.ViewIngredientResponse{
		IngredientCode: ingredient.IngredientCode,
		Name:           ingredient.Name,
		Description:    ingredient.Description,
	}

	images, err := is.repoFactory.IngredientImage().FindByStructFilter(
		ctx,
		&model.IngredientImage{IngredientCode: code},
		options.WithAscOption("order"),
	)
	if err != nil {
		return nil, fmt.Errorf("查询产品图片关系失败: %w", err)
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
