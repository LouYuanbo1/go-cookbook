package productController

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

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) Create(ctx context.Context, req *dto.CreateProductRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockProductService) GetByCode(ctx context.Context, code string) (*dto.ViewProductResponse, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ViewProductResponse), args.Error(1)
}

func (m *MockProductService) FindDishesByProductCodeAndCursor(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewDishCardListWithCursor, error) {
	args := m.Called(ctx, code, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ViewDishCardListWithCursor), args.Error(1)
}

func (m *MockProductService) Update(ctx context.Context, req *dto.UpdateProductRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockProductService) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *MockProductService) Import(ctx context.Context, file *multipart.FileHeader, batchSize int) error {
	args := m.Called(ctx, file, batchSize)
	return args.Error(0)
}

func (m *MockProductService) Export(gctx *gin.Context, batchSize int) error {
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

func TestProductController_CreateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockProductService)
	mockImgUtil := new(MockImgUtil)
	controller := NewProductController(mockService, mockImgUtil)

	router.POST("/api/products", controller.CreateProduct)

	mockService.On("Create", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/products", strings.NewReader("productCode=PROD001&name=TestProduct"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
}

func TestProductController_GetProductByCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockProductService)
	mockImgUtil := new(MockImgUtil)
	controller := NewProductController(mockService, mockImgUtil)

	router.GET("/api/products/:productCode", controller.GetProductByCode)

	expectedResponse := &dto.ViewProductResponse{
		ProductCode:    "PROD001",
		IngredientCode: "ING001",
		Name:           "Test Product",
	}
	mockService.On("GetByCode", mock.Anything, "PROD001").Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/products/PROD001", nil)
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

func TestProductController_FindDishesByProductCodeAndCursor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockProductService)
	mockImgUtil := new(MockImgUtil)
	controller := NewProductController(mockService, mockImgUtil)

	router.GET("/api/products/:productCode/dishes", controller.FindDishesByProductCodeAndCursor)

	expectedResponse := &dto.ViewDishCardListWithCursor{
		Dishes: []*dto.ViewDishCard{
			{DishCode: "DISH001", Name: "Dish 1"},
			{DishCode: "DISH002", Name: "Dish 2"},
		},
		Cursor:  2,
		HasMore: false,
	}
	mockService.On("FindDishesByProductCodeAndCursor", mock.Anything, "PROD001", uint64(0), 10).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/products/PROD001/dishes?cursor=0&limit=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
	data := response["data"].(map[string]interface{})
	dishes := data["dishes"].([]interface{})
	assert.Len(t, dishes, 2)
}

func TestProductController_UpdateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockProductService)
	mockImgUtil := new(MockImgUtil)
	controller := NewProductController(mockService, mockImgUtil)

	router.PATCH("/api/products/:productCode", controller.UpdateProduct)

	mockService.On("Update", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	body := &strings.Builder{}
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "UpdatedName")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPatch, "/api/products/PROD001", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
}

func TestProductController_DeleteProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockProductService)
	mockImgUtil := new(MockImgUtil)
	controller := NewProductController(mockService, mockImgUtil)

	router.DELETE("/api/products/:productCode", controller.DeleteProduct)

	mockService.On("Delete", mock.Anything, "PROD001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/products/PROD001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
}
