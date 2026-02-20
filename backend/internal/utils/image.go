package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/LouYuanbo1/go-webservice/imgutil"
	"github.com/LouYuanbo1/go-webservice/imgutil/options"
	"github.com/xuri/excelize/v2"
)

func IsImage(fileHeader *multipart.FileHeader) (bool, error) {
	fmt.Println("进入IsImage")

	file, err := fileHeader.Open()
	if err != nil {
		return false, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}

	contentType := http.DetectContentType(buffer)

	// 支持的图片MIME类型
	imageMimes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	fmt.Printf("Detected content type: %s\n", contentType)
	fmt.Println(imageMimes[contentType])

	return imageMimes[contentType], nil
}

func ProcessImageFileHeader(imgUtil imgutil.ImgUtil, fileHeader *multipart.FileHeader, dirElems []string, filename string, sortID int) (string, error) {
	isImage, err := IsImage(fileHeader)
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

	fmt.Printf("File size: %d\n", fileHeader.Size)

	dir := filepath.Join(dirElems...)

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp(dir, fmt.Sprintf("%s_*.jpg", filename))
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return "", fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // 确保临时文件被删除
	defer tempFile.Close()

	fmt.Printf("Temp file name: %s\n", tempFile.Name())

	// 将上传的文件内容复制到临时文件
	bytesCopied, err := io.Copy(tempFile, file)
	if err != nil {
		return "", fmt.Errorf("error copying file: %w", err)
	}

	fmt.Printf("Bytes copied: %d\n", bytesCopied)

	// 因为imaging可以从文件路径读取，所以直接使用临时文件路径
	img, err := imgUtil.Load(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("error loading image: %w", err)
	}

	// 调整图片大小
	img = imgUtil.Thumbnail(img)

	filename = fmt.Sprintf("%s_%d.jpg", filename, sortID)
	//filename = imgUtil.WithUnixNanoTimestamp(filename)
	// 保存图片到临时文件
	//在保存图片时，如果指定的文件名已经存在，会直接覆盖原有文件,这是我们想要的
	err = imgUtil.Save(img, filename, options.WithStorageDirOption(dir), options.WithQualityOption(80))
	if err != nil {
		return "", fmt.Errorf("error saving image: %w", err)
	}

	fullPath := filepath.Join(dir, filename)
	urlPath := filepath.ToSlash(fullPath)
	urlPath = fmt.Sprintf("/%s", urlPath)
	return urlPath, nil
}

func ProcessExcelPicture(imgUtil imgutil.ImgUtil, picture excelize.Picture, dirElems []string, filename string, sortID int) (string, error) {
	dir := filepath.Join(dirElems...)

	fmt.Printf("Dir: %s\n", dir)
	fmt.Printf("filename: %s\n", filename)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp(dir, fmt.Sprintf("%s_*.jpg", filename))
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
	img, err := imgUtil.Load(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("error loading image: %w", err)
	}

	// 调整图片大小
	img = imgUtil.Thumbnail(img)

	filename = fmt.Sprintf("%s_%d.jpg", filename, sortID)
	//filename = imgUtil.WithUnixNanoTimestamp(filename)
	// 保存图片到临时文件
	err = imgUtil.Save(img, filename, options.WithStorageDirOption(dir), options.WithQualityOption(80))
	if err != nil {
		return "", fmt.Errorf("error saving image: %w", err)
	}

	fullPath := filepath.Join(dir, filename)
	urlPath := filepath.ToSlash(fullPath)
	urlPath = fmt.Sprintf("/%s", urlPath)

	return urlPath, nil
}
