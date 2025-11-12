package cmd

import (
	"fmt"
	"os"

	"github.com/palagend/cipherkey/internal/service" // 根据实际模块路径修改

	"github.com/spf13/cobra"
)

var (
	qrSize       int
	qrOutput     string
	qrFormat     string
	qrContent    string
	qrErrorLevel string
	qrForeground string
	qrBackground string
)

var qrcodeCmd = &cobra.Command{
	Use:   "qrcode [content]",
	Short: "QR码生成工具",
	Long:  "为地址、支付请求和其他加密数据生成QR码。内容可以通过参数、标准输入或--text标志提供。",
	Run:   runQRCodeGenerator,
	Args:  cobra.MaximumNArgs(1),
}

func init() {
	qrcodeCmd.Flags().IntVarP(&qrSize, "size", "s", 256, "QR码像素尺寸")
	qrcodeCmd.Flags().StringVarP(&qrOutput, "output", "o", "", "输出文件名（图片格式）")
	qrcodeCmd.Flags().StringVarP(&qrFormat, "format", "f", "png", "输出格式（ascii, png, jpg, svg）")
	qrcodeCmd.Flags().StringVarP(&qrContent, "text", "t", "", "要编码的文本内容（参数或标准输入的替代方式）")
	qrcodeCmd.Flags().StringVarP(&qrErrorLevel, "error-level", "e", "medium", "纠错级别（low, medium, high, highest）")
	qrcodeCmd.Flags().StringVarP(&qrForeground, "foreground", "", "000000", "前景色十六进制（默认：黑色）")
	qrcodeCmd.Flags().StringVarP(&qrBackground, "background", "", "FFFFFF", "背景色十六进制（默认：白色）")
}

func runQRCodeGenerator(cmd *cobra.Command, args []string) {
	// 创建服务实例
	qrService := service.NewQRCodeService()

	// 获取输入内容（逻辑不变）
	var content string
	var err error

	if qrContent != "" {
		content = qrContent
	} else {
		content, err = qrService.GetInputContent(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取输入错误: %v\n", err)
			os.Exit(1)
		}
	}

	if content == "" {
		fmt.Fprintln(os.Stderr, "错误: 未提供QR码生成内容")
		os.Exit(1)
	}

	// 生成QR码并获取渲染后的结果
	qrService.GenerateQRCode(content, qrFormat, qrErrorLevel, qrForeground, qrBackground, qrOutput, qrSize)
}
