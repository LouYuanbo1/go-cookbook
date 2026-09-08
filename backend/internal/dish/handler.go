package dish

import (
	"fmt"
	"go-cookbook/internal/common/dto"
	"go-cookbook/internal/common/utils/jwt"
	"go-cookbook/internal/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DishHandler struct {
	dishSvc DishService
}

func NewDishHandler(dishSvc DishService) *DishHandler {
	return &DishHandler{
		dishSvc: dishSvc,
	}
}

func (dh *DishHandler) RegisterRoutes(router *gin.Engine, jwtUtil jwt.JWTUtil) {
	group := router.Group("/api/dishes")
	{
		group.POST("", middleware.JWTMiddleware(jwtUtil), dh.CreateDish)
		group.GET("", dh.FindDishesByCursor)
		group.GET("/export", middleware.JWTMiddleware(jwtUtil), dh.ExportDishes)
		group.GET("/:dishCode", dh.GetDishByCode)
		group.GET("/:dishCode/ingredients", dh.FindDishIngredientsByDishCode)
		group.PATCH("/:dishCode", middleware.JWTMiddleware(jwtUtil), dh.UpdateDish)
		group.DELETE("/:dishCode", middleware.JWTMiddleware(jwtUtil), dh.DeleteDish)
	}
}

func (dh *DishHandler) CreateDish(gctx *gin.Context) {
	var req dto.CreateDishReq
	if err := gctx.ShouldBind(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}

	fmt.Printf("CreateDish req: %v\n", req)

	ctx := gctx.Request.Context()
	if err := dh.dishSvc.Create(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.SuccessNoData())
}

func (dh *DishHandler) FindDishesByCursor(gctx *gin.Context) {
	var req dto.CursorReq[uint64]
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	ctx := gctx.Request.Context()
	dishes, err := dh.dishSvc.FindByCursor(ctx, req.Cursor, req.Limit)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.Success(dishes))
}

func (dh *DishHandler) ExportDishes(gctx *gin.Context) {
	batchSize, err := strconv.Atoi(gctx.DefaultQuery("batchSize", "200"))
	if err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	err = dh.dishSvc.Export(gctx, batchSize)
	if err != nil {
		// 如果服务已经部分写入响应（例如写入Excel过程中出错），这里再写JSON可能已无效
		// 可考虑记录错误日志，并尝试返回错误信息，但更稳妥的做法是在服务内部处理错误响应
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
}

func (dh *DishHandler) GetDishByCode(gctx *gin.Context) {
	dishCode := gctx.Param("dishCode")
	ctx := gctx.Request.Context()
	dish, err := dh.dishSvc.FirstByCode(ctx, dishCode)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.Success(dish))
}

func (dh *DishHandler) FindDishIngredientsByDishCode(gctx *gin.Context) {
	dishCode := gctx.Param("dishCode")
	var req dto.CursorReq[uint64]
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	ctx := gctx.Request.Context()
	ingredients, err := dh.dishSvc.FindDishIngredientsByCode(ctx, dishCode, req.Cursor, req.Limit)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.Success(ingredients))
}

func (dh *DishHandler) UpdateDish(gctx *gin.Context) {
	var req dto.UpdateDishReq
	if err := gctx.ShouldBind(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	req.DishCode = gctx.Param("dishCode")

	ctx := gctx.Request.Context()
	if err := dh.dishSvc.Update(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.SuccessNoData())
}

func (dh *DishHandler) DeleteDish(gctx *gin.Context) {
	dishCode := gctx.Param("dishCode")
	ctx := gctx.Request.Context()
	if err := dh.dishSvc.Delete(ctx, dishCode); err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.SuccessNoData())
}
