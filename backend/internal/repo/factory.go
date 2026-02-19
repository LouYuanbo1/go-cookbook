package repo

import (
	"go-cookbook/internal/model"

	"github.com/LouYuanbo1/go-webservice/gormx"
	"gorm.io/gorm"
)

type RepoFactory interface {
	Product() gormx.GormX[model.Product, uint64, *model.Product]
	ProductImage() gormx.GormX[model.ProductImage, uint64, *model.ProductImage]
	Dish() gormx.GormX[model.Dish, uint64, *model.Dish]
	DishImage() gormx.GormX[model.DishImage, uint64, *model.DishImage]
	Ingredient() gormx.GormX[model.Ingredient, uint64, *model.Ingredient]
	IngredientImage() gormx.GormX[model.IngredientImage, uint64, *model.IngredientImage]
	DishIngredient() gormx.GormX[model.DishIngredient, uint64, *model.DishIngredient]
	Tx() gormx.GormXTx
}

type repoFactory struct {
	db *gorm.DB
}

func NewRepoFactory(db *gorm.DB) RepoFactory {
	return &repoFactory{db: db}
}

func (f *repoFactory) Product() gormx.GormX[model.Product, uint64, *model.Product] {
	return gormx.NewGormX[model.Product, uint64](f.db)
}

func (f *repoFactory) ProductImage() gormx.GormX[model.ProductImage, uint64, *model.ProductImage] {
	return gormx.NewGormX[model.ProductImage, uint64](f.db)
}

func (f *repoFactory) Dish() gormx.GormX[model.Dish, uint64, *model.Dish] {
	return gormx.NewGormX[model.Dish, uint64](f.db)
}

func (f *repoFactory) DishImage() gormx.GormX[model.DishImage, uint64, *model.DishImage] {
	return gormx.NewGormX[model.DishImage, uint64](f.db)
}

func (f *repoFactory) Ingredient() gormx.GormX[model.Ingredient, uint64, *model.Ingredient] {
	return gormx.NewGormX[model.Ingredient, uint64](f.db)
}

func (f *repoFactory) IngredientImage() gormx.GormX[model.IngredientImage, uint64, *model.IngredientImage] {
	return gormx.NewGormX[model.IngredientImage, uint64](f.db)
}

func (f *repoFactory) DishIngredient() gormx.GormX[model.DishIngredient, uint64, *model.DishIngredient] {
	return gormx.NewGormX[model.DishIngredient, uint64](f.db)
}

func (f *repoFactory) Tx() gormx.GormXTx {
	return gormx.NewGormXTx(f.db)
}
