package ingredientService

import (
	"context"
	"fmt"
	"go-cookbook/internal/model"
	"os"
	"path/filepath"

	"github.com/LouYuanbo1/go-webservice/gormx/gen"
)

func (is *ingredientService) Delete(ctx context.Context, code string) error {
	var images []*model.IngredientImage
	var err error
	err = is.repoFactory.Tx().Exec(ctx, func(ctx context.Context, sf *gen.SessionFactory) error {
		ingredientImageSession := gen.NewSessionFromFactory[model.IngredientImage, uint64](sf)
		ingredientSession := gen.NewSessionFromFactory[model.Ingredient, uint64](sf)

		// 第一步：查询所有关联的图片
		err = ingredientImageSession.FindByStructFilter(ctx, &images, &model.IngredientImage{IngredientCode: code})
		if err != nil {
			return fmt.Errorf("查询食材图片关系失败: %w", err)
		}

		// 第二步：删除食材图片关系
		if err := ingredientImageSession.DeleteByStructFilter(ctx, &model.IngredientImage{IngredientCode: code}); err != nil {
			return fmt.Errorf("删除食材图片关系失败: %w", err)
		}

		// 第三步：删除食材
		if err := ingredientSession.DeleteByStructFilter(ctx, &model.Ingredient{IngredientCode: code}); err != nil {
			return fmt.Errorf("删除食材失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("删除食材失败: %w", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	//第二步:删除图片
	for _, img := range images {
		filePath := filepath.Join(wd, img.ImageURL)

		fmt.Printf("删除图片路径: %s\n", filePath)
		if err := is.imgUtil.Delete(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
}
