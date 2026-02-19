package model

import (
	"time"
)

type Ingredient struct {
	ID uint64 `gorm:"primaryKey"`
	// 食材代码,例如"面条","猪肉"等,是一个大的品类,因为不同品牌的商品可以是同一种食材
	IngredientCode string `gorm:"column:ingredient_code;type:varchar(255);not null;unique"`
	// 食材名称,例如"面条","猪肉"等
	Name        string    `gorm:"column:name;type:varchar(255);not null"`
	Description string    `gorm:"column:description;type:text"`
	CreatedAt   time.Time `gorm:"not null;default:current_timestamp"`
	UpdatedAt   time.Time `gorm:"not null;default:current_timestamp"`
	//DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (i *Ingredient) GetID() uint64 {
	return i.ID
}

func (i *Ingredient) PrimaryKey() string {
	return "id"
}

func (i *Ingredient) TableName() string {
	return "ingredients"
}
