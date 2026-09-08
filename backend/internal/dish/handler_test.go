package dish

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-cookbook/internal/common/dto"
	"go-cookbook/internal/common/utils/jwt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockDishService implements DishService for testing
type mockDishService struct {
	mock.Mock
}

func (m *mockDishService) Create(ctx context.Context, req *dto.CreateDishReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockDishService) FirstByCode(ctx context.Context, code string) (*dto.DishResp, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DishResp), args.Error(1)
}

func (m *mockDishService) FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.CursorResp[dto.DishCardResp, uint64], error) {
	args := m.Called(ctx, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CursorResp[dto.DishCardResp, uint64]), args.Error(1)
}

func (m *mockDishService) Update(ctx context.Context, req *dto.UpdateDishReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockDishService) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *mockDishService) FindDishIngredientsByCode(ctx context.Context, dishCode string, cursor uint64, limit int) (*dto.CursorResp[dto.DishIngredientCardResp, uint64], error) {
	args := m.Called(ctx, dishCode, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CursorResp[dto.DishIngredientCardResp, uint64]), args.Error(1)
}

func (m *mockDishService) Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error {
	args := m.Called(ctx, fileHeader, batchSize)
	return args.Error(0)
}

func (m *mockDishService) Export(gctx *gin.Context, batchSize int) error {
	args := m.Called(gctx, batchSize)
	return args.Error(0)
}

func setupDishTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func performDishRequest(router *gin.Engine, method, path string, body interface{}, token ...string) *httptest.ResponseRecorder {
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

// mockJWTUtil for testing
type dishMockJWTUtil struct {
	mock.Mock
}

func (m *dishMockJWTUtil) GenerateToken(role string) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *dishMockJWTUtil) ParseToken(tokenString string) (*jwt.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.CustomClaims), args.Error(1)
}

func TestNewDishHandler(t *testing.T) {
	mockSvc := new(mockDishService)
	handler := NewDishHandler(mockSvc)
	assert.NotNil(t, handler)
}

func TestCreateDish_Success(t *testing.T) {
	mockSvc := new(mockDishService)
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*dto.CreateDishReq")).Return(nil)

	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performDishRequest(router, "POST", "/api/dishes", dto.CreateDishReq{
		DishCode: "D001",
		Name:     "Test Dish",
	}, "test-token")
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	assert.Equal(t, "success", resp["msg"])
	mockSvc.AssertExpectations(t)
}

func TestCreateDish_InvalidRequest(t *testing.T) {
	mockSvc := new(mockDishService)
	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	// Send invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/dishes", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(400), resp["code"])
}

func TestCreateDish_ServiceError(t *testing.T) {
	mockSvc := new(mockDishService)
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*dto.CreateDishReq")).Return(errors.New("service error"))

	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performDishRequest(router, "POST", "/api/dishes", dto.CreateDishReq{
		DishCode: "D001",
		Name:     "Test Dish",
	}, "test-token")
	require.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(500), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestFindDishesByCursor_Success(t *testing.T) {
	mockSvc := new(mockDishService)
	mockSvc.On("FindByCursor", mock.Anything, uint64(0), 10).Return(&dto.CursorResp[dto.DishCardResp, uint64]{
		Items:   []dto.DishCardResp{{ID: 1, DishCode: "D001", Name: "Dish 1"}},
		Cursor:  1,
		HasMore: false,
	}, nil)

	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performDishRequest(router, "GET", "/api/dishes?cursor=0&limit=10", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	assert.Equal(t, "success", resp["msg"])

	data := resp["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	assert.Len(t, items, 1)
	mockSvc.AssertExpectations(t)
}

func TestFindDishesByCursor_InvalidQuery(t *testing.T) {
	mockSvc := new(mockDishService)
	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	// Invalid limit (0)
	w := performDishRequest(router, "GET", "/api/dishes?cursor=0&limit=0", nil)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDishByCode_Success(t *testing.T) {
	mockSvc := new(mockDishService)
	mockSvc.On("FirstByCode", mock.Anything, "D001").Return(&dto.DishResp{
		DishCode: "D001",
		Name:     "Test Dish",
		Recipe:   "Recipe",
	}, nil)

	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performDishRequest(router, "GET", "/api/dishes/D001", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "D001", data["dishCode"])
	assert.Equal(t, "Test Dish", data["name"])
	mockSvc.AssertExpectations(t)
}

func TestGetDishByCode_NotFound(t *testing.T) {
	mockSvc := new(mockDishService)
	mockSvc.On("FirstByCode", mock.Anything, "NONEXISTENT").Return(nil, errors.New("not found"))

	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performDishRequest(router, "GET", "/api/dishes/NONEXISTENT", nil)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUpdateDish_Success(t *testing.T) {
	mockSvc := new(mockDishService)
	mockSvc.On("Update", mock.Anything, mock.AnythingOfType("*dto.UpdateDishReq")).Return(nil)

	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performDishRequest(router, "PATCH", "/api/dishes/D001", dto.UpdateDishReq{
		Name: "Updated Dish",
	}, "test-token")
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestDeleteDish_Success(t *testing.T) {
	mockSvc := new(mockDishService)
	mockSvc.On("Delete", mock.Anything, "D001").Return(nil)

	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performDishRequest(router, "DELETE", "/api/dishes/D001", nil, "test-token")
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestDeleteDish_Error(t *testing.T) {
	mockSvc := new(mockDishService)
	mockSvc.On("Delete", mock.Anything, "D001").Return(errors.New("delete error"))

	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performDishRequest(router, "DELETE", "/api/dishes/D001", nil, "test-token")
	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFindDishIngredientsByDishCode_Success(t *testing.T) {
	mockSvc := new(mockDishService)
	mockSvc.On("FindDishIngredientsByCode", mock.Anything, "D001", uint64(0), 10).Return(&dto.CursorResp[dto.DishIngredientCardResp, uint64]{
		Items: []dto.DishIngredientCardResp{{
			ID: 1, IngredientCode: "ING001", Name: "Flour", Quantity: "200g",
		}},
		Cursor:  1,
		HasMore: false,
	}, nil)

	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performDishRequest(router, "GET", "/api/dishes/D001/ingredients?cursor=0&limit=10", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestDishHandler_RegisterRoutes(t *testing.T) {
	mockSvc := new(mockDishService)
	handler := NewDishHandler(mockSvc)
	router := setupDishTestRouter()
	mockJWT := new(dishMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	// Verify routes are registered
	routes := router.Routes()
	assert.NotEmpty(t, routes)

	// Check that all expected routes are present
	routePaths := make(map[string]string)
	for _, route := range routes {
		routePaths[route.Method+" "+route.Path] = route.Handler
	}

	assert.Contains(t, routePaths, "POST /api/dishes")
	assert.Contains(t, routePaths, "GET /api/dishes")
	assert.Contains(t, routePaths, "GET /api/dishes/:dishCode")
	assert.Contains(t, routePaths, "GET /api/dishes/:dishCode/ingredients")
	assert.Contains(t, routePaths, "PATCH /api/dishes/:dishCode")
	assert.Contains(t, routePaths, "DELETE /api/dishes/:dishCode")
	assert.Contains(t, routePaths, "GET /api/dishes/export")
}
