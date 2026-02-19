package dto

import "mime/multipart"

type CreateIngredientRequest struct {
	IngredientCode string                  `form:"ingredientCode" json:"ingredientCode"`
	Name           string                  `form:"name" json:"name"`
	Description    string                  `form:"description" json:"description"`
	Images         []*multipart.FileHeader `form:"images" json:"images"`
}

type UpdateIngredientRequest struct {
	IngredientCode string `form:"ingredientCode" json:"ingredientCode"`
	Name           string `form:"name" json:"name"`
	Description    string `form:"description" json:"description"`

	/*
		NewImageTempIDs []string                `form:"newImageTempIDs" json:"newImageTempIDs"`
		NewImages       []*multipart.FileHeader `form:"newImages" json:"newImages"`
	*/

	NewImages []NewImageFile `form:"newImages" json:"newImages"`

	Images []ImageRequest `form:"images" json:"images"`
}

type ViewIngredientResponse struct {
	IngredientCode string          `json:"ingredientCode"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Images         []ImageResponse `json:"images"`
}

type ViewIngredientCard struct {
	IngredientCode string        `json:"ingredientCode"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Image          ImageResponse `json:"image"`
}

type ViewIngredientCardListWithCursor struct {
	Ingredients []*ViewIngredientCard `json:"ingredients"`
	Cursor      uint64                `json:"cursor"`
	HasMore     bool                  `json:"hasMore"`
}
