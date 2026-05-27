package model

import (
	"time"
)

type Dish struct {
	ID uint64 `gorm:"primaryKey"`
	// 菜品代码
	DishCode string `gorm:"type:varchar(255);not null;unique"`
	// 菜品名称,例如"红烧肉","鱼香肉丝"等
	Name string `gorm:"type:varchar(255);not null"`
	// 菜品描述
	Description string `gorm:"type:text"`
	// 菜品的具体做法
	Recipe    string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
	UpdatedAt time.Time `gorm:"not null;default:current_timestamp"`
	//DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (d *Dish) GetID() uint64 {
	return d.ID
}

func (d *Dish) PrimaryKey() string {
	return "id"
}

func (d *Dish) TableName() string {
	return "dishes"
}
