package dto

import "mime/multipart"

type dishBase struct {
	DishCode    string `form:"dishCode" json:"dishCode"`
	Name        string `form:"name" json:"name"`
	Description string `form:"description" json:"description"`
	Recipe      string `form:"recipe" json:"recipe"`
}

type CreateDishReq struct {
	dishBase
	Ingredients []CreateDishIngredientReq `form:"ingredients"`
	Images      []*multipart.FileHeader   `form:"images"`
}

type UpdateDishReq struct {
	dishBase
	Ingredients []UpdateDishIngredientReq `form:"ingredients"`
	//测试新添加,扁平化
	NewImages []*multipart.FileHeader `form:"newImages"`
	// 新图片的排序顺序
	NewImagesOrders []int `form:"newImageOrders"`
	// 更新的图片ID和排序顺序列表
	UpdatedImages []UpdateImageReq `form:"updatedImages"`
	//删除的图片ID列表
	DeletedImageIDs []uint64 `form:"deletedImageIDs"`
}

type DishResp struct {
	dishBase
	Images []ImageResp `json:"images"`
}

type DishCardResp struct {
	ID       uint64    `json:"id"`
	DishCode string    `form:"dishCode" json:"dishCode"`
	Name     string    `form:"name" json:"name"`
	Image    ImageResp `json:"image"`
}
