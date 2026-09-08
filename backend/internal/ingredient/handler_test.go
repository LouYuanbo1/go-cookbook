package ingredient

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

// mockIngredientService implements IngredientService for testing
type mockIngredientService struct {
	mock.Mock
}

func (m *mockIngredientService) Create(ctx context.Context, req *dto.CreateIngredientReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockIngredientService) FirstByCode(ctx context.Context, code string) (*dto.IngredientResp, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.IngredientResp), args.Error(1)
}

func (m *mockIngredientService) FindByCursor(ctx context.Context, req *dto.CursorReq[uint64]) (*dto.CursorResp[dto.IngredientCardResp, uint64], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CursorResp[dto.IngredientCardResp, uint64]), args.Error(1)
}

func (m *mockIngredientService) Update(ctx context.Context, req *dto.UpdateIngredientReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockIngredientService) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *mockIngredientService) FindProductsByCode(ctx context.Context, ingredientCode string, cursor uint64, limit int) (*dto.CursorResp[dto.ProductCardResp, uint64], error) {
	args := m.Called(ctx, ingredientCode, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CursorResp[dto.ProductCardResp, uint64]), args.Error(1)
}

func (m *mockIngredientService) Import(ctx context.Context, file *multipart.FileHeader, batchSize int) error {
	args := m.Called(ctx, file, batchSize)
	return args.Error(0)
}

func (m *mockIngredientService) Export(gctx *gin.Context, batchSize int) error {
	args := m.Called(gctx, batchSize)
	return args.Error(0)
}

// ingredientMockJWTUtil for testing
type ingredientMockJWTUtil struct {
	mock.Mock
}

func (m *ingredientMockJWTUtil) GenerateToken(role string) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *ingredientMockJWTUtil) ParseToken(tokenString string) (*jwt.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.CustomClaims), args.Error(1)
}

func setupIngredientTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func performIngredientRequest(router *gin.Engine, method, path string, body interface{}, token ...string) *httptest.ResponseRecorder {
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

func TestNewIngredientHandler(t *testing.T) {
	mockSvc := new(mockIngredientService)
	handler := NewIngredientHandler(mockSvc)
	assert.NotNil(t, handler)
}

func TestCreateIngredient_Success(t *testing.T) {
	mockSvc := new(mockIngredientService)
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*dto.CreateIngredientReq")).Return(nil)

	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performIngredientRequest(router, "POST", "/api/ingredients", dto.CreateIngredientReq{
		IngredientCode: "ING001",
		Name:           "Test Ingredient",
	}, "test-token")
	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	assert.Equal(t, "success", resp["msg"])
	mockSvc.AssertExpectations(t)
}

func TestCreateIngredient_ServiceError(t *testing.T) {
	mockSvc := new(mockIngredientService)
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*dto.CreateIngredientReq")).Return(errors.New("service error"))

	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performIngredientRequest(router, "POST", "/api/ingredients", dto.CreateIngredientReq{
		IngredientCode: "ING001",
		Name:           "Test Ingredient",
	}, "test-token")
	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCreateIngredient_InvalidRequest(t *testing.T) {
	mockSvc := new(mockIngredientService)
	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	// Send invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/ingredients", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetIngredientByCode_Success(t *testing.T) {
	mockSvc := new(mockIngredientService)
	mockSvc.On("FirstByCode", mock.Anything, "ING001").Return(&dto.IngredientResp{
		IngredientCode: "ING001",
		Name:           "Test Ingredient",
	}, nil)

	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performIngredientRequest(router, "GET", "/api/ingredients/ING001", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "ING001", data["ingredientCode"])
	assert.Equal(t, "Test Ingredient", data["name"])
	mockSvc.AssertExpectations(t)
}

func TestGetIngredientByCode_NotFound(t *testing.T) {
	mockSvc := new(mockIngredientService)
	mockSvc.On("FirstByCode", mock.Anything, "NONEXISTENT").Return(nil, errors.New("not found"))

	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performIngredientRequest(router, "GET", "/api/ingredients/NONEXISTENT", nil)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFindIngredientsByCursor_Success(t *testing.T) {
	mockSvc := new(mockIngredientService)
	req := &dto.CursorReq[uint64]{Cursor: 0, Limit: 10}
	mockSvc.On("FindByCursor", mock.Anything, req).Return(&dto.CursorResp[dto.IngredientCardResp, uint64]{
		Items:   []dto.IngredientCardResp{{ID: 1, IngredientCode: "ING001", Name: "Flour"}},
		Cursor:  1,
		HasMore: false,
	}, nil)

	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performIngredientRequest(router, "GET", "/api/ingredients?cursor=0&limit=10", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestFindIngredientsByCursor_InvalidQuery(t *testing.T) {
	mockSvc := new(mockIngredientService)
	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	// Invalid limit (0)
	w := performIngredientRequest(router, "GET", "/api/ingredients?cursor=0&limit=0", nil)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateIngredient_Success(t *testing.T) {
	mockSvc := new(mockIngredientService)
	mockSvc.On("Update", mock.Anything, mock.AnythingOfType("*dto.UpdateIngredientReq")).Return(nil)

	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	// UpdateIngredient requires multipart/form-data
	w := httptest.NewRecorder()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Updated Ingredient")
	writer.Close()
	req, _ := http.NewRequest("PATCH", "/api/ingredients/ING001", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-token")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestUpdateIngredient_InvalidRequest(t *testing.T) {
	mockSvc := new(mockIngredientService)
	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	// Send invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/ingredients/ING001", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteIngredient_Success(t *testing.T) {
	mockSvc := new(mockIngredientService)
	mockSvc.On("Delete", mock.Anything, "ING001").Return(nil)

	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performIngredientRequest(router, "DELETE", "/api/ingredients/ING001", nil, "test-token")
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestDeleteIngredient_Error(t *testing.T) {
	mockSvc := new(mockIngredientService)
	mockSvc.On("Delete", mock.Anything, "ING001").Return(errors.New("delete error"))

	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	w := performIngredientRequest(router, "DELETE", "/api/ingredients/ING001", nil, "test-token")
	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestFindProductsByIngredientCode_Success(t *testing.T) {
	mockSvc := new(mockIngredientService)
	mockSvc.On("FindProductsByCode", mock.Anything, "ING001", uint64(0), 10).Return(&dto.CursorResp[dto.ProductCardResp, uint64]{
		Items: []dto.ProductCardResp{{
			ID: 1, ProductCode: "P001", Name: "Product 1",
		}},
		Cursor:  1,
		HasMore: false,
	}, nil)

	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	w := performIngredientRequest(router, "GET", "/api/ingredients/ING001/products?cursor=0&limit=10", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])
	mockSvc.AssertExpectations(t)
}

func TestIngredientHandler_RegisterRoutes(t *testing.T) {
	mockSvc := new(mockIngredientService)
	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	handler.RegisterRoutes(router, mockJWT)

	// Verify routes are registered
	routes := router.Routes()
	assert.NotEmpty(t, routes)

	// Check that all expected routes are present
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+" "+route.Path] = true
	}

	assert.True(t, routePaths["POST /api/ingredients"])
	assert.True(t, routePaths["GET /api/ingredients"])
	assert.True(t, routePaths["GET /api/ingredients/:ingredientCode"])
	assert.True(t, routePaths["GET /api/ingredients/:ingredientCode/products"])
	assert.True(t, routePaths["PATCH /api/ingredients/:ingredientCode"])
	assert.True(t, routePaths["DELETE /api/ingredients/:ingredientCode"])
	assert.True(t, routePaths["POST /api/ingredients/import"])
	assert.True(t, routePaths["GET /api/ingredients/export"])
}

func TestCreateIngredient_ConcurrentRequests(t *testing.T) {
	mockSvc := new(mockIngredientService)
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*dto.CreateIngredientReq")).Return(nil)

	handler := NewIngredientHandler(mockSvc)
	router := setupIngredientTestRouter()
	mockJWT := new(ingredientMockJWTUtil)
	mockJWT.On("ParseToken", mock.Anything).Return(&jwt.CustomClaims{Role: "admin"}, nil)
	handler.RegisterRoutes(router, mockJWT)

	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			w := performIngredientRequest(router, "POST", "/api/ingredients", dto.CreateIngredientReq{
				IngredientCode: "ING001",
				Name:           "Test",
			}, "test-token")
			assert.Equal(t, http.StatusCreated, w.Code)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
	mockSvc.AssertExpectations(t)
}
