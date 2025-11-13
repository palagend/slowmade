package views

import "embed"

//go:embed templates/*.tmpl
var defaultTemplates embed.FS

// GetDefaultTemplate 读取嵌入的默认模板内容
func GetDefaultTemplate(name string) ([]byte, error) {
    return defaultTemplates.ReadFile("templates/" + name)
}
