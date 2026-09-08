package dto

type CursorReq[ID comparable] struct {
	Cursor ID  `form:"cursor"`
	Limit  int `form:"limit" binding:"min=1,max=100" default:"10"`
}

type CursorResp[T any, ID comparable] struct {
	Items   []T  `json:"items"`
	Cursor  ID   `json:"cursor"`
	HasMore bool `json:"has_more"`
}
