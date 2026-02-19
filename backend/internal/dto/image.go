package dto

import "mime/multipart"

/*
e.g.

	{
	  "dishCode": "D001",
	  "name": "宫保鸡丁",
	  "description": "经典川菜",
	  "images": [
	    {"type": "existing", "id": 10, "order": 0},
	    {"type": "new", "id": 0, "order": 1},      // 对应 newImages[0]
	    {"type": "existing", "id": 15, "order": 2},
	    {"type": "deleted", "id": 20, "order": -1} // 要删除的图片
	  ]
	}
*/

type NewImageFile struct {
	TempID string                `form:"tempID" json:"tempID"`
	File   *multipart.FileHeader `form:"file" json:"file"`
}

type ImageRequest struct {
	Type string `form:"type" json:"type"` // "existing" | "new" | "deleted"
	TempID    string `form:"tempID" json:"tempID"`       // 新图片的临时ID（type="new"时）
	ID        uint64 `form:"id" json:"id"`               // 已存在图片的ID（type="existing"时）
	SortOrder int    `form:"sortOrder" json:"sortOrder"` // 排序位置（从0开始）
}

type ImageResponse struct {
	ID uint64 `json:"id"`
	// 排序,用于显示顺序
	SortOrder int    `json:"sortOrder"`
	ImageURL  string `json:"imageURL"`
}
