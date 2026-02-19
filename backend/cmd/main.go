package main

import (
	"go-cookbook/internal/config"
	authController "go-cookbook/internal/controller/auth"
	dishController "go-cookbook/internal/controller/dish"
	ingredientController "go-cookbook/internal/controller/ingredient"
	productController "go-cookbook/internal/controller/product"
	"go-cookbook/internal/jwt"
	"go-cookbook/internal/repo"
	authService "go-cookbook/internal/service/auth"
	dishService "go-cookbook/internal/service/dish"
	ingredientService "go-cookbook/internal/service/ingredient"
	productService "go-cookbook/internal/service/product"
	"log"
	"os"
	"path/filepath"

	"github.com/LouYuanbo1/go-webservice/gormx"
	"github.com/LouYuanbo1/go-webservice/imgutil"
	"github.com/gin-gonic/gin"
)

func main() {
	appcfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取当前工作目录失败: %v", err)
	}
	appcfg.DB.SchemaFile = filepath.Join(wd, "..", appcfg.DB.SchemaFile)
	db, err := gormx.InitGorm(&appcfg.DB)
	if err != nil {
		log.Fatalf("初始化GORM失败: %v", err)
	}
	factory := repo.NewRepoFactory(db)

	imgUtil := imgutil.NewImgUtil(appcfg.ImgUtil)

	dishService := dishService.NewDishService(factory, imgUtil)
	productService := productService.NewProductService(factory, imgUtil)
	ingredientService := ingredientService.NewIngredientService(factory, imgUtil)

	dishController := dishController.NewDishController(dishService, imgUtil)
	productController := productController.NewProductController(productService, imgUtil)
	ingredientController := ingredientController.NewIngredientController(ingredientService, imgUtil)

	authService := authService.NewAuthService(&appcfg.Auth)
	jwtService := jwt.NewJWTService(&appcfg.Auth, "go-cookbook", []string{"admin"})

	authController := authController.NewAuthController(authService)

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20

	router.Static("/uploads", "./uploads")

	dishController.RegisterRoutes(router, jwtService)
	productController.RegisterRoutes(router, jwtService)
	ingredientController.RegisterRoutes(router, jwtService)
	authController.RegisterRoutes(router)

	if err := router.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
