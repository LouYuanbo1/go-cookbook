package productService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
)

func (ps *productService) FindDishesByProductCodeAndCursor(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewDishCardListWithCursor, error) {
	var results []struct {
		ID         uint64 `gorm:"column:id"`
		DishCode   string `gorm:"column:dish_code"`
		Name       string `gorm:"column:name"`
		ImageID    uint64 `gorm:"column:img_id"`
		ImageOrder int    `gorm:"column:img_order"`
		ImageURL   string `gorm:"column:img_url"`
	}

	result := ps.repoFactory.Dish().GetDBWithContext(ctx).
		Table("dishes d").
		Joins("JOIN dish_ingredients di ON di.dish_code = d.dish_code").
		Joins("JOIN products p ON di.ingredient_code = p.ingredient_code").
		Joins(
			`
			LEFT JOIN (
				SELECT DISTINCT ON (dish_code) 
					dish_code, 
					id, 
					sort_order, 
					image_url
				FROM dish_images
				ORDER BY dish_code, sort_order ASC
			) dimages ON dimages.dish_code = d.dish_code
			`).
		Select(`
				d.id,
				d.dish_code,
				d.name,
				dimages.id AS img_id,
				dimages.sort_order AS img_order,
				dimages.image_url AS img_url
			`).
		Where("p.product_code = ?", code).
		Where("d.id > ?", cursor).
		Limit(limit + 1).
		Find(&results)

	if result.Error != nil {
		return nil, fmt.Errorf("查询食材信息失败: %w", result.Error)
	}

	hasMore := len(results) > limit
	if hasMore {
		results = results[:limit]
	}
	newCursor := cursor
	if len(results) > 0 {
		newCursor = results[len(results)-1].ID
	}

	// 将结果分组为 ViewDishCard
	dishes := make([]*dto.ViewDishCard, 0, len(results))
	for _, r := range results {
		dish := &dto.ViewDishCard{
			ID:       r.ID,
			DishCode: r.DishCode,
			Name:     r.Name,
			Image:    dto.ImageResponse{},
		}
		// 仅当存在封面图时添加
		if r.ImageID > 0 {
			dish.Image = dto.ImageResponse{
				ID: r.ImageID,
				// 排序,用于显示顺序
				SortOrder: r.ImageOrder,
				ImageURL:  r.ImageURL,
			}
		}
		dishes = append(dishes, dish)
	}

	return &dto.ViewDishCardListWithCursor{
		Dishes:  dishes,
		Cursor:  newCursor,
		HasMore: hasMore,
	}, nil
}
