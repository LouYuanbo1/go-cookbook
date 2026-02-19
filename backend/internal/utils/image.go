package utils

import (
	"fmt"
	"mime/multipart"
	"net/http"
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
