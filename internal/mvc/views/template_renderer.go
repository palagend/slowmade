package views

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"text/template"
	"time"

	"github.com/palagend/slowmade/internal/config"
	"github.com/palagend/slowmade/internal/logging"
	"github.com/palagend/slowmade/internal/mvc/models"
)

type TemplateRenderer struct {
	config    *config.TemplateConfig
	templates map[string]*template.Template
	logger    *logging.Logger
}

// 自定义模板函数（移除颜色相关函数，使用ASCII字符）
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
}

func NewTemplateRenderer(cfg *config.TemplateConfig) *TemplateRenderer {
	//logger := log.New(os.Stderr, "[TemplateRenderer] ", log.LstdFlags|log.Lmsgprefix)
	logger := logging.Global().WithFields(map[string]interface{}{
		"component": "TemplateRenderer",
	})

	renderer := &TemplateRenderer{
		config:    cfg,
		templates: make(map[string]*template.Template),
		logger:    logger,
	}

	renderer.logger.Debug("初始化模板渲染器", map[string]interface{}{
		"custom_enabled": cfg.EnableCustom,
		"template_dir":   cfg.CustomTemplateDir,
	})

	if err := renderer.initializeTemplates(); err != nil {
		renderer.logger.Info(fmt.Sprintf("模板初始化失败: %v，使用默认模板", err))
		// 回退到默认模板
		renderer.initializeDefaultTemplates()
	}

	renderer.logger.Debug(fmt.Sprintf("初始化完成，加载模板数量: %d", len(renderer.templates)))
	for name := range renderer.templates {
		renderer.logger.Debug(fmt.Sprintf("  - 已加载模板: %s", name))
	}

	return renderer
}

func (r *TemplateRenderer) initializeTemplates() error {
	r.logger.Debug(fmt.Sprintf("开始初始化模板..."))

	// 如果启用自定义模板且路径存在，优先加载自定义模板
	if r.config.EnableCustom && r.config.CustomTemplateDir != "" {
		r.logger.Info(fmt.Sprintf("尝试加载自定义模板，目录: %s", r.config.CustomTemplateDir))

		// 检查目录是否存在
		if _, err := os.Stat(r.config.CustomTemplateDir); os.IsNotExist(err) {
			r.logger.Info(fmt.Sprintf("自定义模板目录不存在: %s", r.config.CustomTemplateDir))
			return err
		}

		if err := r.loadCustomTemplates(); err == nil {
			r.logger.Info(fmt.Sprintf("自定义模板加载成功"))
			return nil
		} else {
			r.logger.Info(fmt.Sprintf("自定义模板加载失败: %v", err))
		}
	} else {
		r.logger.Debug(fmt.Sprintf("自定义模板未启用或目录为空"))
	}

	r.logger.Debug(fmt.Sprintf("回退到默认模板"))
	// 回退到默认嵌入模板
	return r.initializeDefaultTemplates()
}

func (r *TemplateRenderer) loadCustomTemplates() error {
	r.logger.Debug(fmt.Sprintf("开始扫描自定义模板目录: %s", r.config.CustomTemplateDir))

	templateCount := 0
	err := filepath.WalkDir(r.config.CustomTemplateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			r.logger.Debug(fmt.Sprintf("遍历目录错误: %v", err))
			return err
		}

		if d.IsDir() {
			return nil // 跳过目录
		}

		if filepath.Ext(path) != ".tmpl" {
			r.logger.Debug(fmt.Sprintf("跳过非模板文件: %s", path))
			return nil
		}
		r.logger.Debug(fmt.Sprintf("加载模板文件: %s", path))
		content, err := os.ReadFile(path)
		if err != nil {
			r.logger.Debug(fmt.Sprintf("读取模板文件失败: %v", err))
			return err
		}

		tmplName := filepath.Base(path)
		tmpl, err := template.New(tmplName).Funcs(funcMap).Parse(string(content))
		if err != nil {
			r.logger.Debug(fmt.Sprintf("解析模板失败: %v", err))
			return err
		}

		r.templates[tmplName] = tmpl
		templateCount++
		r.logger.Debug(fmt.Sprintf("成功加载模板: %s", tmplName))

		return nil
	})

	if err != nil {
		r.logger.Debug(fmt.Sprintf("加载自定义模板失败: %v", err))
		return err
	}

	r.logger.Debug(fmt.Sprintf("自定义模板加载完成，共加载 %d 个模板", templateCount))
	return nil
}

func (r *TemplateRenderer) initializeDefaultTemplates() error {
	r.logger.Debug(fmt.Sprintf("开始初始化默认模板"))

	templateFiles := []string{
		"wallet_created.tmpl",
		"wallet_list.tmpl",
		"wallet_info.tmpl",
		"address_qr.tmpl",
		"transaction.tmpl",
	}

	successCount := 0
	for _, filename := range templateFiles {
		r.logger.Debug(fmt.Sprintf("加载默认模板: %s", filename))

		content, err := GetDefaultTemplate(filename)
		if err != nil {
			r.logger.Debug(fmt.Sprintf("获取默认模板内容失败: %v", err))
			continue
		}

		tmpl, err := template.New(filename).Funcs(funcMap).Parse(string(content))
		if err != nil {
			r.logger.Debug(fmt.Sprintf("解析默认模板失败: %v", err))
			continue
		}

		r.templates[filename] = tmpl
		successCount++
		r.logger.Debug(fmt.Sprintf("成功加载默认模板: %s", filename))
	}

	if successCount == 0 {
		r.logger.Debug(fmt.Sprintf("所有默认模板加载失败"))
		return fmt.Errorf("所有默认模板加载失败")
	}

	r.logger.Debug(fmt.Sprintf("默认模板初始化完成，成功加载 %d/%d 个模板",
		successCount, len(templateFiles)))
	return nil
}

// 具体的业务渲染方法
func (r *TemplateRenderer) RenderWalletCreated(wallet *models.VirtualWallet) (string, error) {
	r.logger.Debug(fmt.Sprintf("开始渲染钱包创建模板，钱包ID: %s", wallet.ID))

	tmplName := "wallet_created.tmpl"
	tmpl, exists := r.templates[tmplName]
	if !exists {
		err := fmt.Errorf("模板不存在: %s", tmplName)
		r.logger.Debug(fmt.Sprintf("错误: %v", err))
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, wallet); err != nil {
		r.logger.Debug(fmt.Sprintf("渲染钱包创建模板失败: %v", err))
		return "", err
	}

	result := buf.String()
	r.logger.Debug(fmt.Sprintf("钱包创建模板渲染成功，输出长度: %d 字符", len(result)))
	return result, nil
}

func (r *TemplateRenderer) RenderWalletList(wallets []*models.VirtualWallet) (string, error) {
	r.logger.Debug(fmt.Sprintf("开始渲染钱包列表模板，钱包数量: %d", len(wallets)))

	tmplName := "wallet_list.tmpl"
	tmpl, exists := r.templates[tmplName]
	if !exists {
		err := fmt.Errorf("模板不存在: %s", tmplName)
		r.logger.Debug(fmt.Sprintf("错误: %v", err))
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, wallets); err != nil {
		r.logger.Debug(fmt.Sprintf("渲染钱包列表模板失败: %v", err))
		return "", err
	}

	result := buf.String()
	r.logger.Debug(fmt.Sprintf("钱包列表模板渲染成功，输出长度: %d 字符", len(result)))
	return result, nil
}

func (r *TemplateRenderer) RenderWalletInfo(wallet *models.VirtualWallet) (string, error) {
	r.logger.Debug(fmt.Sprintf("开始渲染钱包信息模板，钱包ID: %s", wallet.ID))

	tmplName := "wallet_info.tmpl"
	tmpl, exists := r.templates[tmplName]
	if !exists {
		err := fmt.Errorf("模板不存在: %s", tmplName)
		r.logger.Debug(fmt.Sprintf("错误: %v", err))
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, wallet); err != nil {
		r.logger.Debug(fmt.Sprintf("渲染钱包信息模板失败: %v", err))
		return "", err
	}

	result := buf.String()
	r.logger.Debug(fmt.Sprintf("钱包信息模板渲染成功，输出长度: %d 字符", len(result)))
	return result, nil
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
	r.logger.Debug(fmt.Sprintf("模板状态报告:"))
	status := r.GetTemplateStatus()
	for name, loaded := range status {
		statusIcon := "[OK]"
		if !loaded {
			statusIcon = "[ERROR]"
		}
		r.logger.Debug(fmt.Sprintf("  %s %s: %v", statusIcon, name, loaded))
	}
	r.logger.Debug(fmt.Sprintf("总模板数: %d", len(r.templates)))
}
