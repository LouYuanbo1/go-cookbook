package model

import (
	"time"
)

type DishImage struct {
	ID uint64 `gorm:"primaryKey"`
	// 菜品代码
	DishCode  string `gorm:"not null"`
	SortOrder int    `gorm:"not null;default:0"`
	// 图片URL
	ImageURL  string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
	UpdatedAt time.Time `gorm:"not null;default:current_timestamp"`
	//DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (di *DishImage) GetID() uint64 {
	return di.ID
}

func (di *DishImage) PrimaryKey() string {
	return "id"
}

func (di *DishImage) TableName() string {
	return "dish_images"
}
