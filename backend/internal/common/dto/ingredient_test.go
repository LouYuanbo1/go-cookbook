package dto

import (
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateIngredientReq_Validate_Success(t *testing.T) {
	tests := []struct {
		name    string
		orders  []int
		images  []*multipart.FileHeader
		wantErr bool
	}{
		{
			name:    "both empty",
			orders:  []int{},
			images:  []*multipart.FileHeader{},
			wantErr: false,
		},
		{
			name:    "nil slices",
			orders:  nil,
			images:  nil,
			wantErr: false,
		},
		{
			name:    "matching length",
			orders:  []int{0, 1, 2},
			images:  make([]*multipart.FileHeader, 3),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := UpdateIngredientReq{
				NewImagesOrders: tt.orders,
				NewImages:       tt.images,
			}
			err := req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateIngredientReq_Validate_Error(t *testing.T) {
	req := UpdateIngredientReq{
		NewImagesOrders: []int{0, 1},
		NewImages:       make([]*multipart.FileHeader, 3),
	}
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have the same length")
}

func TestUpdateIngredientReq_Validate_EmptyImages(t *testing.T) {
	req := UpdateIngredientReq{
		NewImagesOrders: []int{0, 1},
		NewImages:       nil,
	}
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have the same length")
}

func TestUpdateIngredientReq_Validate_EmptyOrders(t *testing.T) {
	req := UpdateIngredientReq{
		NewImagesOrders: nil,
		NewImages:       make([]*multipart.FileHeader, 2),
	}
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have the same length")
}
