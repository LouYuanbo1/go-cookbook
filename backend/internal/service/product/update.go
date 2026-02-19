package productService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"

	"github.com/LouYuanbo1/go-webservice/gormx/options"
)

func (ps *productService) Update(ctx context.Context, req *dto.UpdateProductRequest) error {
	savedNewImages := make(map[string]string) // tempID -> url
	for _, img := range req.NewImages {
		url, err := ps.processImage(img.File, req.ProductCode)
		if err != nil {
			return fmt.Errorf("处理新图片失败: %w", err)
		}
		savedNewImages[img.TempID] = url
	}

	err := ps.repoFactory.Tx().Exec(ctx, func(ctx context.Context) error {

		// 更新产品基本信息
		ps.repoFactory.Product().UpdateByStructFilter(ctx, &model.Product{ProductCode: req.ProductCode}, &model.Product{
			IngredientCode: req.IngredientCode,
			Name:           req.Name,
			Amount:         req.Amount,
			Unit:           req.Unit,
			Description:    req.Description,
			Price:          req.Price,
			AllergenType:   req.AllergenType,
		})

		deletedImages := make([]uint64, 0)
		upsertImages := make([]*model.ProductImage, 0)

		// 删除标记为deleted的图片
		for _, img := range req.Images {
			switch img.Type {
			case "deleted":
				deletedImages = append(deletedImages, img.ID)
			case "existing":
				upsertImages = append(upsertImages, &model.ProductImage{
					ID:        img.ID,
					SortOrder: img.SortOrder,
				})
			case "new":
				url, exists := savedNewImages[img.TempID]
				if exists {
					upsertImages = append(upsertImages, &model.ProductImage{
						ProductCode: req.ProductCode,
						ImageURL:    url,
						SortOrder:   img.SortOrder,
					})
				}
			default:
				return fmt.Errorf("未知图片操作类型: %s", img.Type)
			}
		}

		// 第二步：删除图片
		if err := ps.repoFactory.ProductImage().DeleteByIDs(ctx, deletedImages); err != nil {
			return fmt.Errorf("删除产品图片关系失败: %w", err)
		}

		// 第三步：更新或插入图片关系
		if err := ps.repoFactory.ProductImage().CreateInBatches(
			ctx,
			upsertImages,
			10,
			options.OnConstraintColumns("id"),
			options.UpdateColumnsOption("sort_order"),
		); err != nil {
			return fmt.Errorf("创建新图片失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("更新菜品失败: %w", err)
	}
	return nil
}
