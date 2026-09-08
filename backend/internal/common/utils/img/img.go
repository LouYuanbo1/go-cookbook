package img

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/xuri/excelize/v2"
)

type ImgUtil interface {
	Load(imgPath string) (image.Image, error)
	Thumbnail(img image.Image, width, height int, filter imaging.ResampleFilter) image.Image
	Save(img image.Image, path string, quality int) error
	ProcessFileHeader(fileHeader *multipart.FileHeader, dir string, filename string) error
	ProcessExcelPicture(picture excelize.Picture, dir string, filename string, sortID int) (string, error)
}

type imgUtil struct {
	config Config
}

func NewImgUtil(config Config) ImgUtil {
	return &imgUtil{
		config: config,
	}
}

// 加载图片
func (i *imgUtil) Load(imgPath string) (image.Image, error) {
	img, err := imaging.Open(imgPath)
	if err != nil {
		return nil, fmt.Errorf("load image failed: %w", err)
	}
	return img, nil
}

func (i *imgUtil) Thumbnail(img image.Image, width, height int, filter imaging.ResampleFilter) image.Image {
	return imaging.Thumbnail(img, width, height, filter)
}

// 保存图片,按照配置的质量保存
func (i *imgUtil) Save(img image.Image, path string, quality int) error {
	ext := strings.ToLower(filepath.Ext(path))
	var err error
	switch ext {
	case ".jpg", ".jpeg":
		err = imaging.Save(img, path, imaging.JPEGQuality(quality))
	case ".png":
		level := quality * 9 / 100
		level = max(level, 1)
		level = min(level, 9)
		err = imaging.Save(img, path, imaging.PNGCompressionLevel(png.CompressionLevel(level)))
	default:
		err = imaging.Save(img, path)
	}
	if err != nil {
		return fmt.Errorf("save image failed: %w", err)
	}
	return nil
}

func (i *imgUtil) isImage(fileHeader *multipart.FileHeader) (isImage bool, err error) {
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

func (i *imgUtil) ProcessFileHeader(
	fileHeader *multipart.FileHeader,
	dir string,
	filename string,
) error {
	/*
		isImage, err := i.isImage(fileHeader)
		if err != nil {
			return "", fmt.Errorf("error checking image MIME type: %w", err)
		}
		if !isImage {
			return "", fmt.Errorf("file is not an image")
		}
	*/

	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	ext := filepath.Ext(filename)
	prefix := strings.TrimSuffix(filename, ext)

	// 创建临时文件
	tempFile, err := os.CreateTemp(dir, fmt.Sprintf("%s_*%s", prefix, ext))
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // 确保临时文件被删除
	defer tempFile.Close()

	fmt.Printf("Temp file name: %s\n", tempFile.Name())

	// 将上传的文件内容复制到临时文件
	bytesCopied, err := io.Copy(tempFile, file)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	fmt.Printf("Bytes copied: %d\n", bytesCopied)

	// 因为imaging可以从文件路径读取，所以直接使用临时文件路径
	img, err := i.Load(tempFile.Name())
	if err != nil {
		return fmt.Errorf("error loading image: %w", err)
	}

	// 调整图片大小
	img = i.Thumbnail(img, i.config.Transform.Width, i.config.Transform.Height, imaging.Lanczos)

	// 使用带扩展名的文件名保存
	savePath := filepath.Join(dir, filename)

	fmt.Printf("savePath: %s\n", savePath)

	err = i.Save(img, savePath, i.config.Save.Quality)
	if err != nil {
		return fmt.Errorf("error saving image: %w", err)
	}
	return nil
}

func (i *imgUtil) ProcessExcelPicture(picture excelize.Picture, dir, filename string, sortID int) (string, error) {
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
	img, err := i.Load(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("error loading image: %w", err)
	}

	// 调整图片大小
	img = i.Thumbnail(img, i.config.Transform.Width, i.config.Transform.Height, imaging.Lanczos)

	filename = fmt.Sprintf("%s_%d.jpg", filename, sortID)
	// 保存图片到临时文件
	err = i.Save(img, filepath.Join(dir, filename), i.config.Save.Quality)
	if err != nil {
		return "", fmt.Errorf("error saving image: %w", err)
	}

	fullPath := filepath.Join(dir, filename)
	urlPath := filepath.ToSlash(fullPath)
	urlPath = fmt.Sprintf("/%s", urlPath)

	return urlPath, nil
}
