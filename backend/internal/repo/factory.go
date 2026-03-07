package repo

import (
	"go-cookbook/internal/model"

	"github.com/LouYuanbo1/go-webservice/gormx/gen"
	"gorm.io/gorm"
)

type RepoFactory interface {
	Product() gen.Session[model.Product, uint64, *model.Product]
	ProductImage() gen.Session[model.ProductImage, uint64, *model.ProductImage]
	Dish() gen.Session[model.Dish, uint64, *model.Dish]
	DishImage() gen.Session[model.DishImage, uint64, *model.DishImage]
	Ingredient() gen.Session[model.Ingredient, uint64, *model.Ingredient]
	IngredientImage() gen.Session[model.IngredientImage, uint64, *model.IngredientImage]
	DishIngredient() gen.Session[model.DishIngredient, uint64, *model.DishIngredient]
	Tx() gen.Tx
}

type repoFactory struct {
	db *gorm.DB
}

func NewRepoFactory(db *gorm.DB) RepoFactory {
	return &repoFactory{db: db}
}

func (f *repoFactory) Product() gen.Session[model.Product, uint64, *model.Product] {
	return gen.NewSession[model.Product, uint64](f.db)
}

func (f *repoFactory) ProductImage() gen.Session[model.ProductImage, uint64, *model.ProductImage] {
	return gen.NewSession[model.ProductImage, uint64](f.db)
}

func (f *repoFactory) Dish() gen.Session[model.Dish, uint64, *model.Dish] {
	return gen.NewSession[model.Dish, uint64](f.db)
}

func (f *repoFactory) DishImage() gen.Session[model.DishImage, uint64, *model.DishImage] {
	return gen.NewSession[model.DishImage, uint64](f.db)
}

func (f *repoFactory) Ingredient() gen.Session[model.Ingredient, uint64, *model.Ingredient] {
	return gen.NewSession[model.Ingredient, uint64](f.db)
}

func (f *repoFactory) IngredientImage() gen.Session[model.IngredientImage, uint64, *model.IngredientImage] {
	return gen.NewSession[model.IngredientImage, uint64](f.db)
}

func (f *repoFactory) DishIngredient() gen.Session[model.DishIngredient, uint64, *model.DishIngredient] {
	return gen.NewSession[model.DishIngredient, uint64](f.db)
}

func (f *repoFactory) Tx() gen.Tx {
	return gen.NewTx(f.db)
}
