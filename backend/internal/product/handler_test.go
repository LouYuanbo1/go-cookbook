package product

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-cookbook/internal/common/dto"
	"go-cookbook/internal/common/utils/jwt"
	"go-cookbook/internal/model"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockProductService implements ProductService for testing
type mockProductService struct {
	mock.Mock
}

func (m *mockProductService) Create(ctx context.Context, req *dto.CreateProductReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockProductService) FirstByCode(ctx context.Context, code string) (*dto.ProductResp, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResp), args.Error(1)
}

func (m *mockProductService) Update(ctx context.Context, req *dto.UpdateProductReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockProductService) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *mockProductService) FindDishesByCode(ctx context.Context, productCode string, cursor uint64, limit int) (*dto.CursorResp[dto.DishCardResp, uint64], error) {
	args := m.Called(ctx, productCode, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CursorResp[dto.DishCardResp, uint64]), args.Error(1)
}

func (m *mockProductService) Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error {
	args := m.Called(ctx, fileHeader, batchSize)
	return args.Error(0)
}

func (m *mockProductService) Export(gctx *gin.Context, batchSize int) error {
	args := m.Called(gctx, batchSize)
	return args.Error(0)
}

// productMockJWTUtil for testing
type productMockJWTUtil struct {
	mock.Mock
}

func (m *productMockJWTUtil) GenerateToken(role string) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *productMockJWTUtil) ParseToken(tokenString string) (*jwt.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.CustomClaims), args.Error(1)
}

func setupProductTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func performProductRequest(router *gin.Engine, method, path string, body interface{}, token ...string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}
	router.ServeHTTP(w, req)
	return w
}

func TestNewProductHandler(t *testing.T) {
	mockSvc := new(mockProductService)
	handler := NewProductHandler(mockSvc)
	assert.NotNil(t, handler)
}

func TestCreateProduct_Success(t *testing.T) {
	mockSvc := new(mockProductService)
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*dto.CreateProductReq")).Return(nil)

	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performProductRequest(router, "POST", "/api/products", dto.CreateProductReq{
		ProductCode:    "P001",
		IngredientCode: "ING001",
		Name:           "Test Product",
		Amount:         100,
		Unit:           model.UnitGram,
		Price:          10.5,
		AllergenType:   model.AllergenNone,
	}, "test-token")
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	assert.Equal(t, "success", resp["msg"])
	mockSvc.AssertExpectations(t)
}

func TestCreateProduct_InvalidRequest(t *testing.T) {
	mockSvc := new(mockProductService)
	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	// Send invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/products", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateProduct_ServiceError(t *testing.T) {
	mockSvc := new(mockProductService)
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*dto.CreateProductReq")).Return(errors.New("service error"))

	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performProductRequest(router, "POST", "/api/products", dto.CreateProductReq{
		ProductCode: "P001",
		Name:        "Test",
	}, "test-token")
	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestGetProductByCode_Success(t *testing.T) {
	mockSvc := new(mockProductService)
	mockSvc.On("FirstByCode", mock.Anything, "P001").Return(&dto.ProductResp{
		ProductCode:    "P001",
		IngredientCode: "ING001",
		Name:           "Test Product",
		Price:          10.5,
		Unit:           model.UnitGram,
		AllergenType:   model.AllergenNone,
	}, nil)

	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performProductRequest(router, "GET", "/api/products/P001", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "P001", data["productCode"])
	assert.Equal(t, "Test Product", data["name"])
	mockSvc.AssertExpectations(t)
}

func TestGetProductByCode_NotFound(t *testing.T) {
	mockSvc := new(mockProductService)
	mockSvc.On("FirstByCode", mock.Anything, "NONEXISTENT").Return(nil, errors.New("not found"))

	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performProductRequest(router, "GET", "/api/products/NONEXISTENT", nil)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFindDishesByCode_Success(t *testing.T) {
	mockSvc := new(mockProductService)
	mockSvc.On("FindDishesByCode", mock.Anything, "P001", uint64(0), 10).Return(&dto.CursorResp[dto.DishCardResp, uint64]{
		Items: []dto.DishCardResp{{
			ID: 1, DishCode: "D001", Name: "Dish 1",
		}},
		Cursor:  1,
		HasMore: false,
	}, nil)

	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performProductRequest(router, "GET", "/api/products/P001/dishes?cursor=0&limit=10", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestUpdateProduct_Success(t *testing.T) {
	mockSvc := new(mockProductService)
	mockSvc.On("Update", mock.Anything, mock.AnythingOfType("*dto.UpdateProductReq")).Return(nil)

	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performProductRequest(router, "PATCH", "/api/products/P001", dto.UpdateProductReq{
		Name: "Updated Product",
	}, "test-token")
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestUpdateProduct_InvalidRequest(t *testing.T) {
	mockSvc := new(mockProductService)
	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	// Send invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/products/P001", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteProduct_Success(t *testing.T) {
	mockSvc := new(mockProductService)
	mockSvc.On("Delete", mock.Anything, "P001").Return(nil)

	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performProductRequest(router, "DELETE", "/api/products/P001", nil, "test-token")
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestDeleteProduct_Error(t *testing.T) {
	mockSvc := new(mockProductService)
	mockSvc.On("Delete", mock.Anything, "P001").Return(errors.New("delete error"))

	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performProductRequest(router, "DELETE", "/api/products/P001", nil, "test-token")
	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestProductHandler_RegisterRoutes(t *testing.T) {
	mockSvc := new(mockProductService)
	handler := NewProductHandler(mockSvc)
	router := setupProductTestRouter()
	mockJWT := new(productMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	routes := router.Routes()
	assert.NotEmpty(t, routes)

	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+" "+route.Path] = true
	}

	assert.True(t, routePaths["POST /api/products"])
	assert.True(t, routePaths["GET /api/products/:productCode"])
	assert.True(t, routePaths["GET /api/products/:productCode/dishes"])
	assert.True(t, routePaths["PATCH /api/products/:productCode"])
	assert.True(t, routePaths["DELETE /api/products/:productCode"])
	assert.True(t, routePaths["POST /api/products/import"])
	assert.True(t, routePaths["GET /api/products/export"])
}
