package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageReq_Defaults(t *testing.T) {
	req := PageReq{}
	assert.Equal(t, 0, req.Page)
	assert.Equal(t, 0, req.PageSize)
}

func TestPageReq_WithValues(t *testing.T) {
	req := PageReq{Page: 2, PageSize: 20}
	assert.Equal(t, 2, req.Page)
	assert.Equal(t, 20, req.PageSize)
}

func TestPageResp_Empty(t *testing.T) {
	resp := PageResp[string]{
		Items: []string{},
		Page:  1,
		Size:  10,
	}
	assert.Empty(t, resp.Items)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.Size)
}

func TestPageResp_WithData(t *testing.T) {
	items := []int{10, 20, 30}
	resp := PageResp[int]{
		Items: items,
		Page:  1,
		Size:  3,
	}
	assert.Equal(t, items, resp.Items)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 3, resp.Size)
	assert.Len(t, resp.Items, 3)
}

func TestPageResp_StructItems(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}
	items := []Item{{ID: 1, Name: "test"}}
	resp := PageResp[Item]{
		Items: items,
		Page:  1,
		Size:  10,
	}
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "test", resp.Items[0].Name)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.Size)
}
