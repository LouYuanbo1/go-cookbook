package dto

import (
	"go-cookbook/internal/model"
	"mime/multipart"
)

type CreateProductRequest struct {
	ProductCode    string                  `form:"productCode" json:"productCode"`
	IngredientCode string                  `form:"ingredientCode" json:"ingredientCode"`
	Name           string                  `form:"name" json:"name"`
	Amount         float64                 `form:"amount" json:"amount"`
	Unit           model.UnitType          `form:"unit" json:"unit"`
	Description    string                  `form:"description" json:"description"`
	Price          float64                 `form:"price" json:"price"`
	AllergenType   model.AllergenType      `form:"allergenType" json:"allergenType"`
	Images         []*multipart.FileHeader `form:"images" json:"images"`
}

type UpdateProductRequest struct {
	ProductCode    string             `form:"productCode" json:"productCode"`
	IngredientCode string             `form:"ingredientCode" json:"ingredientCode"`
	Name           string             `form:"name" json:"name"`
	Amount         float64            `form:"amount" json:"amount"`
	Unit           model.UnitType     `form:"unit" json:"unit"`
	Description    string             `form:"description" json:"description"`
	Price          float64            `form:"price" json:"price"`
	AllergenType   model.AllergenType `form:"allergenType" json:"allergenType"`

	// 新上传的图片文件（通过multipart/form-data）
	/*
	NewImageTempIDs []string `form:"newImageTempIDs" json:"newImageTempIDs"`

	NewImages []*multipart.FileHeader `form:"newImages" json:"newImages"`
	*/
	NewImages []NewImageFile `form:"newImages" json:"newImages"`

	// 最终的图片配置（按顺序排列）
	Images []ImageRequest `form:"images" json:"images"`
}

type ViewProductResponse struct {
	ProductCode    string             `json:"productCode"`
	IngredientCode string             `json:"ingredientCode"`
	Name           string             `json:"name"`
	Amount         float64            `json:"amount"`
	Unit           model.UnitType     `json:"unit"`
	Description    string             `json:"description"`
	Price          float64            `json:"price"`
	AllergenType   model.AllergenType `json:"allergenType"`
	Images         []ImageResponse    `json:"images"`
}

type ViewProductCard struct {
	ID           uint64             `json:"id"`
	ProductCode  string             `json:"productCode"`
	Name         string             `json:"name"`
	Amount       float64            `json:"amount"`
	Unit         model.UnitType     `json:"unit"`
	Price        float64            `json:"price"`
	AllergenType model.AllergenType `json:"allergenType"`
	Image        ImageResponse      `json:"image"`
}

type ViewProductCardListWithCursor struct {
	Products []*ViewProductCard `json:"products"`
	Cursor   uint64             `json:"cursor"`
	HasMore  bool               `json:"hasMore"`
}
