package dishController

import (
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/jwt"
	"go-cookbook/internal/middleware"
	dishService "go-cookbook/internal/service/dish"
	"net/http"
	"strconv"

	"github.com/LouYuanbo1/go-webservice/ginutil/multipart"
	"github.com/LouYuanbo1/go-webservice/imgutil"
	"github.com/gin-gonic/gin"
)

type DishController struct {
	dishService dishService.DishService
	imgUtil     imgutil.ImgUtil
}

func NewDishController(dishService dishService.DishService, imgUtil imgutil.ImgUtil) *DishController {
	return &DishController{
		dishService: dishService,
		imgUtil:     imgUtil,
	}
}

func (dc *DishController) RegisterRoutes(router *gin.Engine, jwtService jwt.JWTService) {
	group := router.Group("/api/dishes")
	{
		group.POST("", middleware.JWTMiddleware(jwtService), dc.CreateDish)
		group.GET("", dc.FindDishesByCursor)
		group.GET("/export", middleware.JWTMiddleware(jwtService), dc.ExportDishes)
		group.GET("/:dishCode", dc.GetDishByCode)
		group.GET("/:dishCode/ingredients", dc.FindIngredientsByDishCode)
		group.PATCH("/:dishCode", middleware.JWTMiddleware(jwtService), dc.UpdateDish)
		group.DELETE("/:dishCode", middleware.JWTMiddleware(jwtService), dc.DeleteDish)
	}
}

func (dc *DishController) CreateDish(gctx *gin.Context) {
	var req dto.CreateDishRequest
	/*
		if err := gctx.ShouldBind(&req); err != nil {
			gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
			return
		}
	*/
	if err := multipart.BindMultipart(gctx, &req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}

	fmt.Printf("CreateDish req: %v\n", req)

	ctx := gctx.Request.Context()
	if err := dc.dishService.Create(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}

func (dc *DishController) FindDishesByCursor(gctx *gin.Context) {
	var req dto.CursorRequest
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	ctx := gctx.Request.Context()
	dishes, err := dc.dishService.FindByCursor(ctx, req.Cursor, req.Limit)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": dishes})
}

func (dc *DishController) ExportDishes(gctx *gin.Context) {
	batchSize, err := strconv.Atoi(gctx.DefaultQuery("batchSize", "200"))
	if err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	err = dc.dishService.Export(gctx, batchSize)
	if err != nil {
		// 如果服务已经部分写入响应（例如写入Excel过程中出错），这里再写JSON可能已无效
		// 可考虑记录错误日志，并尝试返回错误信息，但更稳妥的做法是在服务内部处理错误响应
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
}

func (dc *DishController) GetDishByCode(gctx *gin.Context) {
	dishCode := gctx.Param("dishCode")
	ctx := gctx.Request.Context()
	dish, err := dc.dishService.GetByCode(ctx, dishCode)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": dish})
}

func (dc *DishController) FindIngredientsByDishCode(gctx *gin.Context) {
	dishCode := gctx.Param("dishCode")
	var req dto.CursorRequest
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	ctx := gctx.Request.Context()
	ingredients, err := dc.dishService.FindIngredientsByDishCode(ctx, dishCode, req.Cursor, req.Limit)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": ingredients})
}

func (dc *DishController) UpdateDish(gctx *gin.Context) {
	var req dto.UpdateDishRequest
	/*
		if err := gctx.ShouldBind(&req); err != nil {
			gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
			return
		}
	*/
	if err := multipart.BindMultipart(gctx, &req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	req.DishCode = gctx.Param("dishCode")

	ctx := gctx.Request.Context()
	if err := dc.dishService.Update(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}

func (dc *DishController) DeleteDish(gctx *gin.Context) {
	dishCode := gctx.Param("dishCode")
	ctx := gctx.Request.Context()
	if err := dc.dishService.Delete(ctx, dishCode); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}
