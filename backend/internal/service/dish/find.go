package dishService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
)

func (ds *dishService) FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.ViewDishCardListWithCursor, error) {
	var results []struct {
		ID         uint64 `gorm:"column:id"`
		DishCode   string `gorm:"column:dish_code"`
		Name       string `gorm:"column:name"`
		ImageID    uint64 `gorm:"column:img_id"`
		ImageOrder int    `gorm:"column:img_order"`
		ImageURL   string `gorm:"column:img_url"`
	}
	result := ds.repoFactory.Dish().GetDBWithContext(ctx).
		Table("dishes d").
		Joins(`
			LEFT JOIN (
				SELECT DISTINCT ON (dish_code) 
					dish_code, 
					id, 
					sort_order, 
					image_url
				FROM dish_images
				ORDER BY dish_code, sort_order ASC
			) di ON di.dish_code = d.dish_code
		`).
		Select(`
			d.id,
			d.dish_code,
			d.name,
			di.id AS img_id,
			di.sort_order AS img_order,
			di.image_url AS img_url
		`).
		Where("d.id > ?", cursor).
		Order("d.id ASC").
		Limit(limit + 1).
		Find(&results)
	if result.Error != nil {
		return nil, fmt.Errorf("查询产品失败: %w", result.Error)
	}
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

	fmt.Printf("FindByCursor results: %v\n", results)

	dishes := make([]*dto.ViewDishCard, 0, len(results))
	for _, r := range results {
		dish := &dto.ViewDishCard{
			DishCode: r.DishCode,
			Name:     r.Name,
		}
		if r.ImageID > 0 {
			dish.Image = dto.ImageResponse{
				ID:        r.ImageID,
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

/*
[
  {
    IngredientCode: "ING001",
    Name: "面粉",
    Quantity: "200g",
    Note: "高筋面粉",
    ImageID: 1,
    ImageOrder: 1,
    ImageURL: "flour1.jpg"
  },
  {
    IngredientCode: "ING001",
    Name: "面粉",
    Quantity: "200g",
    Note: "高筋面粉",
    ImageID: 2,
    ImageOrder: 2,
    ImageURL: "flour2.jpg"
  },
  {
    IngredientCode: "ING002",
    Name: "鸡蛋",
    Quantity: "2个",
    Note: "新鲜鸡蛋",
    ImageID: 3,
    ImageOrder: 1,
    ImageURL: "egg1.jpg"
  }
]
*/

// 优化,避免N+1查询问题
// 注意,这里使用了Postgresql方言,破坏了兼容性,迁移到其他数据库时需要修改
func (ds *dishService) FindIngredientsByDishCode(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewDishIngredientCardListWithCursor, error) {
	var results []struct {
		ID             uint64 `gorm:"column:id"`
		IngredientCode string
		Name           string
		Quantity       string
		Note           string
		ImageID        uint64 `gorm:"column:img_id"`
		ImageOrder     int    `gorm:"column:img_order"`
		ImageURL       string `gorm:"column:img_url"`
	}

	// 一次性获取所有数据
	err := ds.repoFactory.Ingredient().GetDBWithContext(ctx).
		Table("ingredients i").
		Joins("JOIN dish_ingredients di ON i.ingredient_code = di.ingredient_code").
		Joins(`
			LEFT JOIN (
				SELECT DISTINCT ON (ingredient_code) 
					ingredient_code, 
					id, 
					sort_order, 
					image_url
				FROM ingredient_images
				ORDER BY ingredient_code, sort_order ASC
			) ii ON ii.ingredient_code = i.ingredient_code
		`).
		Select(`
			i.id as id,
			i.ingredient_code,
			i.name,
			di.quantity,
			di.note,
			ii.id as img_id,
			ii.sort_order as img_order,
			ii.image_url as img_url
		`).
		Where("di.dish_code = ?", code).
		Where("i.id > ?", cursor).
		Order("i.id ASC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("查询食材信息失败: %w", err)
	}

	hasMore := len(results) > limit
	if hasMore {
		results = results[:limit]
	}
	newCursor := cursor
	if len(results) > 0 {
		newCursor = results[len(results)-1].ID
	}

	// 将结果分组为 ViewDishIngredientCard
	ingredients := make([]*dto.ViewDishIngredientCard, 0, len(results))
	for _, r := range results {
		ingredient := &dto.ViewDishIngredientCard{
			ID:             r.ID,
			IngredientCode: r.IngredientCode,
			Name:           r.Name,
			Quantity:       r.Quantity,
			Note:           r.Note,
			Image:          dto.ImageResponse{},
		}
		if r.ImageID > 0 {
			ingredient.Image = dto.ImageResponse{
				ID:        r.ImageID,
				SortOrder: r.ImageOrder,
				ImageURL:  r.ImageURL,
			}
		}
		ingredients = append(ingredients, ingredient)
	}

	return &dto.ViewDishIngredientCardListWithCursor{
		DishIngredients: ingredients,
		Cursor:          newCursor,
		HasMore:         hasMore,
	}, nil
}
