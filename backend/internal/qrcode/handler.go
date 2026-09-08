package qrcode

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

type QRCodeHandler struct {
}

func NewQRCodeHandler() *QRCodeHandler {
	return &QRCodeHandler{}
}

func (qh *QRCodeHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/api/qrcode")
	{
		group.POST("", qh.GenerateQRCode)
	}
}

func (qh *QRCodeHandler) GenerateQRCode(gctx *gin.Context) {
	url := gctx.Query("url")
	// 一行生成 PNG 字节
	pngData, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": err.Error()})
		return
	}
	gctx.Data(http.StatusOK, "image/png", pngData)
}
