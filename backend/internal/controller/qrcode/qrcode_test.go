package qrcodeController

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	return r
}

func TestQRCodeController_GenerateQRCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	controller := NewQRCodeController()

	router.POST("/api/qrcode", controller.GenerateQRCode)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/qrcode?url=https://example.com", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Body.Bytes())
}

func TestQRCodeController_GenerateQRCode_EmptyURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	controller := NewQRCodeController()

	router.POST("/api/qrcode", controller.GenerateQRCode)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/qrcode?url=", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}
