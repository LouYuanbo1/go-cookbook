package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCursorReq_Defaults(t *testing.T) {
	req := CursorReq[uint64]{}
	assert.Equal(t, uint64(0), req.Cursor)
	assert.Equal(t, 0, req.Limit)
}

func TestCursorReq_WithValues(t *testing.T) {
	req := CursorReq[uint64]{Cursor: 100, Limit: 20}
	assert.Equal(t, uint64(100), req.Cursor)
	assert.Equal(t, 20, req.Limit)
}

func TestCursorReq_StringCursor(t *testing.T) {
	req := CursorReq[string]{Cursor: "next_page_token", Limit: 10}
	assert.Equal(t, "next_page_token", req.Cursor)
	assert.Equal(t, 10, req.Limit)
}

func TestCursorResp_Empty(t *testing.T) {
	resp := CursorResp[string, uint64]{
		Items:   []string{},
		Cursor:  0,
		HasMore: false,
	}
	assert.Empty(t, resp.Items)
	assert.Equal(t, uint64(0), resp.Cursor)
	assert.False(t, resp.HasMore)
}

func TestCursorResp_WithData(t *testing.T) {
	items := []int{1, 2, 3}
	resp := CursorResp[int, uint64]{
		Items:   items,
		Cursor:  3,
		HasMore: true,
	}
	assert.Equal(t, items, resp.Items)
	assert.Equal(t, uint64(3), resp.Cursor)
	assert.True(t, resp.HasMore)
	assert.Len(t, resp.Items, 3)
}

func TestCursorResp_StructItems(t *testing.T) {
	type Item struct {
		ID   uint64
		Name string
	}
	items := []Item{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	resp := CursorResp[Item, uint64]{
		Items:   items,
		Cursor:  2,
		HasMore: false,
	}
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, "a", resp.Items[0].Name)
	assert.Equal(t, uint64(2), resp.Items[1].ID)
	assert.False(t, resp.HasMore)
}

func TestCursorResp_StringCursor(t *testing.T) {
	items := []string{"item1"}
	resp := CursorResp[string, string]{
		Items:   items,
		Cursor:  "cursor_abc",
		HasMore: false,
	}
	assert.Equal(t, "cursor_abc", resp.Cursor)
	assert.Equal(t, "item1", resp.Items[0])
}
