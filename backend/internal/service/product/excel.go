package productService

import (
	"context"
	"fmt"
	"go-cookbook/internal/model"
	"go-cookbook/internal/utils"
	"log"
	"mime/multipart"
	"strconv"

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

func (ps *productService) Import(ctx context.Context, fileHeader *multipart.FileHeader, batchSize int) error {
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

	var products []*model.Product
	var imageURLs []*model.ProductImage

	for rowNum := 2; rowNum <= len(rows); rowNum++ {

		fmt.Printf("\n处理第 %d 行,行长度 %d...\n", rowNum, len(rows[rowNum-1]))

		if rowNum%batchSize == 0 && len(products) > 0 {

			fmt.Printf("已处理 %d 行，准备提交事务...\n", rowNum)

			if err := ps.repoFactory.Product().CreateInBatches(ctx, products, batchSize,
				options.OnConstraintColumns("product_code"),
				options.UpdateAllOption(),
			); err != nil {
				return fmt.Errorf("创建产品失败: %w", err)
			}

			products = products[:0]
		}

		for i, val := range rows[rowNum-1] {
			switch i {
			case 0:
				products = append(products, &model.Product{
					ProductCode: val,
				})
			case 1:
				products[len(products)-1].IngredientCode = val
			case 2:
				products[len(products)-1].Name = val
			case 3:
				products[len(products)-1].Amount, err = strconv.ParseFloat(val, 64)
				if err != nil {
					return fmt.Errorf("解析数量失败: %w", err)
				}
			case 4:
				products[len(products)-1].Unit = model.UnitType(val)
			case 5:
				products[len(products)-1].Description = val
			case 6:
				products[len(products)-1].Price, err = strconv.ParseFloat(val, 64)
				if err != nil {
					return fmt.Errorf("解析金额失败: %w", err)
				}
			case 7:
				products[len(products)-1].AllergenType = model.AllergenType(val)
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
			if err := ps.repoFactory.ProductImage().CreateInBatches(ctx, imageURLs, batchSize,
				options.OnConstraintColumns("product_code", "sort_order"),
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

		cellProductCode, err := excelize.CoordinatesToCellName(1, row)
		if err != nil {
			continue
		}

		productCode, err := f.GetCellValue(sheetName, cellProductCode)
		if err != nil {
			continue
		}

		if len(picture) > 0 {
			sortOrder := col - len(rows[row-1]) - 1

			url, err := utils.ProcessExcelPicture(
				ps.imgUtil,
				picture[0],
				[]string{"uploads", "products"},
				productCode,
				sortOrder,
			)
			if err != nil {
				continue
			}
			imageURLs = append(imageURLs, &model.ProductImage{
				ProductCode: productCode,
				// 排序,用于显示顺序
				SortOrder: sortOrder,
				ImageURL:  url,
			})
		}
	}

	if len(products) > 0 || len(imageURLs) > 0 {
		if err := ps.repoFactory.Tx().Exec(ctx, func(ctx context.Context) error {
			// 第一步：创建产品
			if len(products) > 0 {
				if err := ps.repoFactory.Product().CreateInBatches(ctx, products, batchSize,
					options.OnConstraintColumns("product_code"),
					options.UpdateAllOption(),
				); err != nil {
					return fmt.Errorf("创建产品失败: %w", err)
				}
			}
			// 第二步：创建产品图片关系
			if len(imageURLs) > 0 {

				if err := ps.repoFactory.ProductImage().CreateInBatches(ctx, imageURLs, batchSize,
					options.OnConstraintColumns("product_code", "sort_order"),
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
	return nil
}

func (ps *productService) Export(gctx *gin.Context, batchSize int) error {

	gctx.Writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	gctx.Writer.Header().Set("Content-Disposition", "attachment; filename=products.xlsx")

	f := excelize.NewFile()
	// 创建一个工作表（默认已有 Sheet1，这里直接使用）

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return fmt.Errorf("创建工作表失败: %w", err)
	}

	defer f.Close()

	// 可选：设置表头
	headers := []any{"商品编码", "食材编码", "商品名称", "商品量值", "商品单位", "商品描述", "商品价格", "过敏原类型"}
	if err := sw.SetRow("A1", headers); err != nil {
		return fmt.Errorf("设置表头失败: %w", err)
	}

	currentRow := 2

	ctx := gctx.Request.Context()

	err = ps.repoFactory.Product().FindInBatches(ctx, batchSize, func(ctx context.Context, batch int, ptrModels []*model.Product) error {
		for _, product := range ptrModels {
			// 构造一行数据（必须与表头列数一致）
			row := []any{product.ProductCode, product.IngredientCode, product.Name, product.Amount, product.Unit, product.Description, product.Price, product.AllergenType}
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
	if err != nil {
		return fmt.Errorf("导出数据失败: %w", err)
	}

	// 7. 直接写入 ResponseWriter (不保存到本地磁盘)
	// 注意：此时 excelize 会在内存中构建 ZIP 结构，然后一次性或分块写入 w
	_, err = f.WriteTo(gctx.Writer)
	if err != nil {
		return fmt.Errorf("写入响应失败: %w", err)
	}
	return nil
}
