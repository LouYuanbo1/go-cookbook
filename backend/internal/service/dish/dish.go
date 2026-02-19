package dishService

import (
	"context"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/repo"

	"github.com/LouYuanbo1/go-webservice/imgutil"
	"github.com/gin-gonic/gin"
)

type DishService interface {
	//创建和更新菜品需要原子性索引使用事务和更复杂的结构体，
	Create(ctx context.Context, req *dto.CreateDishRequest) error

	FindByCursor(ctx context.Context, cursor uint64, limit int) (*dto.ViewDishCardListWithCursor, error)

	Export(gctx *gin.Context, batchSize int) error

	//查询对一致性要求不高,可以使用两次分别查询菜品和对应食材,减少复杂性
	GetByCode(ctx context.Context, code string) (*dto.ViewDishResponse, error)
	//查询对一致性要求不高,可以使用两次分别查询菜品和对应食材,减少复杂性
	FindIngredientsByDishCode(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewDishIngredientCardListWithCursor, error)
	//创建和更新菜品需要原子性索引使用事务和更复杂的结构体，
	Update(ctx context.Context, req *dto.UpdateDishRequest) error
	Delete(ctx context.Context, code string) error
}

type dishService struct {
	repoFactory repo.RepoFactory
	imgUtil     imgutil.ImgUtil
}

func NewDishService(repoFactory repo.RepoFactory, imgUtil imgutil.ImgUtil) DishService {
	return &dishService{repoFactory: repoFactory, imgUtil: imgUtil}
}
