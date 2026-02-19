package productService

import (
	"context"
	"fmt"
	"go-cookbook/internal/model"
)

func (ps *productService) Delete(ctx context.Context, code string) error {
	err := ps.repoFactory.Tx().Exec(ctx, func(ctx context.Context) error {
		images, err := ps.repoFactory.ProductImage().FindByStructFilter(ctx, &model.ProductImage{ProductCode: code})
		if err != nil {
			return fmt.Errorf("查询产品图片关系失败: %w", err)
		}

		// 第一步：删除产品图片关系
		if err := ps.repoFactory.ProductImage().DeleteByStructFilter(ctx, &model.ProductImage{ProductCode: code}); err != nil {
			return fmt.Errorf("删除产品图片关系失败: %w", err)
		}

		//第二步:删除图片
		for _, img := range images {
			if err := ps.imgUtil.Delete(img.ImageURL); err != nil {
				return fmt.Errorf("删除图片失败: %w", err)
			}
		}

		// 第三步：删除产品
		if err := ps.repoFactory.Product().DeleteByStructFilter(ctx, &model.Product{ProductCode: code}); err != nil {
			return fmt.Errorf("删除产品失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("删除产品失败: %w", err)
	}
	return nil
}
