package dto

import "mime/multipart"

type CreateDishRequest struct {
	DishCode    string `form:"dishCode" json:"dishCode"`
	Name        string `form:"name" json:"name"`
	Description string `form:"description" json:"description"`
	Recipe      string `form:"recipe" json:"recipe"`

	Ingredients []*CreateDishIngredientRequest `form:"ingredients" json:"ingredients"`
	Images      []*multipart.FileHeader        `form:"images" json:"images"`
}

type UpdateDishRequest struct {
	DishCode    string `form:"dishCode" json:"dishCode"`
	Name        string `form:"name" json:"name"`
	Description string `form:"description" json:"description"`
	Recipe      string `form:"recipe" json:"recipe"`

	Ingredients []*UpdateDishIngredientRequest `form:"ingredients" json:"ingredients"`

	/*
		NewImageTempIDs []string                `form:"newImageTempIDs" json:"newImageTempIDs"`
		NewImages       []*multipart.FileHeader `form:"newImages" json:"newImages"`
	*/

	NewImages []NewImageFile `form:"newImages" json:"newImages"`

	Images []ImageRequest `form:"images" json:"images"`
}

type ViewDishResponse struct {
	DishCode    string          `json:"dishCode"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Recipe      string          `json:"recipe"`
	Images      []ImageResponse `json:"images"`
}

type ViewDishCard struct {
	ID       uint64        `json:"id"`
	DishCode string        `json:"dishCode"`
	Name     string        `json:"name"`
	Image    ImageResponse `json:"image"`
}

type ViewDishCardListWithCursor struct {
	Dishes  []*ViewDishCard `json:"dishes"`
	Cursor  uint64          `json:"cursor"`
	HasMore bool            `json:"hasMore"`
}
