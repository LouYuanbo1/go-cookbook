package ingredient

import (
	"fmt"
	"go-cookbook/internal/common/dto"
	"go-cookbook/internal/common/utils/jwt"
	"go-cookbook/internal/middleware"
	"strconv"

	"net/http"

	"github.com/gin-gonic/gin"
)

type IngredientHandler struct {
	ingredientSvc IngredientService
}

func NewIngredientHandler(ingredientSvc IngredientService) *IngredientHandler {
	return &IngredientHandler{
		ingredientSvc: ingredientSvc,
	}
}

func (ih *IngredientHandler) RegisterRoutes(router *gin.Engine, jwtUtil jwt.JWTUtil) {
	group := router.Group("/api/ingredients")
	{
		group.POST("", middleware.JWTMiddleware(jwtUtil), ih.CreateIngredient)
		group.GET("", ih.FindIngredientsByCursor)

		group.POST("/import", middleware.JWTMiddleware(jwtUtil), ih.ImportIngredients)
		group.GET("/export", middleware.JWTMiddleware(jwtUtil), ih.ExportIngredients)

		group.GET("/:ingredientCode", ih.FirstIngredientByCode)
		group.GET("/:ingredientCode/products", ih.FindProductsByCode)
		group.PATCH("/:ingredientCode", middleware.JWTMiddleware(jwtUtil), ih.UpdateIngredient)
		group.DELETE("/:ingredientCode", middleware.JWTMiddleware(jwtUtil), ih.DeleteIngredient)
	}
}

func (ih *IngredientHandler) CreateIngredient(gctx *gin.Context) {
	// 第一步：从请求体中解析 JSON 数据
	var req dto.CreateIngredientReq
	if err := gctx.ShouldBind(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}

	ctx := gctx.Request.Context()
	// 第二步：调用服务层创建食材
	if err := ih.ingredientSvc.Create(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	// 第三步：返回成功响应
	gctx.JSON(http.StatusCreated, dto.SuccessNoData())
}

func (ih *IngredientHandler) FindIngredientsByCursor(gctx *gin.Context) {
	var req dto.CursorReq[uint64]
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	ctx := gctx.Request.Context()
	ingredients, err := ih.ingredientSvc.FindByCursor(ctx, &req)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}

	fmt.Printf("FindIngredientsByCursor ingredients: %v\n", ingredients)

	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": ingredients})
}

func (ih *IngredientHandler) ImportIngredients(gctx *gin.Context) {
	fileHeader, err := gctx.FormFile("file")
	if err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	batchSize, err := strconv.Atoi(gctx.DefaultQuery("batchSize", "200"))
	if err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	ctx := gctx.Request.Context()
	if err := ih.ingredientSvc.Import(ctx, fileHeader, batchSize); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}

func (ih *IngredientHandler) ExportIngredients(gctx *gin.Context) {
	batchSize, err := strconv.Atoi(gctx.DefaultQuery("batchSize", "200"))
	if err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	err = ih.ingredientSvc.Export(gctx, batchSize)
	if err != nil {
		// 如果服务已经部分写入响应（例如写入Excel过程中出错），这里再写JSON可能已无效
		// 可考虑记录错误日志，并尝试返回错误信息，但更稳妥的做法是在服务内部处理错误响应
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
}

func (ih *IngredientHandler) FirstIngredientByCode(gctx *gin.Context) {
	ingredientCode := gctx.Param("ingredientCode")

	fmt.Println(ingredientCode)

	ctx := gctx.Request.Context()
	ingredient, err := ih.ingredientSvc.FirstByCode(ctx, ingredientCode)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}

	fmt.Printf("GetIngredientByCode ingredient: %v\n", ingredient)

	gctx.JSON(http.StatusOK, dto.Success(ingredient))
}

func (ih *IngredientHandler) FindProductsByCode(gctx *gin.Context) {
	ingredientCode := gctx.Param("ingredientCode")
	var req dto.CursorReq[uint64]
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	ctx := gctx.Request.Context()
	products, err := ih.ingredientSvc.FindProductsByCode(ctx, ingredientCode, req.Cursor, req.Limit)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.Success(products))
}

func (ih *IngredientHandler) UpdateIngredient(gctx *gin.Context) {

	// 强制解析 multipart 表单并打印
	/*
		testForm, err := gctx.MultipartForm()
		if err != nil {
			log.Println("MultipartForm error:", err)
		} else {
			log.Println("=== 普通字段 ===")
			for key, values := range testForm.Value {
				log.Printf("key: %q, values: %v", key, values)
			}
			log.Println("=== 文件字段 ===")
			for key, headers := range testForm.File {
				log.Printf("key: %q, file count: %d", key, len(headers))
			}
		}
	*/

	var req dto.UpdateIngredientReq

	if err := gctx.ShouldBind(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}

	req.IngredientCode = gctx.Param("ingredientCode")

	ctx := gctx.Request.Context()
	if err := ih.ingredientSvc.Update(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.SuccessNoData())
}

func (ih *IngredientHandler) DeleteIngredient(gctx *gin.Context) {
	ingredientCode := gctx.Param("ingredientCode")
	ctx := gctx.Request.Context()
	if err := ih.ingredientSvc.Delete(ctx, ingredientCode); err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.SuccessNoData())
}
