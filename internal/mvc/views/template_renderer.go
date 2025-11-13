package views

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"text/template"
	"time"

	"github.com/palagend/slowmade/internal/config"
	"github.com/palagend/slowmade/internal/mvc/models"
)

type TemplateRenderer struct {
	config    *config.TemplateConfig
	templates map[string]*template.Template
	debug     bool // 添加调试开关
}

// 自定以模板函数
var funcMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"last": func(index int, slice interface{}) bool {
		v := reflect.ValueOf(slice)
		return index == v.Len()-1
	},
	"truncate": func(s string, length int) string {
		if len(s) <= length {
			return s
		}
		return s[:length]
	},
	"formatTimestamp": func(ts int64) string {
		if ts == 0 {
			return "未知时间"
		}
		return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
	},
	"color": func(color, text string) string {
		// 简单的颜色支持，实际使用时可以集成更复杂的颜色库
		colors := map[string]string{
			"red":     "\033[31m",
			"green":   "\033[32m",
			"yellow":  "\033[33m",
			"blue":    "\033[34m",
			"magenta": "\033[35m",
			"cyan":    "\033[36m",
			"white":   "\033[37m",
			"bold":    "\033[1m",
			"reset":   "\033[0m",
		}
		if code, exists := colors[color]; exists {
			return code + text + "\033[0m"
		}
		return text
	},
}

func NewTemplateRenderer(cfg *config.TemplateConfig) *TemplateRenderer {
	renderer := &TemplateRenderer{
		config:    cfg,
		templates: make(map[string]*template.Template),
		debug:     true, // 默认开启调试，可以通过配置控制
	}

	log.Printf("🔧 [TemplateRenderer] 初始化模板渲染器，自定义模板启用: %v, 模板目录: %s",
		cfg.EnableCustom, cfg.CustomTemplateDir)

	if err := renderer.initializeTemplates(); err != nil {
		log.Printf("⚠️  [TemplateRenderer] 模板初始化失败: %v，使用默认模板", err)
		// 回退到默认模板
		renderer.initializeDefaultTemplates()
	}

	log.Printf("✅ [TemplateRenderer] 初始化完成，加载模板数量: %d", len(renderer.templates))
	for name := range renderer.templates {
		log.Printf("   - 已加载模板: %s", name)
	}

	return renderer
}

func (r *TemplateRenderer) initializeTemplates() error {
	log.Printf("🔧 [TemplateRenderer] 开始初始化模板...")

	// 如果启用自定义模板且路径存在，优先加载自定义模板
	if r.config.EnableCustom && r.config.CustomTemplateDir != "" {
		log.Printf("📁 [TemplateRenderer] 尝试加载自定义模板，目录: %s", r.config.CustomTemplateDir)

		// 检查目录是否存在
		if _, err := os.Stat(r.config.CustomTemplateDir); os.IsNotExist(err) {
			log.Printf("❌ [TemplateRenderer] 自定义模板目录不存在: %s", r.config.CustomTemplateDir)
			return err
		}

		if err := r.loadCustomTemplates(); err == nil {
			log.Printf("✅ [TemplateRenderer] 自定义模板加载成功")
			return nil
		} else {
			log.Printf("❌ [TemplateRenderer] 自定义模板加载失败: %v", err)
		}
	} else {
		log.Printf("ℹ️  [TemplateRenderer] 自定义模板未启用或目录为空")
	}

	log.Printf("🔧 [TemplateRenderer] 回退到默认模板")
	// 回退到默认嵌入模板
	return r.initializeDefaultTemplates()
}

func (r *TemplateRenderer) loadCustomTemplates() error {
	log.Printf("📂 [TemplateRenderer] 开始扫描自定义模板目录: %s", r.config.CustomTemplateDir)

	templateCount := 0
	err := filepath.WalkDir(r.config.CustomTemplateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("❌ [TemplateRenderer] 遍历目录错误: %v", err)
			return err
		}

		if d.IsDir() {
			return nil // 跳过目录
		}

		if filepath.Ext(path) != ".tmpl" {
			if r.debug {
				log.Printf("ℹ️  [TemplateRenderer] 跳过非模板文件: %s", path)
			}
			return nil
		}

		log.Printf("📄 [TemplateRenderer] 加载模板文件: %s", path)

		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("❌ [TemplateRenderer] 读取模板文件失败: %v", err)
			return err
		}

		tmplName := filepath.Base(path)
		tmpl, err := template.New(tmplName).Funcs(funcMap).Parse(string(content))
		if err != nil {
			log.Printf("❌ [TemplateRenderer] 解析模板失败: %v", err)
			return err
		}

		r.templates[tmplName] = tmpl
		templateCount++
		log.Printf("✅ [TemplateRenderer] 成功加载模板: %s", tmplName)

		return nil
	})

	if err != nil {
		log.Printf("❌ [TemplateRenderer] 加载自定义模板失败: %v", err)
		return err
	}

	log.Printf("✅ [TemplateRenderer] 自定义模板加载完成，共加载 %d 个模板", templateCount)
	return nil
}

func (r *TemplateRenderer) initializeDefaultTemplates() error {
	log.Printf("🔧 [TemplateRenderer] 开始初始化默认模板")

	templateFiles := []string{
		"wallet_created.tmpl",
		"wallet_list.tmpl",
		"wallet_info.tmpl",
		"address_qr.tmpl",
		"transaction.tmpl",
	}

	successCount := 0
	for _, filename := range templateFiles {
		log.Printf("📄 [TemplateRenderer] 加载默认模板: %s", filename)

		content, err := GetDefaultTemplate(filename)
		if err != nil {
			log.Printf("❌ [TemplateRenderer] 获取默认模板内容失败: %v", err)
			continue
		}

		tmpl, err := template.New(filename).Funcs(funcMap).Parse(string(content))
		if err != nil {
			log.Printf("❌ [TemplateRenderer] 解析默认模板失败: %v", err)
			continue
		}

		r.templates[filename] = tmpl
		successCount++
		log.Printf("✅ [TemplateRenderer] 成功加载默认模板: %s", filename)
	}

	if successCount == 0 {
		log.Printf("❌ [TemplateRenderer] 所有默认模板加载失败")
		return fmt.Errorf("所有默认模板加载失败")
	}

	log.Printf("✅ [TemplateRenderer] 默认模板初始化完成，成功加载 %d/%d 个模板",
		successCount, len(templateFiles))
	return nil
}

// 具体的业务渲染方法
func (r *TemplateRenderer) RenderWalletCreated(wallet *models.VirtualWallet) (string, error) {
	if r.debug {
		log.Printf("🎨 [TemplateRenderer] 开始渲染钱包创建模板，钱包ID: %s", wallet.ID)
	}

	tmplName := "wallet_created.tmpl"
	tmpl, exists := r.templates[tmplName]
	if !exists {
		err := fmt.Errorf("模板不存在: %s", tmplName)
		log.Printf("❌ [TemplateRenderer] %v", err)
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, wallet); err != nil {
		log.Printf("❌ [TemplateRenderer] 渲染钱包创建模板失败: %v", err)
		return "", err
	}

	result := buf.String()
	if r.debug {
		log.Printf("✅ [TemplateRenderer] 钱包创建模板渲染成功，输出长度: %d 字符", len(result))
	}
	return result, nil
}

func (r *TemplateRenderer) RenderWalletList(wallets []*models.VirtualWallet) (string, error) {
	if r.debug {
		log.Printf("🎨 [TemplateRenderer] 开始渲染钱包列表模板，钱包数量: %d", len(wallets))
	}

	tmplName := "wallet_list.tmpl"
	tmpl, exists := r.templates[tmplName]
	if !exists {
		err := fmt.Errorf("模板不存在: %s", tmplName)
		log.Printf("❌ [TemplateRenderer] %v", err)
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, wallets); err != nil {
		log.Printf("❌ [TemplateRenderer] 渲染钱包列表模板失败: %v", err)
		return "", err
	}

	result := buf.String()
	if r.debug {
		log.Printf("✅ [TemplateRenderer] 钱包列表模板渲染成功，输出长度: %d 字符", len(result))
	}
	return result, nil
}

func (r *TemplateRenderer) RenderWalletInfo(wallet *models.VirtualWallet) (string, error) {
	if r.debug {
		log.Printf("🎨 [TemplateRenderer] 开始渲染钱包信息模板，钱包ID: %s", wallet.ID)
	}

	tmplName := "wallet_info.tmpl"
	tmpl, exists := r.templates[tmplName]
	if !exists {
		err := fmt.Errorf("模板不存在: %s", tmplName)
		log.Printf("❌ [TemplateRenderer] %v", err)
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, wallet); err != nil {
		log.Printf("❌ [TemplateRenderer] 渲染钱包信息模板失败: %v", err)
		return "", err
	}

	result := buf.String()
	if r.debug {
		log.Printf("✅ [TemplateRenderer] 钱包信息模板渲染成功，输出长度: %d 字符", len(result))
	}
	return result, nil
}

// 添加调试控制方法
func (r *TemplateRenderer) SetDebug(debug bool) {
	r.debug = debug
	log.Printf("🔧 [TemplateRenderer] 调试模式: %v", debug)
}

// 添加模板状态检查方法
func (r *TemplateRenderer) GetTemplateStatus() map[string]bool {
	status := make(map[string]bool)
	expectedTemplates := []string{
		"wallet_created.tmpl",
		"wallet_list.tmpl",
		"wallet_info.tmpl",
		"address_qr.tmpl",
		"transaction.tmpl",
	}

	for _, tmpl := range expectedTemplates {
		status[tmpl] = r.templates[tmpl] != nil
	}

	return status
}

// 打印模板状态信息
func (r *TemplateRenderer) PrintStatus() {
	log.Printf("📊 [TemplateRenderer] 模板状态报告:")
	status := r.GetTemplateStatus()
	for name, loaded := range status {
		statusIcon := "✅"
		if !loaded {
			statusIcon = "❌"
		}
		log.Printf("   %s %s: %v", statusIcon, name, loaded)
	}
	log.Printf("   总模板数: %d", len(r.templates))
}
