package productService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"
	"go-cookbook/internal/utils"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/LouYuanbo1/go-webservice/gormx/gen"
)

func (ps *productService) Update(ctx context.Context, req *dto.UpdateProductRequest) error {

	fmt.Printf("UpdateProductRequest: %v\n", req)

	tempIDToFileHeader := make(map[string]*multipart.FileHeader) // tempID -> fileHeader
	for _, img := range req.NewImages {
		tempIDToFileHeader[img.TempID] = img.File
	}

	deletedImages := make([]uint64, 0)
	deletedImageURLs := make([]string, 0)
	upsertImages := make([]*model.ProductImage, 0)

	// 删除标记为deleted的图片
	for _, img := range req.Images {
		switch img.Type {
		case "deleted":
			deletedImages = append(deletedImages, img.ID)
			img, err := ps.repoFactory.ProductImage().GetByID(ctx, img.ID)
			if err != nil {
				return fmt.Errorf("查询删除图片失败: %w", err)
			}
			deletedImageURLs = append(deletedImageURLs, img.ImageURL)
		case "existing":
			upsertImages = append(upsertImages, &model.ProductImage{
				ID:          img.ID,
				ProductCode: req.ProductCode,
				SortOrder:   img.SortOrder,
			})
			ps.repoFactory.ProductImage().UpdateByStructFilter(ctx, &model.ProductImage{
				ID: img.ID,
			}, &model.ProductImage{
				//使用负顺序,避免唯一约束影响调换顺序
				SortOrder: -img.SortOrder,
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

	err := ps.repoFactory.Tx().Exec(ctx, func(ctx context.Context, sf *gen.SessionFactory) error {
		productSession := gen.NewSessionFromFactory[model.Product, uint64](sf)
		productImageSession := gen.NewSessionFromFactory[model.ProductImage, uint64](sf)

		//第一步:更新产品基本信息
		productSession.UpdateByStructFilter(ctx, &model.Product{ProductCode: req.ProductCode}, &model.Product{
			IngredientCode: req.IngredientCode,
			Name:           req.Name,
			Amount:         req.Amount,
			Unit:           req.Unit,
			Description:    req.Description,
			Price:          req.Price,
			AllergenType:   req.AllergenType,
		})

		fmt.Printf("DeletedImages: %v\n", deletedImages)

		// 第二步：删除图片
		if len(deletedImages) > 0 {
			if err := productImageSession.DeleteByIDs(ctx, deletedImages); err != nil {
				return fmt.Errorf("删除产品图片关系失败: %w", err)
			}
		}

		fmt.Printf("UpsertImages: %v\n", upsertImages)

		// 第三步：更新或插入图片关系
		if len(upsertImages) > 0 {
			if err := productImageSession.CreateInBatches(
				ctx,
				upsertImages,
				10,
				gormx.OnConstraintColumns("id"),
				gormx.UpdateColumns("sort_order"),
			); err != nil {
				return fmt.Errorf("创建新图片失败: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("更新菜品失败: %w", err)
	}

	// 第五步:删除图片文件
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	for _, url := range deletedImageURLs {
		filePath := filepath.Join(wd, url)
		fmt.Printf("删除图片路径: %s\n", filePath)
		if err := ps.imgUtil.Delete(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
}
