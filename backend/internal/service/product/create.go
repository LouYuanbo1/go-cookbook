package productService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"
	"go-cookbook/internal/utils"

	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/LouYuanbo1/go-webservice/gormx/gen"
)

func (ps *productService) Create(ctx context.Context, req *dto.CreateProductRequest) error {
	// 处理图片
	imageURLs := make([]*model.ProductImage, 0, len(req.Images))
	for i, fileHeader := range req.Images {
		url, err := utils.ProcessImageFileHeader(
			ps.imgUtil,
			fileHeader,
			[]string{"uploads", "products"},
			req.ProductCode,
			i,
		)
		if err != nil {
			return fmt.Errorf("处理图片失败: %w", err)
		}
		imageURLs = append(imageURLs, &model.ProductImage{
			ProductCode: req.ProductCode,
			// 排序,用于显示顺序
			SortOrder: i,
			ImageURL:  url,
		})
	}

	err := ps.repoFactory.Tx().Exec(ctx, func(ctx context.Context, sf *gen.SessionFactory) error {
		productSession := gen.NewSessionFromFactory[model.Product, uint64](sf)
		productImageSession := gen.NewSessionFromFactory[model.ProductImage, uint64](sf)

		// 第一步：创建产品
		if err := productSession.Create(ctx, &model.Product{
			ProductCode:    req.ProductCode,
			IngredientCode: req.IngredientCode,
			Name:           req.Name,
			Amount:         req.Amount,
			Unit:           req.Unit,
			Description:    req.Description,
			Price:          req.Price,
			AllergenType:   req.AllergenType,
		},
			gormx.OnConstraintColumns("product_code"),
			gormx.UpdateAll(),
		); err != nil {
			return fmt.Errorf("创建产品失败: %w", err)
		}
		// 第二步：创建产品图片关系
		if err := productImageSession.CreateInBatches(ctx, imageURLs, 10,
			gormx.OnConstraintColumns("product_code", "sort_order"),
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
