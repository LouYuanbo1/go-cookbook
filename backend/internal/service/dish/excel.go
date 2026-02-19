package dishService

import (
	"context"
	"fmt"
	"go-cookbook/internal/model"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (ds *dishService) Export(gctx *gin.Context, batchSize int) error {

	gctx.Writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	gctx.Writer.Header().Set("Content-Disposition", "attachment; filename=dishes.xlsx")

	f := excelize.NewFile()
	// 创建一个工作表（默认已有 Sheet1，这里直接使用）

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return fmt.Errorf("创建工作表失败: %w", err)
	}

	defer f.Close()

	// 可选：设置表头
	headers := []any{"菜品编码", "菜品名称", "菜品描述", "菜品做法"}
	if err := sw.SetRow("A1", headers); err != nil {
		return fmt.Errorf("设置表头失败: %w", err)
	}

	currentRow := 2

	ctx := gctx.Request.Context()

	err = ds.repoFactory.Dish().
		FindInBatches(ctx, batchSize, func(ctx context.Context, batch int, ptrModels []*model.Dish) error {
			for _, dish := range ptrModels {
				// 构造一行数据（必须与表头列数一致）
				row := []any{dish.DishCode, dish.Name, dish.Description, dish.Recipe}
				// 计算单元格坐标，例如 A2, B2, ... 然后下一行 A3...
				cell, err := excelize.CoordinatesToCellName(1, currentRow) // 1 表示第 A 列
				if err != nil {
					return fmt.Errorf("计算单元格坐标失败: %w", err)
				}
				if err := sw.SetRow(cell, row); err != nil {
					return fmt.Errorf("写入行 %d 失败: %v", currentRow, err)
				}
				currentRow++
			}
			// 完成流式写入
			if err := sw.Flush(); err != nil {
				return fmt.Errorf("写入行 %d 失败: %w", currentRow, err)
			}
			// 可选：打印进度
			log.Printf("已处理批次 %d,写入 %d 行", batch, len(ptrModels))
			return nil
		})

	// 7. 直接写入 ResponseWriter (不保存到本地磁盘)
	// 注意：此时 excelize 会在内存中构建 ZIP 结构，然后一次性或分块写入 w
	_, err = f.WriteTo(gctx.Writer)
	if err != nil {
		return fmt.Errorf("写入响应失败: %w", err)
	}
	return nil
}
