package ingredientController

import (
	"fmt"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/jwt"
	"go-cookbook/internal/middleware"
	ingredientService "go-cookbook/internal/service/ingredient"
	"log"
	"strconv"

	"net/http"

	"github.com/LouYuanbo1/go-webservice/ginutil/multipart"
	"github.com/LouYuanbo1/go-webservice/imgutil"
	"github.com/gin-gonic/gin"
)

type IngredientController struct {
	ingredientService ingredientService.IngredientService
	imgUtil           imgutil.ImgUtil
}

func NewIngredientController(ingredientService ingredientService.IngredientService, imgUtil imgutil.ImgUtil) *IngredientController {
	return &IngredientController{
		ingredientService: ingredientService,
		imgUtil:           imgUtil,
	}
}

func (ic *IngredientController) RegisterRoutes(router *gin.Engine, jwtService jwt.JWTService) {
	group := router.Group("/api/ingredients")
	{
		group.POST("", middleware.JWTMiddleware(jwtService), ic.CreateIngredient)
		group.GET("", ic.FindIngredientsByCursor)

		group.POST("/import", middleware.JWTMiddleware(jwtService), ic.ImportIngredients)
		group.GET("/export", middleware.JWTMiddleware(jwtService), ic.ExportIngredients)

		group.GET("/:ingredientCode", ic.GetIngredientByCode)
		group.GET("/:ingredientCode/products", ic.FindProductsByIngredientCodeAndCursor)
		group.PATCH("/:ingredientCode", middleware.JWTMiddleware(jwtService), ic.UpdateIngredient)
		group.DELETE("/:ingredientCode", middleware.JWTMiddleware(jwtService), ic.DeleteIngredient)
	}
}

func (ic *IngredientController) CreateIngredient(gctx *gin.Context) {
	// 第一步：从请求体中解析 JSON 数据
	var req dto.CreateIngredientRequest
	if err := gctx.ShouldBind(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}

	fmt.Printf("CreateIngredient req: %v\n", req)

	ctx := gctx.Request.Context()
	// 第二步：调用服务层创建食材
	if err := ic.ingredientService.Create(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	// 第三步：返回成功响应
	gctx.JSON(http.StatusCreated, gin.H{"code": http.StatusCreated, "msg": "success"})
}

func (ic *IngredientController) FindIngredientsByCursor(gctx *gin.Context) {
	var req dto.CursorRequest
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	ctx := gctx.Request.Context()
	ingredients, err := ic.ingredientService.FindByCursor(ctx, req.Cursor, req.Limit)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}

	fmt.Printf("FindIngredientsByCursor ingredients: %v\n", ingredients)

	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": ingredients})
}

func (ic *IngredientController) ImportIngredients(gctx *gin.Context) {
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
	if err := ic.ingredientService.Import(ctx, fileHeader, batchSize); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}

func (ic *IngredientController) ExportIngredients(gctx *gin.Context) {
	batchSize, err := strconv.Atoi(gctx.DefaultQuery("batchSize", "200"))
	if err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	err = ic.ingredientService.Export(gctx, batchSize)
	if err != nil {
		// 如果服务已经部分写入响应（例如写入Excel过程中出错），这里再写JSON可能已无效
		// 可考虑记录错误日志，并尝试返回错误信息，但更稳妥的做法是在服务内部处理错误响应
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
}

func (ic *IngredientController) GetIngredientByCode(gctx *gin.Context) {
	ingredientCode := gctx.Param("ingredientCode")

	fmt.Println(ingredientCode)

	ctx := gctx.Request.Context()
	ingredient, err := ic.ingredientService.GetByCode(ctx, ingredientCode)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}

	fmt.Printf("GetIngredientByCode ingredient: %v\n", ingredient)

	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": ingredient})
}

func (ic *IngredientController) FindProductsByIngredientCodeAndCursor(gctx *gin.Context) {
	ingredientCode := gctx.Param("ingredientCode")
	var req dto.CursorRequest
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	ctx := gctx.Request.Context()
	products, err := ic.ingredientService.FindProductsByIngredientCodeAndCursor(ctx, ingredientCode, req.Cursor, req.Limit)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": products})
}

func (ic *IngredientController) UpdateIngredient(gctx *gin.Context) {

	// 强制解析 multipart 表单并打印
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

	var req dto.UpdateIngredientRequest

	/*
		if err := utils.BindMultipart(gctx, &req); err != nil {
			gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
			return
		}
	*/

	if err := multipart.BindMultipart(gctx, &req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}

	fmt.Printf("UpdateIngredient req: %v\n", req.Images)

	req.IngredientCode = gctx.Param("ingredientCode")

	ctx := gctx.Request.Context()
	if err := ic.ingredientService.Update(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}

func (ic *IngredientController) DeleteIngredient(gctx *gin.Context) {
	ingredientCode := gctx.Param("ingredientCode")
	ctx := gctx.Request.Context()
	if err := ic.ingredientService.Delete(ctx, ingredientCode); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}
