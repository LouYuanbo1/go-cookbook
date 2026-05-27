package model

import (
	"time"
)

type ProductImage struct {
	ID          uint64 `gorm:"primaryKey"`
	ProductCode string `gorm:"not null"`
	// 排序,用于显示顺序
	SortOrder int       `gorm:"not null;default:0"`
	ImageURL  string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
	UpdatedAt time.Time `gorm:"not null;default:current_timestamp"`
	//DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (pi *ProductImage) GetID() uint64 {
	return pi.ID
}

func (pi *ProductImage) PrimaryKey() string {
	return "id"
}

func (pi *ProductImage) TableName() string {
	return "product_images"
}
