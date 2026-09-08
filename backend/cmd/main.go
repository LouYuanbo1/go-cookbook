package main

import (
	"fmt"
	"go-cookbook/internal/common/utils/img"
	"go-cookbook/internal/common/utils/jwt"
	"go-cookbook/internal/common/utils/tempfs"
	"go-cookbook/internal/config"
	"go-cookbook/internal/qrcode"
	"log"
	"os"
	"path/filepath"
	"time"

	"go-cookbook/internal/auth"
	"go-cookbook/internal/dish"
	"go-cookbook/internal/ingredient"
	"go-cookbook/internal/product"

	"github.com/LouYuanbo1/go-webservice/breaker"
	"github.com/LouYuanbo1/go-webservice/cache"
	"github.com/LouYuanbo1/go-webservice/cache/driver/redis"
	"github.com/LouYuanbo1/go-webservice/gormc"
	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/LouYuanbo1/go-webservice/singleflightx"
	"github.com/gin-gonic/gin"
	red "github.com/redis/go-redis/v9"
)

func main() {
	appcfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}
	fmt.Printf("配置: %+v\n", appcfg)
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取当前工作目录失败: %v", err)
	}
	appcfg.DB.SchemaFile = filepath.Join(wd, "..", appcfg.DB.SchemaFile)
	db, err := gormx.InitGorm(&appcfg.DB)
	if err != nil {
		log.Fatalf("初始化GORM失败: %v", err)
	}
	imgUtil := img.NewImgUtil(appcfg.Img)

	redisBreaker := breaker.NewBreaker(
		breaker.WithName("redis"),
	)

	redisHook := redis.NewBreakerHook(redisBreaker)

	redisClient, err := redis.InitRedisClient(&appcfg.Redis, []red.Hook{redisHook}...)
	if err != nil {
		panic(err)
	}
	cacher, err := redis.NewRedisCache(redisClient, singleflightx.NewSingleFlight())
	if err != nil {
		panic(err)
	}
	client := cache.NewClient(cacher)

	gormxDB := gormx.NewDB(db)

	gormcDB := gormc.NewCacheDB(gormxDB, client, &gormc.Config{
		TTL:                                20 * time.Second,
		CacheSafeGapBetweenIndexAndPrimary: 5 * time.Second,
	})

	tempFs := tempfs.NewTempFs(24*time.Hour, 24*time.Hour)

	dishService := dish.NewDishService(gormcDB, imgUtil, tempFs, filepath.Join("tempfile", "dish"), filepath.Join("uploads", "dish"))
	productService := product.NewProductService(gormcDB, imgUtil, tempFs, filepath.Join("tempfile", "product"), filepath.Join("uploads", "product"))
	ingredientService := ingredient.NewIngredientService(gormcDB, imgUtil, tempFs, filepath.Join("tempfile", "ingredient"), filepath.Join("uploads", "ingredient"))

	qrCodeController := qrcode.NewQRCodeHandler()
	dishController := dish.NewDishHandler(dishService)
	productController := product.NewProductHandler(productService)
	ingredientController := ingredient.NewIngredientHandler(ingredientService)

	authService := auth.NewAuthService(appcfg.Auth.Password)
	jwtService := jwt.NewJWTUtil(appcfg.Auth.Password, 24*7, "go-cookbook", []string{"admin"})

	authController := auth.NewAuthHandler(authService)

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20

	router.Static("/uploads", "./uploads")

	qrCodeController.RegisterRoutes(router)
	dishController.RegisterRoutes(router, jwtService)
	productController.RegisterRoutes(router, jwtService)
	ingredientController.RegisterRoutes(router, jwtService)
	authController.RegisterRoutes(router)

	if err := router.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
