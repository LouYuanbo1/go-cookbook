package model

import (
	"time"
)

type DishIngredient struct {
	ID uint64 `gorm:"primaryKey"`
	// 菜品代码
	DishCode string `gorm:"uniqueIndex:idx_dish_ingredients_dish_ingredient_code"`
	// 食材代码,一个较大的分类,例如"面条",可以是康师傅的面条,也可以是海底捞的面条,这些面条都是菜品的可选项
	IngredientCode string `gorm:"uniqueIndex:idx_dish_ingredients_dish_ingredient_code"`
	// 数量,例如"100g","200ml","3个","适量",等
	Quantity string `gorm:"type:varchar(100)"`
	// 备注,可以描述食材选哪种商品,例如"选康师傅红烧牛肉面更好"
	Note      string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
	UpdatedAt time.Time `gorm:"not null;default:current_timestamp"`
	//DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (di *DishIngredient) GetID() uint64 {
	return di.ID
}

func (di *DishIngredient) PrimaryKey() string {
	return "id"
}

func (di *DishIngredient) TableName() string {
	return "dish_ingredients"
}
