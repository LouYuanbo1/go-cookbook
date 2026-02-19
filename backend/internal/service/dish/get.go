package dishService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"

	"github.com/LouYuanbo1/go-webservice/gormx/options"
)

func (ds *dishService) GetByCode(ctx context.Context, code string) (*dto.ViewDishResponse, error) {
	dish, err := ds.repoFactory.Dish().GetByStructFilter(ctx, &model.Dish{DishCode: code})
	if err != nil {
		return nil, fmt.Errorf("查询菜品基本信息失败: %w", err)
	}

	dishResp := dto.ViewDishResponse{
		DishCode:    dish.DishCode,
		Name:        dish.Name,
		Description: dish.Description,
		Recipe:      dish.Recipe,
	}

	images, err := ds.repoFactory.DishImage().FindByStructFilter(ctx, &model.DishImage{DishCode: code}, options.WithAscOption("order"))
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
