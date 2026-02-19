package ingredientService

import (
	"context"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/repo"
	"mime/multipart"

	"github.com/LouYuanbo1/go-webservice/imgutil"
	"github.com/gin-gonic/gin"
)

type IngredientService interface {
	Create(ctx context.Context, req *dto.CreateIngredientRequest) error
	GetByCode(ctx context.Context, code string) (*dto.ViewIngredientResponse, error)

	Import(ctx context.Context, file *multipart.FileHeader, batchSize int) error

	Export(gctx *gin.Context, batchSize int) error

	FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.ViewIngredientCardListWithCursor, error)
	FindProductsByIngredientCodeAndCursor(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewProductCardListWithCursor, error)
	Update(ctx context.Context, req *dto.UpdateIngredientRequest) error
	Delete(ctx context.Context, code string) error
}

type ingredientService struct {
	repoFactory repo.RepoFactory
	imgUtil     imgutil.ImgUtil
}

func NewIngredientService(repoFactory repo.RepoFactory, imgUtil imgutil.ImgUtil) IngredientService {
	return &ingredientService{repoFactory: repoFactory, imgUtil: imgUtil}
}
