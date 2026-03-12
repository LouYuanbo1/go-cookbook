package ingredientService

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

func (is *ingredientService) Update(ctx context.Context, req *dto.UpdateIngredientRequest) error {
	tempIDToFileHeader := make(map[string]*multipart.FileHeader) // tempID -> fileHeader
	for _, img := range req.NewImages {
		tempIDToFileHeader[img.TempID] = img.File
	}

	deletedImages := make([]uint64, 0)
	deletedImageURLs := make([]string, 0)
	upsertImages := make([]*model.IngredientImage, 0)
	// 删除标记为deleted的图片
	for _, img := range req.Images {
		switch img.Type {
		case "deleted":
			deletedImages = append(deletedImages, img.ID)
			img, err := is.repoFactory.IngredientImage().GetByID(ctx, img.ID)
			if err != nil {
				return fmt.Errorf("查询删除图片失败: %w", err)
			}
			deletedImageURLs = append(deletedImageURLs, img.ImageURL)
		case "existing":
			upsertImages = append(upsertImages, &model.IngredientImage{
				ID:             img.ID,
				IngredientCode: req.IngredientCode,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
			is.repoFactory.IngredientImage().UpdateByStructFilter(ctx, &model.IngredientImage{
				ID: img.ID,
			}, &model.IngredientImage{
				//使用负顺序,避免唯一约束影响调换顺序
				SortOrder: -img.SortOrder,
			})
		case "new":
			fmt.Println("New Image:", img.TempID)
			url, err := utils.ProcessImageFileHeader(
				is.imgUtil,
				tempIDToFileHeader[img.TempID],
				[]string{"uploads", "ingredients"},
				req.IngredientCode,
				img.SortOrder,
			)
			if err != nil {
				return fmt.Errorf("处理新图片失败: %w", err)
			}
			fmt.Printf("TempID: %s, URL: %s", img.TempID, url)
			upsertImages = append(upsertImages, &model.IngredientImage{
				IngredientCode: req.IngredientCode,
				ImageURL:       url,
				// 排序,用于显示顺序
				SortOrder: img.SortOrder,
			})
		default:
			return fmt.Errorf("未知图片操作类型: %s", img.Type)
		}
	}

	err := is.repoFactory.Tx().Exec(ctx, func(ctx context.Context, sf *gen.SessionFactory) error {
		ingredientSession := gen.NewSessionFromFactory[model.Ingredient, uint64](sf)
		ingredientImageSession := gen.NewSessionFromFactory[model.IngredientImage, uint64](sf)
		//第一步:更新食材基本信息
		ingredientSession.UpdateByStructFilter(ctx, &model.Ingredient{
			IngredientCode: req.IngredientCode,
		}, &model.Ingredient{
			Name:        req.Name,
			Description: req.Description,
		})

		fmt.Printf("deletedImages: %v\n", deletedImages)

		if len(deletedImages) > 0 {
			// 第二步：删除图片
			if err := ingredientImageSession.DeleteByIDs(ctx, deletedImages); err != nil {
				return fmt.Errorf("删除食材图片关系失败: %w", err)
			}
		}

		fmt.Printf("upsertImages: %v\n", upsertImages)

		if len(upsertImages) > 0 {
			// 第三步：更新或插入图片关系
			if err := ingredientImageSession.CreateInBatches(
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
		return fmt.Errorf("更新食材失败: %w", err)
	}

	// 第五步:删除图片文件
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	for _, url := range deletedImageURLs {
		filePath := filepath.Join(wd, url)
		fmt.Printf("删除图片路径: %s\n", filePath)
		if err := is.imgUtil.Delete(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
}
