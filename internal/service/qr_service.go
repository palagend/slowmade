package service

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"os"
	"strings"
	"time"

	"github.com/palagend/cipherkey/internal/model/qrmodel"
	"github.com/palagend/cipherkey/internal/view"
	"github.com/skip2/go-qrcode"
)

// QRCodeService 定义QR码生成服务接口
type QRCodeService interface {
	GenerateQRCode(content, format, errorLevel, foreground, background, output string, size int) error
	GetInputContent(args []string) (string, error)
}

type qrcodeService struct{}

// NewQRCodeService 创建QR码服务实例
func NewQRCodeService() QRCodeService {
	return &qrcodeService{}
}

// GetInputContent 从多种来源获取输入内容
func (s *qrcodeService) GetInputContent(args []string) (string, error) {
	// 优先级1: 从位置参数获取
	if len(args) > 0 {
		return args[0], nil
	}

	// 优先级2: 从标准输入管道获取
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		reader := bufio.NewReader(os.Stdin)
		var input strings.Builder

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					if line != "" {
						input.WriteString(line)
					}
					break
				}
				return "", err
			}
			input.WriteString(line)
		}
		return strings.TrimSpace(input.String()), nil
	}

	return "", nil
}

// GenerateQRCode 生成二维码的核心方法
func (s *qrcodeService) GenerateQRCode(content, format, errorLevel, foreground, background, output string, size int) error {

	switch strings.ToLower(format) {
	case "ascii":
		data, _ := s.generateASCIIQRCode(content, errorLevel)
		view.RenderASCII(data)
		return nil
	case "png", "jpg", "jpeg", "svg":
		data, _ := s.generateImageQRCode(content, format, errorLevel, foreground, background, output, size)
		view.RenderImage(data)
		return nil
	default:
		return fmt.Errorf("不支持的格式: %s", format)
	}
}

// getErrorCorrectionLevel 获取纠错级别[1](@ref)
func (s *qrcodeService) getErrorCorrectionLevel(level string) (qrcode.RecoveryLevel, error) {
	switch strings.ToLower(level) {
	case "low":
		return qrcode.Low, nil
	case "medium":
		return qrcode.Medium, nil
	case "high":
		return qrcode.High, nil
	case "highest":
		return qrcode.Highest, nil
	default:
		return qrcode.Medium, fmt.Errorf("无效的纠错级别，使用默认值: medium")
	}
}

// generateASCIIQRCode 生成ASCII艺术二维码
func (s *qrcodeService) generateASCIIQRCode(content string, errorLevel string) (*qrmodel.ASCIIArtData, error) {
	level, _ := s.getErrorCorrectionLevel(errorLevel)
	if content == "" {
		return &qrmodel.ASCIIArtData{
			QRCodeData: qrmodel.QRCodeData{
				Content:     content,
				Success:     false,
				Message:     "内容不能为空",
				GeneratedAt: time.Now(),
			}}, fmt.Errorf("内容不能为空")
	}

	// 创建QR码对象
	qr, err := qrcode.New(content, level)
	if err != nil {
		return &qrmodel.ASCIIArtData{
			QRCodeData: qrmodel.QRCodeData{
				Content:     content,
				Success:     false,
				Message:     fmt.Sprintf("创建QR码失败: %v", err),
				GeneratedAt: time.Now(),
			}}, err
	}

	// 生成ASCII艺术二维码
	asciiArt := qr.ToSmallString(false)

	// 构建完整的数据模型
	return &qrmodel.ASCIIArtData{
		QRCodeData: qrmodel.QRCodeData{
			Content:     content,
			Size:        len(asciiArt), // 使用ASCII艺术长度作为尺寸参考
			Format:      "ASCII",
			ErrorLevel:  errorLevel,
			GeneratedAt: time.Now(),
			Success:     true,
			Message:     "ASCII二维码生成成功",
		},
		ArtContent: asciiArt,
	}, nil
}

// GenerateQRCodeImage 生成图片格式的二维码
func (s *qrcodeService) generateImageQRCode(content, format, errorLevel, foreground, background, output string, size int) (*qrmodel.ImageOutputData, error) {
	if content == "" {
		return nil, fmt.Errorf("内容不能为空")
	}

	level, err := s.getErrorCorrectionLevel(errorLevel)
	if err != nil {
		level = qrcode.Medium
	}

	// 转换颜色
	fgColor, err := parseColor(foreground)
	if err != nil {
		fgColor = color.Black
	}

	bgColor, err := parseColor(background)
	if err != nil {
		bgColor = color.White
	}

	// 生成二维码
	qr, err := qrcode.New(content, level)
	if err != nil {
		return &qrmodel.ImageOutputData{
			QRCodeData: qrmodel.QRCodeData{
				Content:     content,
				Success:     false,
				Message:     fmt.Sprintf("生成二维码失败: %v", err),
				GeneratedAt: time.Now(),
			},
		}, err
	}

	// 设置颜色
	qr.ForegroundColor = fgColor
	qr.BackgroundColor = bgColor

	var filePath string
	if output == "" {
		filePath = fmt.Sprintf("qrcode_%d.%s", time.Now().Unix(), strings.ToLower(format))
	} else {
		filePath = output
	}

	// 根据格式保存图片
	switch strings.ToUpper(format) {
	case "PNG":
		err = qr.WriteFile(size, filePath)
	case "JPEG", "JPG":
		// 对于JPEG格式，需要特殊处理
		err = s.saveAsJPEG(qr, size, filePath)
	default:
		// 默认保存为PNG
		err = qr.WriteFile(size, filePath)
		format = "PNG"
	}

	if err != nil {
		return &qrmodel.ImageOutputData{
			QRCodeData: qrmodel.QRCodeData{
				Content:     content,
				Format:      format,
				ErrorLevel:  errorLevel,
				Success:     false,
				Message:     fmt.Sprintf("保存图片失败: %v", err),
				GeneratedAt: time.Now(),
			},
		}, err
	}

	imageData := &qrmodel.ImageOutputData{
		QRCodeData: qrmodel.QRCodeData{
			Content:     content,
			Size:        size,
			Format:      format,
			ErrorLevel:  errorLevel,
			Foreground:  foreground,
			Background:  background,
			OutputFile:  filePath,
			GeneratedAt: time.Now(),
			Success:     true,
			Message:     "二维码图片生成成功",
		},
		FilePath:   filePath,
		Dimensions: fmt.Sprintf("%dx%d", size, size),
	}

	return imageData, nil
}

// 辅助函数

// parseColor 解析颜色字符串
func parseColor(colorStr string) (color.Color, error) {
	if colorStr == "" {
		return color.Black, nil
	}

	// 简单的颜色解析，支持常见颜色名称和hex
	switch strings.ToLower(colorStr) {
	case "black":
		return color.Black, nil
	case "white":
		return color.White, nil
	case "red":
		return color.RGBA{R: 255, G: 0, B: 0, A: 255}, nil
	case "green":
		return color.RGBA{R: 0, G: 255, B: 0, A: 255}, nil
	case "blue":
		return color.RGBA{R: 0, G: 0, B: 255, A: 255}, nil
	default:
		// 可以扩展支持hex颜色解析
		return color.Black, nil
	}
}

// saveAsJPEG 将二维码保存为JPEG格式
func (s *qrcodeService) saveAsJPEG(qr *qrcode.QRCode, size int, filename string) error {
	// 这里需要实现JPEG保存逻辑
	// 由于go-qrcode主要支持PNG，可能需要使用其他图像处理库
	// 作为示例，我们暂时也保存为PNG
	return qr.WriteFile(size, strings.TrimSuffix(filename, ".jpg")+".png")
}
