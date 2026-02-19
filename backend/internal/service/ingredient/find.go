package ingredientService

import (
	"context"
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"
)

// 注意,这里使用了Postgresql方言(ON DISTINCT),破坏了兼容性,迁移到其他数据库时需要修改
func (is *ingredientService) FindProductsByIngredientCodeAndCursor(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewProductCardListWithCursor, error) {
	var results []struct {
		ID           uint64             `gorm:"column:id"`
		ProductCode  string             `gorm:"column:product_code"`
		Name         string             `gorm:"column:name"`
		Amount       float64            `gorm:"column:amount"`
		Unit         model.UnitType     `gorm:"column:unit"`
		Price        float64            `gorm:"column:price"`
		AllergenType model.AllergenType `gorm:"column:allergen_type"`
		ImageID      uint64             `gorm:"column:img_id"`
		ImageOrder   int                `gorm:"column:img_order"`
		ImageURL     string             `gorm:"column:img_url"`
	}

	result := is.repoFactory.Product().GetDBWithContext(ctx).
		Table("products p").
		// 使用 DISTINCT ON 获取每个 product_code 的第一条图片（按 order ASC）
		Joins(`
			LEFT JOIN (
				SELECT DISTINCT ON (product_code) 
					product_code, 
					id, 
					sort_order, 
					image_url
				FROM product_images
				ORDER BY product_code, sort_order ASC
			) pi ON pi.product_code = p.product_code
		`).
		Select(`
			p.id,
			p.product_code,
			p.name,
			p.amount,
			p.unit,
			p.price,
			p.allergen_type,
			pi.id AS img_id,
			pi.sort_order AS img_order,
			pi.image_url AS img_url
		`).
		Where("p.ingredient_code = ?", code).
		Where("p.id > ?", cursor).
		Order("p.id ASC").
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

	// 直接转换，无需分组
	products := make([]*dto.ViewProductCard, 0, len(results))
	for _, r := range results {
		product := &dto.ViewProductCard{
			ID:           r.ID,
			ProductCode:  r.ProductCode,
			Name:         r.Name,
			Amount:       r.Amount,
			Unit:         r.Unit,
			Price:        r.Price,
			AllergenType: r.AllergenType,
			Image:        dto.ImageResponse{},
		}
		// 仅当存在封面图时添加
		if r.ImageID > 0 {
			product.Image = dto.ImageResponse{
				ID: r.ImageID,
				// 排序,用于显示顺序
				SortOrder: r.ImageOrder,
				ImageURL:  r.ImageURL,
			}
		}
		products = append(products, product)
	}

	return &dto.ViewProductCardListWithCursor{
		Products: products,
		Cursor:   newCursor,
		HasMore:  hasMore,
	}, nil
}

/*
慢SQL,平均查询时间300ms,可能的性能瓶颈,需要测试
1.使用LATERAL:
SELECT

	i.id,
	i.ingredient_code,
	i.name,
	i.description,
	ii.id AS img_id,
	ii.sort_order AS img_order,
	ii.image_url AS img_url

FROM ingredients i
LEFT JOIN LATERAL (

	SELECT id, sort_order, image_url
	FROM ingredient_images
	WHERE ingredient_code = i.ingredient_code
	ORDER BY sort_order ASC
	LIMIT 1

) ii ON true
WHERE i.id > 0
ORDER BY i.id ASC
LIMIT 16;

2.索引覆盖:
CREATE INDEX idx_images_code_sort ON ingredient_images (ingredient_code, sort_order)
INCLUDE (id, image_url);
*/
func (is *ingredientService) FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.ViewIngredientCardListWithCursor, error) {
	var results []struct {
		ID             uint64 `gorm:"column:id"`
		IngredientCode string `gorm:"column:ingredient_code"`
		Name           string `gorm:"column:name"`
		Description    string `gorm:"column:description"`
		ImageID        uint64 `gorm:"column:img_id"`
		ImageOrder     int    `gorm:"column:img_order"`
		ImageURL       string `gorm:"column:img_url"`
	}
	result := is.repoFactory.Ingredient().GetDBWithContext(ctx).
		Table("ingredients i").
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
			i.id,
			i.ingredient_code,
			i.name,
			i.description,
			ii.id AS img_id,
			ii.sort_order AS img_order,
			ii.image_url AS img_url
		`).
		Where("i.id > ?", cursor).
		Order("i.id ASC").
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

	ingredientCards := make([]*dto.ViewIngredientCard, 0, len(results))
	for _, r := range results {
		ingredient := &dto.ViewIngredientCard{
			IngredientCode: r.IngredientCode,
			Name:           r.Name,
			Description:    r.Description,
		}
		if r.ImageID > 0 {
			ingredient.Image = dto.ImageResponse{
				ID:        r.ImageID,
				SortOrder: r.ImageOrder,
				ImageURL:  r.ImageURL,
			}
		}
		ingredientCards = append(ingredientCards, ingredient)
	}

	return &dto.ViewIngredientCardListWithCursor{
		Ingredients: ingredientCards,
		Cursor:      newCursor,
		HasMore:     hasMore,
	}, nil
}
