package auth

import (
	"context"
	"errors"
	"fmt"
	"go-cookbook/internal/common/utils/jwt"

	"golang.org/x/crypto/bcrypt"
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
	hashedPassword []byte
	jwtUtil        jwt.JWTUtil
}

func NewAuthService(password string) AuthService {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	jwtUtil := jwt.NewJWTUtil(password, 24*7, "authService", []string{"authService"})
	return &authService{
		hashedPassword: hashedPassword,
		jwtUtil:        jwtUtil,
	}
}

func (a *authService) checkPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(a.hashedPassword, []byte(password))
	return err == nil
}

func (a *authService) generateToken(role string) (string, error) {
	return a.jwtUtil.GenerateToken(role)
}

func (a *authService) ParseToken(tokenString string) (*jwt.CustomClaims, error) {
	claims, err := a.jwtUtil.ParseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", ErrCheckToken)
	}
	return claims, nil
}

func (a *authService) AdminLogin(ctx context.Context, password string) (string, error) {
	if !a.checkPassword(password) {
		return "", fmt.Errorf("auth: %w", ErrPasswordNotMatch)
	}
	token, err := a.generateToken("admin")
	if err != nil {
		return "", fmt.Errorf("auth: %w", ErrGenerateToken)
	}
	return token, nil
}
