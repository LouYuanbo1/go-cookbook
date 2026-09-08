package dto

type dishIngredientBase struct {
	IngredientCode string `form:"ingredientCode" json:"ingredientCode"`
	Quantity       string `form:"quantity" json:"quantity"`
	Note           string `form:"note" json:"note"`
}

type CreateDishIngredientReq struct {
	dishIngredientBase
}

type UpdateDishIngredientReq struct {
	//"existing" | "new" | "deleted"
	Type string `form:"type" json:"type"`
	dishIngredientBase
}

type DishIngredientResp struct {
	dishIngredientBase
}

type DishIngredientCardResp struct {
	ID uint64 `json:"id"`
	dishIngredientBase
	Name  string    `json:"name"`
	Image ImageResp `json:"image"`
}

type ExcelDishIngredient struct {
	IngredientCode string `gorm:"column:ingredient_code"`
	Name           string `gorm:"column:name"`
	Quantity       string `gorm:"column:quantity"`
	Note           string `gorm:"column:note"`
}
