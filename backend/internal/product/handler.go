package product

import (
	"go-cookbook/internal/common/dto"
	"go-cookbook/internal/common/utils/jwt"
	"go-cookbook/internal/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productSvc ProductService
}

func NewProductHandler(productSvc ProductService) *ProductHandler {
	return &ProductHandler{
		productSvc: productSvc,
	}
}

func (ph *ProductHandler) RegisterRoutes(router *gin.Engine, jwtUtil jwt.JWTUtil) {
	group := router.Group("/api/products")
	{
		group.POST("", middleware.JWTMiddleware(jwtUtil), ph.CreateProduct)
		group.POST("/import", middleware.JWTMiddleware(jwtUtil), ph.ImportProducts)
		group.GET("/export", middleware.JWTMiddleware(jwtUtil), ph.ExportProducts)
		group.GET("/:productCode", ph.GetProductByCode)
		group.GET("/:productCode/dishes", ph.FindDishesByCode)
		group.PATCH("/:productCode", middleware.JWTMiddleware(jwtUtil), ph.UpdateProduct)
		group.DELETE("/:productCode", middleware.JWTMiddleware(jwtUtil), ph.DeleteProduct)
	}
}

func (ph *ProductHandler) CreateProduct(gctx *gin.Context) {
	var req dto.CreateProductReq
	if err := gctx.ShouldBind(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	ctx := gctx.Request.Context()
	if err := ph.productSvc.Create(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.SuccessNoData())
}

func (ph *ProductHandler) GetProductByCode(gctx *gin.Context) {
	productCode := gctx.Param("productCode")
	ctx := gctx.Request.Context()
	product, err := ph.productSvc.FirstByCode(ctx, productCode)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.Success(product))
}

func (ph *ProductHandler) FindDishesByCode(gctx *gin.Context) {
	productCode := gctx.Param("productCode")
	var req dto.CursorReq[uint64]
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	ctx := gctx.Request.Context()
	dishList, err := ph.productSvc.FindDishesByCode(ctx, productCode, req.Cursor, req.Limit)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.Success(dishList))
}

func (ph *ProductHandler) ImportProducts(gctx *gin.Context) {
	fileHeader, err := gctx.FormFile("file")
	if err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	batchSize, err := strconv.Atoi(gctx.DefaultQuery("batchSize", "200"))
	if err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	ctx := gctx.Request.Context()
	if err := ph.productSvc.Import(ctx, fileHeader, batchSize); err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.SuccessNoData())
}

func (ph *ProductHandler) ExportProducts(gctx *gin.Context) {
	batchSize, err := strconv.Atoi(gctx.DefaultQuery("batchSize", "200"))
	if err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	err = ph.productSvc.Export(gctx, batchSize)
	if err != nil {
		// 如果服务已经部分写入响应（例如写入Excel过程中出错），这里再写JSON可能已无效
		// 可考虑记录错误日志，并尝试返回错误信息，但更稳妥的做法是在服务内部处理错误响应
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.SuccessNoData())
}

func (ph *ProductHandler) UpdateProduct(gctx *gin.Context) {
	var req dto.UpdateProductReq
	if err := gctx.ShouldBind(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}

	/*
		if err := multipart.BindMultipart(gctx, &req); err != nil {
			gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
			return
		}
	*/

	req.ProductCode = gctx.Param("productCode")

	ctx := gctx.Request.Context()
	if err := ph.productSvc.Update(ctx, &req); err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.SuccessNoData())
}

func (ph *ProductHandler) DeleteProduct(gctx *gin.Context) {
	productId := gctx.Param("productCode")

	ctx := gctx.Request.Context()
	if err := ph.productSvc.Delete(ctx, productId); err != nil {
		gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError(err.Error()))
		return
	}
	gctx.JSON(http.StatusOK, dto.SuccessNoData())
}
