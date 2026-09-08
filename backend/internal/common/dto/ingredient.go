package dto

import (
	"fmt"
	"mime/multipart"
)

type ingredientBase struct {
	IngredientCode string `form:"ingredientCode" json:"ingredientCode"`
	Name           string `form:"name" json:"name"`
	Description    string `form:"description" json:"description"`
}

type CreateIngredientReq struct {
	ingredientBase
	Images []*multipart.FileHeader `form:"images"`
}

type UpdateIngredientReq struct {
	ingredientBase
	//测试新添加,扁平化
	NewImages []*multipart.FileHeader `form:"newImages"`
	// 新图片的排序顺序
	NewImagesOrders []int `form:"newImageOrders"`
	// 更新的图片ID和排序顺序列表
	UpdatedImages []UpdateImageReq `form:"updatedImages"`
	//删除的图片ID列表
	DeletedImageIDs []uint64 `form:"deletedImageIDs"`
}

func (uir *UpdateIngredientReq) Validate() error {
	if len(uir.NewImagesOrders) != len(uir.NewImages) {
		return fmt.Errorf("newImagesOrders and newImages must have the same length")
	}
	return nil
}

type IngredientResp struct {
	ingredientBase
	Images []ImageResp `json:"images"`
}

type IngredientCardResp struct {
	ID             uint64    `json:"id"`
	IngredientCode string    `json:"ingredientCode"`
	Name           string    `json:"name"`
	Image          ImageResp `json:"image"`
}
