package main

import (
	"fmt"
	"go-cookbook/internal/config"
	authController "go-cookbook/internal/controller/auth"
	dishController "go-cookbook/internal/controller/dish"
	ingredientController "go-cookbook/internal/controller/ingredient"
	productController "go-cookbook/internal/controller/product"
	qrcodeController "go-cookbook/internal/controller/qrcode"
	authService "go-cookbook/internal/service/auth"
	dishService "go-cookbook/internal/service/dish"
	ingredientService "go-cookbook/internal/service/ingredient"
	"go-cookbook/internal/service/jwt"
	productService "go-cookbook/internal/service/product"
	"go-cookbook/internal/utils/imgutil"
	"log"
	"os"
	"path/filepath"
	"time"

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
	imgUtil := imgutil.NewImgUtil(appcfg.ImgUtil)

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

	dishService := dishService.NewDishService(gormcDB, imgUtil)
	productService := productService.NewProductService(gormcDB, imgUtil)
	ingredientService := ingredientService.NewIngredientService(gormcDB, imgUtil)

	qrCodeController := qrcodeController.NewQRCodeController()
	dishController := dishController.NewDishController(dishService, imgUtil)
	productController := productController.NewProductController(productService, imgUtil)
	ingredientController := ingredientController.NewIngredientController(ingredientService, imgUtil)

	authService := authService.NewAuthService(appcfg.Auth.Password)
	jwtService := jwt.NewJWTService(appcfg.Auth.Password, 24*7, "go-cookbook", []string{"admin"})

	authController := authController.NewAuthController(authService)

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
