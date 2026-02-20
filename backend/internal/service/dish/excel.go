package dishService

import (
	"context"
	"fmt"
	"go-cookbook/internal/model"
	"go-cookbook/internal/utils"
	"log"
	"mime/multipart"

	"github.com/LouYuanbo1/go-webservice/gormx/options"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// 非常恐怖的面条代码,强烈与excel和数据库结构耦合并且暂时无法优化,需要重点关注
/*
为了减小文件体积，Excel默认会对插入的图片进行压缩。当你手动将图片拖拽缩小到单元格大小时，Excel可能会在保存文件时，
根据其内部的压缩设置，丢弃图片的像素数据以“适应”这个较小的显示尺寸。一旦原始像素丢失，即使后续在代码中尝试高清读取，也无法挽回。
修改Excel的全局设置：在 Excel 中，点击「文件」->「选项」->「高级」，在“图像大小和质量”部分，勾选 “不压缩文件中的图像”，
并将“默认分辨率”设置为 “高保真” 或最高ppi值。注意：这个设置需要在插入图片前就配置好，才能保证原始图片数据被完整保留。
*/

func (ds *dishService) Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error {
	// 解析 Excel 文件
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("打开 Excel 文件失败: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("获取工作表行失败: %w", err)
	}

	var dishes []*model.Dish
	var dishIngredients []*model.DishIngredient
	var imageURLs []*model.DishImage

	for rowNum := 2; rowNum <= len(rows); rowNum++ {

		fmt.Printf("\n处理第 %d 行,行长度 %d...\n", rowNum, len(rows[rowNum-1]))

		if rowNum%batchSize == 0 && (len(dishes) > 0 || len(dishIngredients) > 0) {

			fmt.Printf("已处理 %d 行，准备提交事务...\n", rowNum)

			if err := ds.repoFactory.Tx().Exec(ctx, func(ctx context.Context) error {
				// 第一步：创建菜品
				if len(dishes) > 0 {
					if err := ds.repoFactory.Dish().CreateInBatches(ctx, dishes, batchSize,
						options.OnConstraintColumns("dish_code"),
						options.UpdateAllOption(),
					); err != nil {
						return fmt.Errorf("创建产品失败: %w", err)
					}
				}
				// 第二步：创建菜品食材关系
				if len(dishIngredients) > 0 {
					if err := ds.repoFactory.DishIngredient().CreateInBatches(ctx, dishIngredients, batchSize,
						options.OnConstraintColumns("dish_code", "ingredient_code"),
						options.UpdateAllOption(),
					); err != nil {
						return fmt.Errorf("创建产品失败: %w", err)
					}
				}

				return nil
			}); err != nil {
				return fmt.Errorf("提交事务失败: %w", err)
			}

			dishes = dishes[:0]
			dishIngredients = dishIngredients[:0]
		}

		for i, val := range rows[rowNum-1] {
			if i < 4 {
				switch i {
				case 0:
					dishes = append(dishes, &model.Dish{
						DishCode: val,
					})
				case 1:
					dishes[len(dishes)-1].Name = val
				case 2:
					dishes[len(dishes)-1].Description = val
				case 3:
					dishes[len(dishes)-1].Recipe = val
				}
			} else {
				switch i%3 - 1 {
				case 0:
					dishIngredients = append(dishIngredients, &model.DishIngredient{
						DishCode:       dishes[len(dishes)-1].DishCode,
						IngredientCode: val,
					})
				case 1:
					dishIngredients[len(dishIngredients)-1].Quantity = val
				case 2:
					dishIngredients[len(dishIngredients)-1].Note = val
				}
			}
		}

		pictureCells, err := f.GetPictureCells(sheetName)
		if err != nil {
			return fmt.Errorf("获取图片单元格失败: %w", err)
		}

		for i, cell := range pictureCells {

			if i%batchSize == 0 && len(imageURLs) > 0 {

				fmt.Printf("已处理 %d 图片，准备提交事务...\n", i)

				if err := ds.repoFactory.DishImage().CreateInBatches(ctx, imageURLs, batchSize,
					options.OnConstraintColumns("dish_code", "sort_order"),
					options.UpdateAllOption(),
				); err != nil {
					return fmt.Errorf("创建产品图片关系失败: %w", err)
				}

				imageURLs = imageURLs[:0]
			}

			picture, err := f.GetPictures(sheetName, cell)
			if err != nil {
				continue
			}

			fmt.Printf("图片 %s 处理中...\n", cell)
			col, row, err := excelize.CellNameToCoordinates(cell)
			if err != nil {
				continue
			}

			cellDishCode, err := excelize.CoordinatesToCellName(1, row)
			if err != nil {
				continue
			}

			dishCode, err := f.GetCellValue(sheetName, cellDishCode)
			if err != nil {
				continue
			}

			if len(picture) > 0 {
				sortOrder := col - len(rows[row-1]) - 1

				url, err := utils.ProcessExcelPicture(
					ds.imgUtil,
					picture[0],
					[]string{"uploads", "dishes"},
					dishCode,
					sortOrder,
				)
				if err != nil {
					continue
				}

				imageURLs = append(imageURLs, &model.DishImage{
					DishCode: dishCode,
					// 排序,用于显示顺序
					SortOrder: sortOrder,
					ImageURL:  url,
				})
			}
		}

		if len(dishes) > 0 || len(dishIngredients) > 0 || len(imageURLs) > 0 {

			if err := ds.repoFactory.Tx().Exec(ctx, func(ctx context.Context) error {
				// 第一步：创建菜品
				if len(dishes) > 0 {
					if err := ds.repoFactory.Dish().CreateInBatches(ctx, dishes, batchSize,
						options.OnConstraintColumns("dish_code"),
						options.UpdateAllOption(),
					); err != nil {
						return fmt.Errorf("创建产品失败: %w", err)
					}
				}
				// 第二步：创建菜品食材关系
				if len(dishIngredients) > 0 {
					if err := ds.repoFactory.DishIngredient().CreateInBatches(ctx, dishIngredients, batchSize,
						options.OnConstraintColumns("dish_code", "ingredient_code"),
						options.UpdateAllOption(),
					); err != nil {
						return fmt.Errorf("创建产品失败: %w", err)
					}
				}
				// 第三步：创建菜品图片关系
				if len(imageURLs) > 0 {

					if err := ds.repoFactory.DishImage().CreateInBatches(ctx, imageURLs, batchSize,
						options.OnConstraintColumns("dish_code", "sort_order"),
						options.UpdateAllOption(),
					); err != nil {
						return fmt.Errorf("创建产品图片关系失败: %w", err)
					}
				}
				return nil
			}); err != nil {
				return fmt.Errorf("提交事务失败: %w", err)
			}
		}
	}
	return nil
}

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
