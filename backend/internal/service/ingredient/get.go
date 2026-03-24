package ingredientService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"

	"github.com/LouYuanbo1/go-webservice/gormx"
)

func (is *ingredientService) GetByCode(ctx context.Context, code string) (*dto.ViewIngredientResponse, error) {
	var ingredient model.Ingredient
	err := is.repoFactory.Ingredient().GetByStructFilter(ctx, &ingredient, &model.Ingredient{IngredientCode: code})
	if err != nil {
		return nil, fmt.Errorf("查询食材基本信息失败: %w", err)
	}

	ingredientResp := dto.ViewIngredientResponse{
		IngredientCode: ingredient.IngredientCode,
		Name:           ingredient.Name,
		Description:    ingredient.Description,
	}

	var images []*model.IngredientImage
	err = is.repoFactory.IngredientImage().FindByStructFilter(
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
