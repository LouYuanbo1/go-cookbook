package ingredientController

import (
	"context"
	"encoding/json"
	"image"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go-cookbook/internal/dto"
	"go-cookbook/internal/utils/imgutil"
)

type MockIngredientService struct {
	mock.Mock
}

func (m *MockIngredientService) Create(ctx context.Context, req *dto.CreateIngredientRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockIngredientService) GetByCode(ctx context.Context, code string) (*dto.ViewIngredientResponse, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ViewIngredientResponse), args.Error(1)
}

func (m *MockIngredientService) FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.ViewIngredientCardListWithCursor, error) {
	args := m.Called(ctx, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ViewIngredientCardListWithCursor), args.Error(1)
}

func (m *MockIngredientService) FindProductsByIngredientCodeAndCursor(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewProductCardListWithCursor, error) {
	args := m.Called(ctx, code, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ViewProductCardListWithCursor), args.Error(1)
}

func (m *MockIngredientService) Update(ctx context.Context, req *dto.UpdateIngredientRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockIngredientService) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *MockIngredientService) Import(ctx context.Context, file *multipart.FileHeader, batchSize int) error {
	args := m.Called(ctx, file, batchSize)
	return args.Error(0)
}

func (m *MockIngredientService) Export(gctx *gin.Context, batchSize int) error {
	args := m.Called(gctx, batchSize)
	return args.Error(0)
}

type MockImgUtil struct {
	mock.Mock
}

func (m *MockImgUtil) Load(imgPath string) (image.Image, error) {
	args := m.Called(imgPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(image.Image), args.Error(1)
}

func (m *MockImgUtil) Thumbnail(img image.Image, opts ...imgutil.TransformOption) image.Image {
	args := m.Called(img)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(image.Image)
}

func (m *MockImgUtil) Save(img image.Image, filename string, opts ...imgutil.SaveOption) error {
	args := m.Called(img, filename)
	return args.Error(0)
}

func (m *MockImgUtil) Delete(imgPath string) error {
	args := m.Called(imgPath)
	return args.Error(0)
}

func (m *MockImgUtil) WithFormatTimestamp(imgPath string, format string) string {
	args := m.Called(imgPath, format)
	return args.String(0)
}

func (m *MockImgUtil) WithUnixNanoTimestamp(imgPath string) string {
	args := m.Called(imgPath)
	return args.String(0)
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	return r
}

func TestIngredientController_CreateIngredient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockIngredientService)
	mockImgUtil := new(MockImgUtil)
	controller := NewIngredientController(mockService, mockImgUtil)

	router.POST("/api/ingredients", controller.CreateIngredient)

	mockService.On("Create", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/ingredients", strings.NewReader("ingredientCode=ING001&name=TestIngredient"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
}

func TestIngredientController_GetIngredientByCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockIngredientService)
	mockImgUtil := new(MockImgUtil)
	controller := NewIngredientController(mockService, mockImgUtil)

	router.GET("/api/ingredients/:ingredientCode", controller.GetIngredientByCode)

	expectedResponse := &dto.ViewIngredientResponse{
		IngredientCode: "ING001",
		Name:           "Test Ingredient",
		Description:    "Test Description",
	}
	mockService.On("GetByCode", mock.Anything, "ING001").Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/ingredients/ING001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
	assert.NotNil(t, response["data"])
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data)
}

func TestIngredientController_FindIngredientsByCursor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockIngredientService)
	mockImgUtil := new(MockImgUtil)
	controller := NewIngredientController(mockService, mockImgUtil)

	router.GET("/api/ingredients", controller.FindIngredientsByCursor)

	expectedResponse := &dto.ViewIngredientCardListWithCursor{
		Ingredients: []*dto.ViewIngredientCard{
			{IngredientCode: "ING001", Name: "Ingredient 1"},
			{IngredientCode: "ING002", Name: "Ingredient 2"},
		},
		Cursor:  2,
		HasMore: false,
	}
	mockService.On("FindByCursor", mock.Anything, uint64(0), 10).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/ingredients?cursor=0&limit=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
	data := response["data"].(map[string]interface{})
	ingredients := data["ingredients"].([]interface{})
	assert.Len(t, ingredients, 2)
}

func TestIngredientController_FindProductsByIngredientCodeAndCursor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockIngredientService)
	mockImgUtil := new(MockImgUtil)
	controller := NewIngredientController(mockService, mockImgUtil)

	router.GET("/api/ingredients/:ingredientCode/products", controller.FindProductsByIngredientCodeAndCursor)

	expectedResponse := &dto.ViewProductCardListWithCursor{
		Products: []*dto.ViewProductCard{
			{ProductCode: "PROD001", Name: "Product 1"},
			{ProductCode: "PROD002", Name: "Product 2"},
		},
		Cursor:  2,
		HasMore: false,
	}
	mockService.On("FindProductsByIngredientCodeAndCursor", mock.Anything, "ING001", uint64(0), 10).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/ingredients/ING001/products?cursor=0&limit=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
	data := response["data"].(map[string]interface{})
	products := data["products"].([]interface{})
	assert.Len(t, products, 2)
}

func TestIngredientController_UpdateIngredient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockIngredientService)
	mockImgUtil := new(MockImgUtil)
	controller := NewIngredientController(mockService, mockImgUtil)

	router.PATCH("/api/ingredients/:ingredientCode", controller.UpdateIngredient)

	mockService.On("Update", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	body := &strings.Builder{}
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "UpdatedName")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPatch, "/api/ingredients/ING001", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
}

func TestIngredientController_DeleteIngredient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockIngredientService)
	mockImgUtil := new(MockImgUtil)
	controller := NewIngredientController(mockService, mockImgUtil)

	router.DELETE("/api/ingredients/:ingredientCode", controller.DeleteIngredient)

	mockService.On("Delete", mock.Anything, "ING001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/ingredients/ING001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
}
