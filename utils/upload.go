package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backend/configs"
	"backend/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var allowedVideoTypes = map[string]bool{
	"video/mp4":  true,
	"video/avi":  true,
	"video/mov":  true,
	"video/webm": true,
}

// SaveUploadedFile 保存上传的文件
func SaveUploadedFile(c *gin.Context, file *multipart.FileHeader, subDir string) (*types.UploadResponse, error) {
	// 检查文件大小
	if file.Size > configs.AppConfig.Upload.MaxFileSize {
		return nil, fmt.Errorf("文件大小不能超过 %d MB", configs.AppConfig.Upload.MaxFileSize/(1024*1024))
	}

	// 检查文件类型
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 读取文件头来检测MIME类型
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return nil, err
	}

	// 重置文件指针
	src.Seek(0, io.SeekStart)

	// 检查MIME类型
	contentType := file.Header.Get("Content-Type")
	if !isAllowedFileType(contentType) {
		return nil, fmt.Errorf("不支持的文件类型: %s", contentType)
	}

	// 生成文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%s%s", 
		time.Now().Format("20060102_150405"), 
		uuid.New().String()[:8], 
		ext)

	// 创建保存路径
	uploadPath := filepath.Join(configs.AppConfig.Upload.Path, subDir)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return nil, err
	}

	// 完整的文件路径
	fullPath := filepath.Join(uploadPath, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		return nil, err
	}

	// 生成访问URL
	url := fmt.Sprintf("/uploads/%s/%s", subDir, filename)

	return &types.UploadResponse{
		URL:      url,
		Filename: filename,
		Size:     file.Size,
	}, nil
}

// SaveMultipleFiles 保存多个文件
func SaveMultipleFiles(c *gin.Context, files []*multipart.FileHeader, subDir string) ([]*types.UploadResponse, error) {
	var responses []*types.UploadResponse

	for _, file := range files {
		response, err := SaveUploadedFile(c, file, subDir)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// isAllowedFileType 检查是否为允许的文件类型
func isAllowedFileType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return allowedImageTypes[contentType] || allowedVideoTypes[contentType]
}

// IsImageFile 检查是否为图片文件
func IsImageFile(contentType string) bool {
	return allowedImageTypes[strings.ToLower(contentType)]
}

// IsVideoFile 检查是否为视频文件
func IsVideoFile(contentType string) bool {
	return allowedVideoTypes[strings.ToLower(contentType)]
}

// DeleteFile 删除文件
func DeleteFile(filePath string) error {
	// 构建完整路径
	fullPath := filepath.Join(configs.AppConfig.Upload.Path, filePath)
	
	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // 文件不存在，认为删除成功
	}

	// 删除文件
	return os.Remove(fullPath)
}

// GetFileURL 获取文件的完整URL
func GetFileURL(relativePath string) string {
	if relativePath == "" {
		return ""
	}
	
	// 如果已经是完整URL，直接返回
	if strings.HasPrefix(relativePath, "http://") || strings.HasPrefix(relativePath, "https://") {
		return relativePath
	}
	
	// 如果不以/uploads开头，添加前缀
	if !strings.HasPrefix(relativePath, "/uploads") {
		relativePath = "/uploads/" + strings.TrimPrefix(relativePath, "/")
	}
	
	return relativePath
}