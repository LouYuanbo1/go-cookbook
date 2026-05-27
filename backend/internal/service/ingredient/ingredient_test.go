package ingredientService

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
		&model.Ingredient{},
		&model.IngredientImage{},
		&model.Product{},
		&model.ProductImage{},
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

func TestIngredientService_Create(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewIngredientService(cacheDB, mockImgUtil)

	req := &dto.CreateIngredientRequest{
		IngredientCode: "ING001",
		Name:           "鸡胸肉",
		Description:    "新鲜鸡胸肉",
		Images:         nil,
	}

	err := service.Create(context.Background(), req)
	assert.NoError(t, err)

	var ingredient model.Ingredient
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).First(&ingredient, "ingredient_code = ?", "ING001").Error
	assert.NoError(t, err)
	assert.Equal(t, "鸡胸肉", ingredient.Name)
	assert.Equal(t, "新鲜鸡胸肉", ingredient.Description)
}

func TestIngredientService_GetByCode(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewIngredientService(cacheDB, mockImgUtil)

	ingredient := &model.Ingredient{
		IngredientCode: "ING002",
		Name:           "花生米",
		Description:    "红皮花生",
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(ingredient).Error
	assert.NoError(t, err)

	resp, err := service.GetByCode(context.Background(), "ING002")
	assert.NoError(t, err)
	assert.Equal(t, "ING002", resp.IngredientCode)
	assert.Equal(t, "花生米", resp.Name)
	assert.Equal(t, "红皮花生", resp.Description)
}

func TestIngredientService_FindByCursor(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewIngredientService(cacheDB, mockImgUtil)

	for i := 1; i <= 5; i++ {
		ingredient := &model.Ingredient{
			IngredientCode: "ING00" + string(rune('0'+i)),
			Name:           "食材" + string(rune('0'+i)),
			Description:    "描述" + string(rune('0'+i)),
		}
		err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(ingredient).Error
		assert.NoError(t, err)
	}

	result, err := service.FindByCursor(context.Background(), 0, 3)
	assert.NoError(t, err)
	assert.Len(t, result.Ingredients, 3)
	assert.True(t, result.HasMore)
	assert.Equal(t, uint64(3), result.Cursor)
}

func TestIngredientService_FindProductsByIngredientCodeAndCursor(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewIngredientService(cacheDB, mockImgUtil)

	ingredient := &model.Ingredient{
		IngredientCode: "ING003",
		Name:           "测试食材",
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(ingredient).Error
	assert.NoError(t, err)

	for i := 1; i <= 3; i++ {
		product := &model.Product{
			ProductCode:    "PROD00" + string(rune('0'+i)),
			IngredientCode: "ING003",
			Name:           "产品" + string(rune('0'+i)),
			Amount:         float64(i * 100),
			Unit:           model.UnitGram,
			Price:          float64(i * 10),
		}
		err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(product).Error
		assert.NoError(t, err)
	}

	result, err := service.FindProductsByIngredientCodeAndCursor(context.Background(), "ING003", 0, 2)
	assert.NoError(t, err)
	assert.Len(t, result.Products, 2)
	assert.True(t, result.HasMore)
	assert.Equal(t, "PROD001", result.Products[0].ProductCode)
}

func TestIngredientService_Update(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewIngredientService(cacheDB, mockImgUtil)

	ingredient := &model.Ingredient{
		IngredientCode: "ING004",
		Name:           "原名称",
		Description:    "原描述",
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(ingredient).Error
	assert.NoError(t, err)

	req := &dto.UpdateIngredientRequest{
		IngredientCode: "ING004",
		Name:           "新名称",
		Description:    "新描述",
		Images:         []dto.ImageRequest{},
		NewImages:      []dto.NewImageFile{},
	}

	err = service.Update(context.Background(), req)
	assert.NoError(t, err)

	var updatedIngredient model.Ingredient
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).First(&updatedIngredient, "ingredient_code = ?", "ING004").Error
	assert.NoError(t, err)
	assert.Equal(t, "新名称", updatedIngredient.Name)
	assert.Equal(t, "新描述", updatedIngredient.Description)
}

func TestIngredientService_Delete(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewIngredientService(cacheDB, mockImgUtil)

	ingredient := &model.Ingredient{
		IngredientCode: "ING005",
		Name:           "待删除食材",
		Description:    "描述",
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(ingredient).Error
	assert.NoError(t, err)

	err = service.Delete(context.Background(), "ING005")
	assert.NoError(t, err)

	var deletedIngredient model.Ingredient
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).First(&deletedIngredient, "ingredient_code = ?", "ING005").Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}
