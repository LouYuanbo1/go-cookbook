package dto

import (
	"go-cookbook/internal/model"
	"mime/multipart"
)

type productBase struct {
	ProductCode    string             `form:"productCode" json:"productCode"`
	IngredientCode string             `form:"ingredientCode" json:"ingredientCode"`
	Name           string             `form:"name" json:"name"`
	Description    string             `form:"description" json:"description"`
	Amount         float64            `form:"amount" json:"amount"`
	Unit           model.UnitType     `form:"unit" json:"unit"`
	Price          float64            `form:"price" json:"price"`
	AllergenType   model.AllergenType `form:"allergenType" json:"allergenType"`
}

type CreateProductReq struct {
	productBase
	Images []*multipart.FileHeader `form:"images"`
}

type UpdateProductReq struct {
	productBase
	/*
		NewImageTempIDs []string `form:"newImageTempIDs" json:"newImageTempIDs"`
		NewImages []*multipart.FileHeader `form:"newImages" json:"newImages"`
		NewImages []NewImageFile `form:"newImages" json:"newImages"`
		// 最终的图片配置（按顺序排列）
		Images []ImageRequest `form:"images" json:"images"`
	*/
	//测试新添加,扁平化
	NewImages []*multipart.FileHeader `form:"newImages"`
	// 新图片的排序顺序
	NewImagesOrders []int `form:"newImageOrders"`
	// 更新的图片ID和排序顺序列表
	UpdatedImages []UpdateImageReq `form:"updatedImages"`
	//删除的图片ID列表
	DeletedImageIDs []uint64 `form:"deletedImageIDs"`
}

type ProductResp struct {
	productBase
	Images []ImageResp `json:"images"`
}

type ProductCardResp struct {
	ID             uint64             `json:"id"`
	ProductCode    string             `json:"productCode"`
	IngredientCode string             `json:"ingredientCode"`
	Name           string             `json:"name"`
	Amount         float64            `json:"amount"`
	Unit           model.UnitType     `json:"unit"`
	Price          float64            `json:"price"`
	AllergenType   model.AllergenType `json:"allergenType"`
	Image          ImageResp          `json:"image"`
}
