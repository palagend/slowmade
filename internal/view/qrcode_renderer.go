package view

import (
	"embed"
	"log"
	"os"

	"github.com/palagend/cipherkey/internal/model/qrmodel"
)

//go:embed templates/qrcode
var qrmodelTemplateFs embed.FS

// RenderASCII 渲染ASCII艺术模板
func RenderASCII(data *qrmodel.ASCIIArtData) {
	engine := NewTemplateRenderer(false)
	if err := engine.LoadTemplates(qrmodelTemplateFs, "*.tmpl"); err != nil {
		log.Fatalf("渲染失败")
	}

	engine.Execute(os.Stdout, "ascii", data)
}

// RenderImage 渲染图片输出模板
func RenderImage(data *qrmodel.ImageOutputData) {
	engine := NewTemplateRenderer(false)
	if err := engine.LoadTemplates(qrmodelTemplateFs, "*.tmpl"); err != nil {
		log.Fatalf("渲染失败")
	}
	engine.Execute(os.Stdout, "image", data)
}
