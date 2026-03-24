package productService

import (
	"context"
	"fmt"
	"go-cookbook/internal/model"
	"os"
	"path/filepath"

	"github.com/LouYuanbo1/go-webservice/gormx/gen"
)

func (ps *productService) Delete(ctx context.Context, code string) error {
	var images []*model.ProductImage
	var err error
	err = ps.repoFactory.Tx().Exec(ctx, func(ctx context.Context, sf *gen.SessionFactory) error {
		productSession := gen.NewSessionFromFactory[model.Product, uint64](sf)
		productImageSession := gen.NewSessionFromFactory[model.ProductImage, uint64](sf)

		// 第一步：查询所有关联的图片
		err = productImageSession.FindByStructFilter(ctx, &images, &model.ProductImage{ProductCode: code})
		if err != nil {
			return fmt.Errorf("查询产品图片关系失败: %w", err)
		}

		// 第二步：删除产品图片关系
		if err := productImageSession.DeleteByStructFilter(ctx, &model.ProductImage{ProductCode: code}); err != nil {
			return fmt.Errorf("删除产品图片关系失败: %w", err)
		}

		// 第三步：删除产品
		if err := productSession.DeleteByStructFilter(ctx, &model.Product{ProductCode: code}); err != nil {
			return fmt.Errorf("删除产品失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("删除产品失败: %w", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	//第二步:删除图片
	for _, img := range images {
		filePath := filepath.Join(wd, img.ImageURL)

		fmt.Printf("删除图片路径: %s\n", filePath)

		if err := ps.imgUtil.Delete(filePath); err != nil {
			return fmt.Errorf("删除图片失败: %w", err)
		}
	}
	return nil
}
