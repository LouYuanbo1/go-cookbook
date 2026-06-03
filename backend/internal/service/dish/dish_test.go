package dishService

import (
	"context"
	"image"
	"strconv"
	"testing"

	"time"

	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/cache/driver/redis"
	"github.com/LouYuanbo1/go-webservice/gormc"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/LouYuanbo1/go-webservice/singleflightx"
	"github.com/alicebob/miniredis/v2"
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
		&model.Dish{},
		&model.DishImage{},
		&model.DishIngredient{},
		&model.Ingredient{},
		&model.IngredientImage{},
		&model.Product{},
		&model.ProductImage{},
	)
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	gormxDB := gormx.NewDB(db)

	mr, err := miniredis.Run()
	assert.NoError(t, err)

	// 在测试结束时关闭 miniredis
	t.Cleanup(func() {
		mr.Close()
	})

	// 将端口字符串转换为整数
	port, err := strconv.Atoi(mr.Port())
	assert.NoError(t, err)

	// 使用 miniredis 的地址创建配置
	config := &redis.Config{
		Host: mr.Host(),
		Port: port,
	}

	client, err := redis.InitRedisClient(config)
	assert.NoError(t, err)

	cacher, err := redis.NewRedisCache(client, singleflightx.NewSingleFlight())
	assert.NoError(t, err)

	cacheClient := cache.NewClient(cacher)

	cacheDB := gormc.NewCacheDB(gormxDB, cacheClient, &gormc.Config{
		TTL: time.Minute,
	})

	return cacheDB
}

func TestDishService_Create(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewDishService(cacheDB, mockImgUtil)

	req := &dto.CreateDishRequest{
		DishCode:    "DISH001",
		Name:        "宫保鸡丁",
		Description: "经典川菜",
		Recipe:      "1. 准备食材...",
		Ingredients: []*dto.CreateDishIngredientRequest{
			{IngredientCode: "ING001", Quantity: "200g", Note: "鸡胸肉"},
			{IngredientCode: "ING002", Quantity: "50g", Note: "花生米"},
		},
		Images: nil,
	}

	err := service.Create(context.Background(), req)
	assert.NoError(t, err)

	var dish model.Dish
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).First(&dish, "dish_code = ?", "DISH001").Error
	assert.NoError(t, err)
	assert.Equal(t, "宫保鸡丁", dish.Name)
	assert.Equal(t, "经典川菜", dish.Description)

	var ingredients []model.DishIngredient
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).Where("dish_code = ?", "DISH001").Find(&ingredients).Error
	assert.NoError(t, err)
	assert.Len(t, ingredients, 2)
}

func TestDishService_GetByCode(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewDishService(cacheDB, mockImgUtil)

	dish := &model.Dish{
		DishCode:    "DISH002",
		Name:        "鱼香肉丝",
		Description: "酸甜可口",
		Recipe:      "做法...",
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(dish).Error
	assert.NoError(t, err)

	resp, err := service.GetByCode(context.Background(), "DISH002")
	assert.NoError(t, err)
	assert.Equal(t, "DISH002", resp.DishCode)
	assert.Equal(t, "鱼香肉丝", resp.Name)
}

func TestDishService_FindByCursor(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewDishService(cacheDB, mockImgUtil)

	for i := 1; i <= 5; i++ {
		dish := &model.Dish{
			DishCode:    "DISH00" + string(rune('0'+i)),
			Name:        "菜品" + string(rune('0'+i)),
			Description: "描述" + string(rune('0'+i)),
			Recipe:      "做法" + string(rune('0'+i)),
		}
		err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(dish).Error
		assert.NoError(t, err)
	}

	result, err := service.FindByCursor(context.Background(), 0, 3)
	assert.NoError(t, err)
	assert.Len(t, result.Dishes, 3)
	assert.True(t, result.HasMore)
	assert.Equal(t, uint64(3), result.Cursor)
}

func TestDishService_Update(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewDishService(cacheDB, mockImgUtil)

	dish := &model.Dish{
		DishCode:    "DISH003",
		Name:        "原名称",
		Description: "原描述",
		Recipe:      "原做法",
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(dish).Error
	assert.NoError(t, err)

	req := &dto.UpdateDishRequest{
		DishCode:    "DISH003",
		Name:        "新名称",
		Description: "新描述",
		Recipe:      "新做法",
		Ingredients: []*dto.UpdateDishIngredientRequest{},
		NewImages:   []dto.NewImageFile{},
		Images:      []dto.ImageRequest{},
	}

	err = service.Update(context.Background(), req)
	assert.NoError(t, err)

	var updatedDish model.Dish
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).First(&updatedDish, "dish_code = ?", "DISH003").Error
	assert.NoError(t, err)
	assert.Equal(t, "新名称", updatedDish.Name)
	assert.Equal(t, "新描述", updatedDish.Description)
	assert.Equal(t, "新做法", updatedDish.Recipe)
}

func TestDishService_Delete(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewDishService(cacheDB, mockImgUtil)

	dish := &model.Dish{
		DishCode:    "DISH004",
		Name:        "待删除菜品",
		Description: "描述",
		Recipe:      "做法",
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(dish).Error
	assert.NoError(t, err)

	err = service.Delete(context.Background(), "DISH004")
	assert.NoError(t, err)

	var deletedDish model.Dish
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).First(&deletedDish, "dish_code = ?", "DISH004").Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestDishService_FindIngredientsByDishCode(t *testing.T) {
	cacheDB := setupTestDB(t)
	mockImgUtil := new(MockImgUtil)
	service := NewDishService(cacheDB, mockImgUtil)

	ingredient1 := &model.Ingredient{
		IngredientCode: "ING001",
		Name:           "鸡肉",
	}
	ingredient2 := &model.Ingredient{
		IngredientCode: "ING002",
		Name:           "花生",
	}
	err := cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(ingredient1).Error
	assert.NoError(t, err)
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(ingredient2).Error
	assert.NoError(t, err)

	dish := &model.Dish{
		DishCode:    "DISH005",
		Name:        "测试菜品",
		Description: "描述",
		Recipe:      "做法",
	}
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(dish).Error
	assert.NoError(t, err)

	dishIngredient1 := &model.DishIngredient{
		DishCode:       "DISH005",
		IngredientCode: "ING001",
		Quantity:       "100g",
		Note:           "鸡胸肉",
	}
	dishIngredient2 := &model.DishIngredient{
		DishCode:       "DISH005",
		IngredientCode: "ING002",
		Quantity:       "50g",
		Note:           "花生米",
	}
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(dishIngredient1).Error
	assert.NoError(t, err)
	err = cacheDB.GetXDB().GetDBWithContext(context.Background()).Create(dishIngredient2).Error
	assert.NoError(t, err)

	result, err := service.FindIngredientsByDishCode(context.Background(), "DISH005", 0, 10)
	assert.NoError(t, err)
	assert.Len(t, result.DishIngredients, 2)
	assert.Equal(t, "ING001", result.DishIngredients[0].IngredientCode)
	assert.Equal(t, "鸡肉", result.DishIngredients[0].Name)
}
