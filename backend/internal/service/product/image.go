package productService

import (
	"fmt"
	"go-cookbook/internal/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/LouYuanbo1/go-webservice/imgutil/options"
	"github.com/xuri/excelize/v2"
)

func (ps *productService) processImage(fileHeader *multipart.FileHeader, productCode string) (string, error) {
	isImage, err := utils.IsImage(fileHeader)
	if err != nil {
		return "", fmt.Errorf("error checking image MIME type: %w", err)
	}
	if !isImage {
		return "", fmt.Errorf("file is not an image")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	dir := filepath.Join("uploads", "products")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp(dir, fmt.Sprintf("%s_*.jpg", productCode))
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // 确保临时文件被删除
	defer tempFile.Close()

	// 将上传的文件内容复制到临时文件
	_, err = io.Copy(tempFile, file)
	if err != nil {
		return "", fmt.Errorf("error copying file: %w", err)
	}

	// 因为imaging可以从文件路径读取，所以直接使用临时文件路径
	img, err := ps.imgUtil.Load(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("error loading image: %w", err)
	}

	// 调整图片大小
	img = ps.imgUtil.Thumbnail(img)

	filename := fmt.Sprintf("%s.jpg", productCode)
	filename = ps.imgUtil.WithUnixNanoTimestamp(filename)
	// 保存图片到临时文件
	err = ps.imgUtil.Save(img, filename, options.WithStorageDirOption(dir))
	if err != nil {
		return "", fmt.Errorf("error saving image: %w", err)
	}

	fullPath := filepath.Join(dir, filename)
	urlPath := filepath.ToSlash(fullPath)
	urlPath = fmt.Sprintf("/%s", urlPath)

	return urlPath, nil
}

func (ps *productService) processExcelPicture(picture excelize.Picture, productCode string) (string, error) {
	dir := filepath.Join("uploads", "products")

	fmt.Printf("Dir: %s\n", dir)
	fmt.Printf("productCode: %s\n", productCode)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp(dir, fmt.Sprintf("%s_*.jpg", productCode))
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // 确保临时文件被删除
	defer tempFile.Close()

	// 将图片内容写入临时文件
	if _, err := tempFile.Write(picture.File); err != nil {
		return "", fmt.Errorf("error writing to temp file: %w", err)
	}

	// 因为imaging可以从文件路径读取，所以直接使用临时文件路径
	img, err := ps.imgUtil.Load(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("error loading image: %w", err)
	}

	// 调整图片大小
	img = ps.imgUtil.Thumbnail(img)

	filename := fmt.Sprintf("%s.jpg", productCode)
	filename = ps.imgUtil.WithUnixNanoTimestamp(filename)
	// 保存图片到临时文件
	err = ps.imgUtil.Save(img, filename, options.WithStorageDirOption(dir), options.WithQualityOption(80))
	if err != nil {
		return "", fmt.Errorf("error saving image: %w", err)
	}

	fullPath := filepath.Join(dir, filename)
	urlPath := filepath.ToSlash(fullPath)
	urlPath = fmt.Sprintf("/%s", urlPath)

	return urlPath, nil
}
