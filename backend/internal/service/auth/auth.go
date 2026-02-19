package authService

import (
	"context"
	"errors"
	"fmt"
	"go-cookbook/internal/config"
	"go-cookbook/internal/jwt"

	"github.com/LouYuanbo1/go-webservice/cryptutil"
)

var (
	ErrPasswordNotMatch = errors.New("password not match")
	ErrGenerateToken    = errors.New("generate token failed")
	ErrCheckToken       = errors.New("check token failed")
)

type AuthService interface {
	AdminLogin(ctx context.Context, password string) (string, error)
	ParseToken(tokenString string) (*jwt.CustomClaims, error)
}

type authService struct {
	bcryptX        cryptutil.CryptUtil
	hashedPassword []byte
	jwtService     jwt.JWTService
}

func NewAuthService(authConfig *config.AuthConfig) AuthService {
	cryptUtil := cryptutil.NewCryptUtil(authConfig.CryptUtil)
	hashedPassword, _ := cryptUtil.Encrypt(authConfig.Password)
	jwtService := jwt.NewJWTService(authConfig, "authService", []string{"authService"})
	return &authService{
		bcryptX:        cryptUtil,
		hashedPassword: hashedPassword,
		jwtService:     jwtService,
	}
}

func (a *authService) CheckPassword(password string) bool {
	err := a.bcryptX.CheckSecret(password, a.hashedPassword)
	return err == nil
}

func (a *authService) GenerateToken(role string) (string, error) {
	return a.jwtService.GenerateToken(role)
}

func (a *authService) ParseToken(tokenString string) (*jwt.CustomClaims, error) {
	claims, err := a.ParseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", ErrCheckToken)
	}
	return claims, nil
}

func (a *authService) AdminLogin(ctx context.Context, password string) (string, error) {
	if !a.CheckPassword(password) {
		return "", fmt.Errorf("auth: %w", ErrPasswordNotMatch)
	}
	token, err := a.GenerateToken("admin")
	if err != nil {
		return "", fmt.Errorf("auth: %w", ErrGenerateToken)
	}
	return token, nil
}
