package dto

type CursorRequest struct {
	Cursor uint64 `form:"cursor" json:"cursor"`
	Limit  int    `form:"limit" json:"limit"`
}
