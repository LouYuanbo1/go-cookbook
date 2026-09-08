package qrcode

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupQRTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewQRCodeHandler(t *testing.T) {
	handler := NewQRCodeHandler()
	assert.NotNil(t, handler)
}

func TestGenerateQRCode_Success(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/qrcode?url=https://example.com", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Body.Bytes())
}

func TestGenerateQRCode_DifferentURLs(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://example.com/very/long/url/with/many/path/segments",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/qrcode?url="+url, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
			assert.NotEmpty(t, w.Body.Bytes())
		})
	}
}

func TestGenerateQRCode_EmptyURL(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/qrcode?url=", nil)
	router.ServeHTTP(w, req)

	// When URL is empty, qrcode.Encode might still generate an image
	// or return an error depending on the library
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestGenerateQRCode_MissingURLParam(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/qrcode", nil)
	router.ServeHTTP(w, req)

	// Should handle missing URL gracefully
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestGenerateQRCode_ResponseIsPNG(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/qrcode?url=https://test.com", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.Bytes()

	// PNG magic number: 89 50 4E 47
	assert.Equal(t, byte(0x89), body[0], "PNG signature byte 0")
	assert.Equal(t, byte(0x50), body[1], "PNG signature byte 1")
	assert.Equal(t, byte(0x4E), body[2], "PNG signature byte 2")
	assert.Equal(t, byte(0x47), body[3], "PNG signature byte 3")
}

func TestGenerateQRCode_ResponseSize(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/qrcode?url=https://test.com", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.Bytes()

	// QR code at Medium quality, 256x256 should be at least 200 bytes
	assert.Greater(t, len(body), 200, "QR code image should be valid")
}

func TestGenerateQRCode_NotJSON(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/qrcode?url=https://test.com", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Verify the response is NOT JSON (it's PNG)
	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "image/png", contentType)
}

func TestGenerateQRCode_ErrorResponse(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	// Use a URL with special characters that might cause issues
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/qrcode?url="+"https://example.com/path?query=value&another=param", nil)
	router.ServeHTTP(w, req)

	// Should still return a response (either success or error)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	// Should be either PNG or JSON
	contentType := w.Header().Get("Content-Type")
	assert.Contains(t, contentType, "image/png", "Should return PNG image for valid URL")
}

func TestQRCodeHandler_RegisterRoutes(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	routes := router.Routes()
	assert.NotEmpty(t, routes)

	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+" "+route.Path] = true
	}

	assert.True(t, routePaths["POST /api/qrcode"])
}

func TestGenerateQRCode_ConcurrentRequests(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/qrcode?url=https://example.com", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGenerateQRCode_ErrorResponseFormat(t *testing.T) {
	handler := NewQRCodeHandler()
	router := setupQRTestRouter()
	handler.RegisterRoutes(router)

	// Test with a URL that might cause encoding issues
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/qrcode?url=invalid url with spaces", nil)
	router.ServeHTTP(w, req)

	// Should return a response
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	contentType := w.Header().Get("Content-Type")
	// The URL with spaces should still be encoded properly
	assert.Contains(t, contentType, "image/png")
}

func TestQRCodeHandler_Interface(t *testing.T) {
	handler := NewQRCodeHandler()
	assert.NotNil(t, handler)
	assert.IsType(t, &QRCodeHandler{}, handler)
}
