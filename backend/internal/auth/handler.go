package auth

import (
	"errors"
	"go-cookbook/internal/common/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc AuthService
}

func NewAuthHandler(authSvc AuthService) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
	}
}

func (ah *AuthHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/api/auth")
	{
		group.POST("/admin/login", ah.AdminLogin)
	}
}

func (ah *AuthHandler) AdminLogin(gctx *gin.Context) {
	var req AdminLoginReq
	if err := gctx.ShouldBindJSON(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, dto.FailBadReq(err.Error()))
		return
	}
	ctx := gctx.Request.Context()
	token, err := ah.authSvc.AdminLogin(ctx, req.Password)
	if err != nil {
		if errors.Is(err, ErrPasswordNotMatch) {
			gctx.JSON(http.StatusUnauthorized, dto.FailUnauthorized("密码错误"))
			return
		} else if errors.Is(err, ErrGenerateToken) {
			gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError("生成token失败"))
			return
		} else {
			// 其他错误
			gctx.JSON(http.StatusInternalServerError, dto.FailInternalServerError("未知错误"))
			return
		}
	}
	gctx.JSON(http.StatusOK, dto.Success(token))
}
