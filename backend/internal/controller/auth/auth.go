package authController

import (
	"errors"
	authService "go-cookbook/internal/service/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService authService.AuthService
}

func NewAuthController(authService authService.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (ac *AuthController) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/api/auth")
	{
		group.POST("/admin/login", ac.AdminLogin)
	}
}

type AdminLoginReq struct {
	Password string `json:"password" form:"password" binding:"required"`
}

func (ac *AuthController) AdminLogin(gctx *gin.Context) {
	var req AdminLoginReq
	if err := gctx.ShouldBindJSON(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "error", "error": err.Error()})
		return
	}
	ctx := gctx.Request.Context()
	token, err := ac.authService.AdminLogin(ctx, req.Password)
	if err != nil {
		if errors.Is(err, authService.ErrPasswordNotMatch) {
			gctx.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "error", "error": "密码错误"})
			return
		} else if errors.Is(err, authService.ErrGenerateToken) {
			gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": "生成token失败"})
			return
		} else {
			// 其他错误
			gctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "error", "error": "未知错误"})
			return
		}
	}
	gctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "success", "data": token})
}
