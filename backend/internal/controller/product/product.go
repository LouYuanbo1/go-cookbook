package productController

import (
	"go-cookbook/internal/dto"
	"go-cookbook/internal/jwt"
	"go-cookbook/internal/middleware"
	productService "go-cookbook/internal/service/product"
	"net/http"
	"strconv"

	"github.com/LouYuanbo1/go-webservice/ginutil/multipart"
	"github.com/LouYuanbo1/go-webservice/imgutil"
	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productService productService.ProductService
	imgUtil        imgutil.ImgUtil
}

func NewProductController(productService productService.ProductService, imgUtil imgutil.ImgUtil) *ProductController {
	return &ProductController{
		productService: productService,
		imgUtil:        imgUtil,
	}
}

func (pc *ProductController) RegisterRoutes(router *gin.Engine, jwtService jwt.JWTService) {
	group := router.Group("/api/products")
	{
		group.POST("", middleware.JWTMiddleware(jwtService), pc.CreateProduct)
		group.POST("/import", middleware.JWTMiddleware(jwtService), pc.ImportProducts)
		group.GET("/export", middleware.JWTMiddleware(jwtService), pc.ExportProducts)
		group.GET("/:productCode", pc.GetProductByCode)
		group.GET("/:productCode/dishes", pc.FindDishesByProductCodeAndCursor)
		group.PATCH("/:productCode", middleware.JWTMiddleware(jwtService), pc.UpdateProduct)
		group.DELETE("/:productCode", middleware.JWTMiddleware(jwtService), pc.DeleteProduct)
	}
}

func (pc *ProductController) CreateProduct(gctx *gin.Context) {
	var req dto.CreateProductRequest
	if err := gctx.ShouldBind(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	ctx := gctx.Request.Context()
	if err := pc.productService.Create(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}

func (pc *ProductController) GetProductByCode(gctx *gin.Context) {
	productCode := gctx.Param("productCode")
	ctx := gctx.Request.Context()
	product, err := pc.productService.GetByCode(ctx, productCode)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": product})
}

func (pc *ProductController) FindDishesByProductCodeAndCursor(gctx *gin.Context) {
	productCode := gctx.Param("productCode")
	var req dto.CursorRequest
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	ctx := gctx.Request.Context()
	dishList, err := pc.productService.FindDishesByProductCodeAndCursor(ctx, productCode, req.Cursor, req.Limit)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": dishList})
}

func (pc *ProductController) ImportProducts(gctx *gin.Context) {
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
	if err := pc.productService.Import(ctx, fileHeader, batchSize); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}

func (pc *ProductController) ExportProducts(gctx *gin.Context) {
	batchSize, err := strconv.Atoi(gctx.DefaultQuery("batchSize", "200"))
	if err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	err = pc.productService.Export(gctx, batchSize)
	if err != nil {
		// 如果服务已经部分写入响应（例如写入Excel过程中出错），这里再写JSON可能已无效
		// 可考虑记录错误日志，并尝试返回错误信息，但更稳妥的做法是在服务内部处理错误响应
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
}

func (pc *ProductController) UpdateProduct(gctx *gin.Context) {
	var req dto.UpdateProductRequest
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

	req.ProductCode = gctx.Param("productCode")

	ctx := gctx.Request.Context()
	if err := pc.productService.Update(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}

func (pc *ProductController) DeleteProduct(gctx *gin.Context) {
	productId := gctx.Param("productCode")
	ctx := gctx.Request.Context()
	if err := pc.productService.Delete(ctx, productId); err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success"})
}
