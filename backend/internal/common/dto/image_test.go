package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageResp_Defaults(t *testing.T) {
	resp := ImageResp{}
	assert.Equal(t, uint64(0), resp.ID)
	assert.Equal(t, 0, resp.SortOrder)
	assert.Empty(t, resp.ImageURL)
}

func TestImageResp_WithValues(t *testing.T) {
	resp := ImageResp{
		ID:        1,
		SortOrder: 2,
		ImageURL:  "/uploads/test.jpg",
	}
	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, 2, resp.SortOrder)
	assert.Equal(t, "/uploads/test.jpg", resp.ImageURL)
}

func TestUpdateImageReq_Defaults(t *testing.T) {
	req := UpdateImageReq{}
	assert.Equal(t, uint64(0), req.ID)
	assert.Equal(t, 0, req.SortOrder)
}

func TestUpdateImageReq_WithValues(t *testing.T) {
	req := UpdateImageReq{
		ID:        10,
		SortOrder: 3,
	}
	assert.Equal(t, uint64(10), req.ID)
	assert.Equal(t, 3, req.SortOrder)
}

func TestCreateImageReq_Defaults(t *testing.T) {
	req := CreateImageReq{}
	assert.Equal(t, uint64(0), req.ID)
	assert.Equal(t, 0, req.SortOrder)
}

func TestCreateImageReq_WithValues(t *testing.T) {
	req := CreateImageReq{
		ID:        5,
		SortOrder: 1,
	}
	assert.Equal(t, uint64(5), req.ID)
	assert.Equal(t, 1, req.SortOrder)
}
