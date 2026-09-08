package dto

import (
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDishReq_Defaults(t *testing.T) {
	req := CreateDishReq{}
	assert.Empty(t, req.DishCode)
	assert.Empty(t, req.Name)
	assert.Empty(t, req.Description)
	assert.Empty(t, req.Recipe)
	assert.Empty(t, req.Ingredients)
	assert.Empty(t, req.Images)
}

func TestDishResp_Defaults(t *testing.T) {
	resp := DishResp{}
	assert.Empty(t, resp.DishCode)
	assert.Empty(t, resp.Name)
	assert.Empty(t, resp.Description)
	assert.Empty(t, resp.Recipe)
	assert.Empty(t, resp.Images)
}

func TestDishCardResp_Defaults(t *testing.T) {
	resp := DishCardResp{}
	assert.Equal(t, uint64(0), resp.ID)
	assert.Empty(t, resp.DishCode)
	assert.Empty(t, resp.Name)
	assert.Equal(t, ImageResp{}, resp.Image)
}

func TestCreateDishReq_WithIngredients(t *testing.T) {
	req := CreateDishReq{
		dishBase: dishBase{
			DishCode:    "D001",
			Name:        "Test Dish",
			Description: "A test dish",
			Recipe:      "Step 1, Step 2",
		},
		Ingredients: []CreateDishIngredientReq{
			{IngredientCode: "ING001", Quantity: "100g", Note: "test note"},
		},
	}
	assert.Equal(t, "D001", req.DishCode)
	assert.Equal(t, "Test Dish", req.Name)
	assert.Len(t, req.Ingredients, 1)
	assert.Equal(t, "ING001", req.Ingredients[0].IngredientCode)
}

func TestUpdateDishReq_WithImages(t *testing.T) {
	req := UpdateDishReq{
		dishBase: dishBase{
			DishCode: "D001",
			Name:     "Updated Dish",
		},
		NewImages:       []*multipart.FileHeader{{Filename: "img1.jpg"}},
		NewImagesOrders: []int{0},
		UpdatedImages:   []UpdateImageReq{{ID: 1, SortOrder: 0}},
		DeletedImageIDs: []uint64{2, 3},
	}
	assert.Equal(t, "D001", req.DishCode)
	assert.Len(t, req.NewImages, 1)
	assert.Len(t, req.UpdatedImages, 1)
	assert.Len(t, req.DeletedImageIDs, 2)
	assert.Equal(t, uint64(1), req.UpdatedImages[0].ID)
}
