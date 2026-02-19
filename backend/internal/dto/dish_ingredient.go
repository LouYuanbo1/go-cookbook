package dto

type CreateDishIngredientRequest struct {
	IngredientCode string `form:"ingredientCode" json:"ingredientCode"`
	Quantity       string `form:"quantity" json:"quantity"`
	Note           string `form:"note" json:"note"`
}

type UpdateDishIngredientRequest struct {
	//"existing" | "new" | "deleted"
	Type           string `form:"type" json:"type"`
	IngredientCode string `form:"ingredientCode" json:"ingredientCode"`
	Quantity       string `form:"quantity" json:"quantity"`
	Note           string `form:"note" json:"note"`
}

type ViewDishIngredientResponse struct {
	IngredientCode string `json:"ingredientCode"`
	Quantity       string `json:"quantity"`
	Note           string `json:"note"`
}

type ViewDishIngredientCard struct {
	ID             uint64        `json:"id"`
	IngredientCode string        `json:"ingredientCode"`
	Name           string        `json:"name"`
	Quantity       string        `json:"quantity"`
	Note           string        `json:"note"`
	Image          ImageResponse `json:"image"`
}

type ViewDishIngredientCardListWithCursor struct {
	DishIngredients []*ViewDishIngredientCard `json:"dishIngredients"`
	Cursor          uint64                    `json:"cursor"`
	HasMore         bool                      `json:"hasMore"`
}

type ExcelDishIngredient struct {
	IngredientCode string `gorm:"column:ingredient_code"`
	Name           string `gorm:"column:name"`
	Quantity       string `gorm:"column:quantity"`
	Note           string `gorm:"column:note"`
}
