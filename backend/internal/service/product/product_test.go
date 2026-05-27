package productService

import (
	"context"
	"image"
	"testing"
	"time"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/cache/driver/local"
	"github.com/LouYuanbo1/go-webservice/gormc"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/LouYuanbo1/go-webservice/singleflightx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-cookbook/internal/dto"
	"go-cookbook/internal/model"
	"go-cookbook/internal/utils/imgutil"
)

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

func setupTestDB(t *testing.T) *gormc.CacheDB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	err = db.AutoMigrate(
		&model.Product{},
		&model.ProductImage{},
		&model.Ingredient{},
		&model.IngredientImage{},
		&model.Dish{},
		&model.DishImage{},
		&model.DishIngredient{},
	)
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	gormxDB := gormx.NewDB(db)

	localDriver := local.NewDriver(&local.Config{
		CacheSize: 1024 * 1024 * 100,
	}, singleflightx.NewSingleFlight())

	cacheImpl, err := cache.Open(localDriver)
	if err != nil {
		t.Fatalf("failed to open cache: %v", err)
	}

	cacheClient := cache.NewClient(cacheImpl)

	cacheDB := gormc.NewCacheDB(gormxDB, cacheClient, &gormc.Config{
		TTL: time.Minute,
	})

	return cacheDB
}

func TestProductService_Create(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewProductService(cacheDB, mockImgUtil)

	req := &dto.CreateProductRequest{
		ProductCode:    "PROD001",
		IngredientCode: "ING001",
		Name:           "鸡胸肉100g",
		Amount:         100,
		Unit:           model.UnitGram,
		Description:    "新鲜鸡胸肉",
		Price:          10.5,
		AllergenType:   model.AllergenNone,
		Images:         nil,
	}

	err := service.Create(context.Background(), req)
	assert.NoError(t, err)

	var product model.Product
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).First(&product, "product_code = ?", "PROD001").Error
	assert.NoError(t, err)
	assert.Equal(t, "鸡胸肉100g", product.Name)
	assert.Equal(t, float64(100), product.Amount)
	assert.Equal(t, model.UnitGram, product.Unit)
}

func TestProductService_GetByCode(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewProductService(cacheDB, mockImgUtil)

	product := &model.Product{
		ProductCode:    "PROD002",
		IngredientCode: "ING002",
		Name:           "花生米50g",
		Amount:         50,
		Unit:           model.UnitGram,
		Description:    "红皮花生",
		Price:          5.0,
		AllergenType:   model.AllergenNone,
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(product).Error
	assert.NoError(t, err)

	resp, err := service.GetByCode(context.Background(), "PROD002")
	assert.NoError(t, err)
	assert.Equal(t, "PROD002", resp.ProductCode)
	assert.Equal(t, "花生米50g", resp.Name)
	assert.Equal(t, float64(50), resp.Amount)
}

func TestProductService_FindDishesByProductCodeAndCursor(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewProductService(cacheDB, mockImgUtil)

	ingredient := &model.Ingredient{
		IngredientCode: "ING003",
		Name:           "测试食材",
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(ingredient).Error
	assert.NoError(t, err)

	product := &model.Product{
		ProductCode:    "PROD003",
		IngredientCode: "ING003",
		Name:           "测试产品",
		Amount:         100,
		Unit:           model.UnitGram,
		Price:          10.0,
	}
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(product).Error
	assert.NoError(t, err)

	dish := &model.Dish{
		DishCode:    "DISH001",
		Name:        "测试菜品",
		Description: "描述",
		Recipe:      "做法",
	}
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(dish).Error
	assert.NoError(t, err)

	dishIngredient := &model.DishIngredient{
		DishCode:       "DISH001",
		IngredientCode: "ING003",
		Quantity:       "100g",
	}
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(dishIngredient).Error
	assert.NoError(t, err)

	result, err := service.FindDishesByProductCodeAndCursor(context.Background(), "PROD003", 0, 10)
	assert.NoError(t, err)
	assert.Len(t, result.Dishes, 1)
	assert.Equal(t, "DISH001", result.Dishes[0].DishCode)
}

func TestProductService_Update(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewProductService(cacheDB, mockImgUtil)

	product := &model.Product{
		ProductCode:    "PROD004",
		IngredientCode: "ING004",
		Name:           "原名称",
		Amount:         100,
		Unit:           model.UnitGram,
		Description:    "原描述",
		Price:          10.0,
		AllergenType:   model.AllergenNone,
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(product).Error
	assert.NoError(t, err)

	req := &dto.UpdateProductRequest{
		ProductCode:    "PROD004",
		IngredientCode: "ING004",
		Name:           "新名称",
		Amount:         200,
		Unit:           model.UnitGram,
		Description:    "新描述",
		Price:          20.0,
		AllergenType:   model.AllergenNone,
		Images:         []dto.ImageRequest{},
		NewImages:      []dto.NewImageFile{},
	}

	err = service.Update(context.Background(), req)
	assert.NoError(t, err)

	var updatedProduct model.Product
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).First(&updatedProduct, "product_code = ?", "PROD004").Error
	assert.NoError(t, err)
	assert.Equal(t, "新名称", updatedProduct.Name)
	assert.Equal(t, float64(200), updatedProduct.Amount)
	assert.Equal(t, float64(20.0), updatedProduct.Price)
}

func TestProductService_Delete(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewProductService(cacheDB, mockImgUtil)

	product := &model.Product{
		ProductCode:    "PROD005",
		IngredientCode: "ING005",
		Name:           "待删除产品",
		Amount:         100,
		Unit:           model.UnitGram,
		Price:          10.0,
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(product).Error
	assert.NoError(t, err)

	err = service.Delete(context.Background(), "PROD005")
	assert.NoError(t, err)

	var deletedProduct model.Product
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).First(&deletedProduct, "product_code = ?", "PROD005").Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}
