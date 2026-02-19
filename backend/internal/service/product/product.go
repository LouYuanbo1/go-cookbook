package productService

import (
	"context"
	"go-cookbook/internal/dto"
	"go-cookbook/internal/repo"
	"mime/multipart"

	"github.com/LouYuanbo1/go-webservice/imgutil"
	"github.com/gin-gonic/gin"
)

type ProductService interface {
	//创建和更新产品需要原子性索引使用事务和更复杂的结构体，
	Create(ctx context.Context, req *dto.CreateProductRequest) error
	//查询对一致性要求不高,可以使用两次分别查询产品和对应菜品,减少复杂性
	//使用Get获取基础信息,之后使用Find获取对应菜品,考虑到食材可能有极多对应菜品，e.g"盐",所以这里使用游标
	GetByCode(ctx context.Context, code string) (*dto.ViewProductResponse, error)

	Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error
	Export(gctx *gin.Context, batchSize int) error

	FindDishesByProductCodeAndCursor(ctx context.Context, code string, cursor uint64, limit int) (*dto.ViewDishCardListWithCursor, error)
	//更新产品需要原子性索引使用事务和更复杂的结构体，
	Update(ctx context.Context, req *dto.UpdateProductRequest) error
	Delete(ctx context.Context, code string) error
}

type productService struct {
	repoFactory repo.RepoFactory
	imgUtil     imgutil.ImgUtil
}

func NewProductService(repoFactory repo.RepoFactory, imgUtil imgutil.ImgUtil) ProductService {
	return &productService{repoFactory: repoFactory, imgUtil: imgUtil}
}
