package dishController

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

type MockDishService struct {
	mock.Mock
}

func (m *MockDishService) Create(ctx context.Context, req *dto.CreateDishRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockDishService) GetByCode(ctx context.Context, code string) (*dto.ViewDishResponse, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ViewDishResponse), args.Error(1)
}

func (m *MockDishService) FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.ViewDishCardListWithCursor, error) {
	args := m.Called(ctx, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ViewDishCardListWithCursor), args.Error(1)
}

func (m *MockDishService) Update(ctx context.Context, req *dto.UpdateDishRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockDishService) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *MockDishService) FindIngredientsByDishCode(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewDishIngredientCardListWithCursor, error) {
	args := m.Called(ctx, code, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ViewDishIngredientCardListWithCursor), args.Error(1)
}

func (m *MockDishService) Import(ctx context.Context, file *multipart.FileHeader, batchSize int) error {
	args := m.Called(ctx, file, batchSize)
	return args.Error(0)
}

func (m *MockDishService) Export(gctx *gin.Context, batchSize int) error {
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

func TestDishController_CreateDish(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockDishService)
	mockImgUtil := new(MockImgUtil)
	controller := NewDishController(mockService, mockImgUtil)

	router.POST("/api/dishes", controller.CreateDish)

	mockService.On("Create", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	body := &strings.Builder{}
	writer := multipart.NewWriter(body)
	writer.WriteField("dishCode", "DISH001")
	writer.WriteField("name", "TestDish")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/dishes", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
}

func TestDishController_GetDishByCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockDishService)
	mockImgUtil := new(MockImgUtil)
	controller := NewDishController(mockService, mockImgUtil)

	router.GET("/api/dishes/:dishCode", controller.GetDishByCode)

	expectedResponse := &dto.ViewDishResponse{
		DishCode:    "DISH001",
		Name:        "Test Dish",
		Description: "Test Description",
		Recipe:      "Test Recipe",
	}
	mockService.On("GetByCode", mock.Anything, "DISH001").Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/dishes/DISH001", nil)
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

func TestDishController_FindDishesByCursor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockDishService)
	mockImgUtil := new(MockImgUtil)
	controller := NewDishController(mockService, mockImgUtil)

	router.GET("/api/dishes", controller.FindDishesByCursor)

	expectedResponse := &dto.ViewDishCardListWithCursor{
		Dishes: []*dto.ViewDishCard{
			{DishCode: "DISH001", Name: "Dish 1"},
			{DishCode: "DISH002", Name: "Dish 2"},
		},
		Cursor:  2,
		HasMore: false,
	}
	mockService.On("FindByCursor", mock.Anything, uint64(0), 10).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/dishes?cursor=0&limit=10", nil)
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

func TestDishController_UpdateDish(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockDishService)
	mockImgUtil := new(MockImgUtil)
	controller := NewDishController(mockService, mockImgUtil)

	router.PATCH("/api/dishes/:dishCode", controller.UpdateDish)

	mockService.On("Update", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	body := &strings.Builder{}
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "UpdatedName")
	writer.Close()

	req, _ := http.NewRequest(http.MethodPatch, "/api/dishes/DISH001", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
}

func TestDishController_DeleteDish(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	mockService := new(MockDishService)
	mockImgUtil := new(MockImgUtil)
	controller := NewDishController(mockService, mockImgUtil)

	router.DELETE("/api/dishes/:dishCode", controller.DeleteDish)

	mockService.On("Delete", mock.Anything, "DISH001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/dishes/DISH001", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["msg"])
}
