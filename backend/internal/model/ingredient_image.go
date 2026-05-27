package model

import (
	"time"
)

type IngredientImage struct {
	ID uint64 `gorm:"primaryKey"`
	// 食材代码
	IngredientCode string `gorm:"not null"`
	// 排序,用于显示顺序
	SortOrder int `gorm:"not null;default:0"`
	// 图片URL
	ImageURL  string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
	UpdatedAt time.Time `gorm:"not null;default:current_timestamp"`
	//DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (ii *IngredientImage) GetID() uint64 {
	return ii.ID
}

func (ii *IngredientImage) PrimaryKey() string {
	return "id"
}

func (ii *IngredientImage) TableName() string {
	return "ingredient_images"
}
