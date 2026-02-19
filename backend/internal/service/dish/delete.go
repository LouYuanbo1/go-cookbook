package dishService

import (
	"context"
	"fmt"
	"go-cookbook/internal/model"
)

func (ds *dishService) Delete(ctx context.Context, code string) error {
	err := ds.repoFactory.Tx().Exec(ctx, func(ctx context.Context) error {
		// 第一步：删除菜品食材关系
		if err := ds.repoFactory.DishIngredient().DeleteByStructFilter(ctx, &model.DishIngredient{DishCode: code}); err != nil {
			return fmt.Errorf("删除菜品食材关系失败: %w", err)
		}

		images, err := ds.repoFactory.DishImage().FindByStructFilter(ctx, &model.DishImage{DishCode: code})
		if err != nil {
			return fmt.Errorf("查询菜品图片关系失败: %w", err)
		}

		for _, img := range images {
			if err := ds.imgUtil.Delete(img.ImageURL); err != nil {
				return fmt.Errorf("删除图片失败: %w", err)
			}
		}

		// 第二步：删除图片关系
		if err := ds.repoFactory.DishImage().DeleteByStructFilter(ctx, &model.DishImage{DishCode: code}); err != nil {
			return fmt.Errorf("删除菜品图片失败: %w", err)
		}

		// 第三步：删除菜品图片
		for _, img := range images {
			if err := ds.imgUtil.Delete(img.ImageURL); err != nil {
				return fmt.Errorf("删除图片失败: %w", err)
			}
		}

		// 第四步：删除菜品
		if err := ds.repoFactory.Dish().DeleteByStructFilter(ctx, &model.Dish{DishCode: code}); err != nil {
			return fmt.Errorf("删除菜品失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("删除菜品失败: %w", err)
	}
	return nil
}
