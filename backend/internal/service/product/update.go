package productService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"
	"go-cookbook/internal/utils"
	"mime/multipart"

	"github.com/LouYuanbo1/go-webservice/gormx/options"
)

func (ps *productService) Update(ctx context.Context, req *dto.UpdateProductRequest) error {
	tempIDToFileHeader := make(map[string]*multipart.FileHeader) // tempID -> fileHeader
	for _, img := range req.NewImages {
		tempIDToFileHeader[img.TempID] = img.File
	}

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
			fmt.Println("New Image:", img.TempID)
			url, err := utils.ProcessImageFileHeader(
				ps.imgUtil,
				tempIDToFileHeader[img.TempID],
				[]string{"uploads", "products"},
				req.ProductCode,
				img.SortOrder,
			)
			if err != nil {
				return fmt.Errorf("处理新图片失败: %w", err)
			}
			fmt.Printf("TempID: %s, URL: %s", img.TempID, url)
			upsertImages = append(upsertImages, &model.ProductImage{
				ProductCode: req.ProductCode,
				ImageURL:    url,
				SortOrder:   img.SortOrder,
			})
		default:
			return fmt.Errorf("未知图片操作类型: %s", img.Type)
		}
	}

	err := ps.repoFactory.Tx().Exec(ctx, func(ctx context.Context) error {
		//第一步:更新产品基本信息
		ps.repoFactory.Product().UpdateByStructFilter(ctx, &model.Product{ProductCode: req.ProductCode}, &model.Product{
			IngredientCode: req.IngredientCode,
			Name:           req.Name,
			Amount:         req.Amount,
			Unit:           req.Unit,
			Description:    req.Description,
			Price:          req.Price,
			AllergenType:   req.AllergenType,
		})

		// 第二步：删除图片
		if len(deletedImages) > 0 {
			if err := ps.repoFactory.ProductImage().DeleteByIDs(ctx, deletedImages); err != nil {
				return fmt.Errorf("删除产品图片关系失败: %w", err)
			}
		}

		// 第三步：更新或插入图片关系
		if len(upsertImages) > 0 {
			if err := ps.repoFactory.ProductImage().CreateInBatches(
				ctx,
				upsertImages,
				10,
				options.OnConstraintColumns("id"),
				options.UpdateColumnsOption("sort_order"),
			); err != nil {
				return fmt.Errorf("创建新图片失败: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("更新菜品失败: %w", err)
	}
	return nil
}
