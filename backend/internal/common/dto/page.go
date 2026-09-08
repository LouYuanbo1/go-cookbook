package dto

type PageReq struct {
	Page     int `form:"page" binding:"min=1" default:"1"`
	PageSize int `form:"page_size" binding:"min=1,max=100" default:"10"`
}

// PageResp 通用分页返回结构
type PageResp[T any] struct {
	Items []T `json:"items"`
	Page  int `json:"page"`
	Size  int `json:"size"`
}
