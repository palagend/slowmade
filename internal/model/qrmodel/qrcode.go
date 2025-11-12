package qrmodel

import (
	"time"
)

// QRCodeData 二维码生成数据模型
type QRCodeData struct {
	Content     string    `json:"content"`
	Size        int       `json:"size"`
	Format      string    `json:"format"`
	ErrorLevel  string    `json:"error_level"`
	Foreground  string    `json:"foreground"`
	Background  string    `json:"background"`
	OutputFile  string    `json:"output_file"`
	GeneratedAt time.Time `json:"generated_at"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
}

// ASCIIArtData ASCII艺术二维码数据模型
type ASCIIArtData struct {
	QRCodeData
	ArtContent string `json:"art_content"`
}

// ImageOutputData 图片输出数据模型
type ImageOutputData struct {
	QRCodeData
	FilePath   string `json:"file_path"`
	Dimensions string `json:"dimensions"`
}

// NewQRCodeData 创建二维码数据实例
func NewQRCodeData(content, format, errorLevel, foreground, background, output string, size int) *QRCodeData {
	return &QRCodeData{
		Content:     content,
		Size:        size,
		Format:      format,
		ErrorLevel:  errorLevel,
		Foreground:  foreground,
		Background:  background,
		OutputFile:  output,
		GeneratedAt: time.Now(),
		Success:     true,
	}
}
