package qrcodeController

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

type QRCodeController struct {
}

func NewQRCodeController() *QRCodeController {
	return &QRCodeController{}
}

func (qc *QRCodeController) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/api/qrcode")
	{
		group.POST("", qc.GenerateQRCode)
	}
}

func (qc *QRCodeController) GenerateQRCode(gctx *gin.Context) {
	url := gctx.Query("url")
	// 一行生成 PNG 字节
	pngData, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.Data(http.StatusOK, "image/png", pngData)
}
