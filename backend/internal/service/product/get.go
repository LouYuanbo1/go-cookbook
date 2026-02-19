package productService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"

	"github.com/LouYuanbo1/go-webservice/gormx/options"
)

func (ps *productService) GetByCode(ctx context.Context, code string) (*dto.ViewProductResponse, error) {
	product, err := ps.repoFactory.Product().GetByStructFilter(ctx, &model.Product{ProductCode: code})
	if err != nil {
		return nil, fmt.Errorf("查询产品基本信息失败: %w", err)
	}

	productResp := dto.ViewProductResponse{
		ProductCode:    product.ProductCode,
		IngredientCode: product.IngredientCode,
		Name:           product.Name,
		Amount:         product.Amount,
		Unit:           product.Unit,
		Description:    product.Description,
		Price:          product.Price,
		AllergenType:   product.AllergenType,
	}

	images, err := ps.repoFactory.ProductImage().FindByStructFilter(
		ctx,
		&model.ProductImage{ProductCode: code},
		options.WithAscOption("order"),
	)
	if err != nil {
		return nil, fmt.Errorf("查询产品图片关系失败: %w", err)
	}

	productResp.Images = make([]dto.ImageResponse, 0, len(images))
	for _, img := range images {
		productResp.Images = append(productResp.Images, dto.ImageResponse{
			ID: img.ID,
			// 排序,用于显示顺序
			SortOrder: img.SortOrder,
			ImageURL:  img.ImageURL,
		})
	}

	return &productResp, nil
}
