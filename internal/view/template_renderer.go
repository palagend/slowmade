package view

import (
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
)

// TemplateRenderer 通用模板渲染器接口
type TemplateRenderer interface {
	// LoadTemplates 从文件系统加载模板
	LoadTemplates(templateFS fs.FS, patterns ...string) error

	// Execute 渲染指定名称的模板
	Execute(w io.Writer, name string, data interface{}) error

	// ExecuteTemplate 渲染模板（支持布局和部分模板）
	ExecuteTemplate(w io.Writer, name string, data interface{}, layouts ...string) error

	// AddFunc 添加自定义模板函数
	AddFunc(name string, fn interface{}) TemplateRenderer

	// AddFuncMap 批量添加模板函数
	AddFuncMap(funcMap template.FuncMap) TemplateRenderer

	// Reload 重新加载模板（开发模式使用）
	Reload() error
}

type templateRenderer struct {
	templates  *template.Template
	once       sync.Once
	funcMap    template.FuncMap
	mu         sync.RWMutex
	templateFS fs.FS
	patterns   []string
	reloadable bool // 是否支持热重载（开发模式）
}

// NewTemplateRenderer 创建模板渲染器实例
func NewTemplateRenderer(reloadable bool) TemplateRenderer {
	return &templateRenderer{
		funcMap:    make(template.FuncMap),
		reloadable: reloadable,
	}
}

// 初始化默认模板函数
func (tr *templateRenderer) initDefaultFuncs() {
	defaultFuncs := template.FuncMap{
		"formatSize": func(size int) string {
			if size < 1024 {
				return string(rune(size)) + " B"
			} else if size < 1024*1024 {
				return string(rune(size/1024)) + " KB"
			}
			return string(rune(size/(1024*1024))) + " MB"
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatDateISO": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"trim":  strings.TrimSpace,
		"join":  strings.Join,
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"neq": func(a, b interface{}) bool {
			return a != b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"gte": func(a, b int) bool {
			return a >= b
		},
		"lt": func(a, b int) bool {
			return a < b
		},
		"lte": func(a, b int) bool {
			return a <= b
		},
		"dict": func(values ...interface{}) map[string]interface{} {
			result := make(map[string]interface{})
			if len(values)%2 != 0 {
				return result
			}
			for i := 0; i < len(values); i += 2 {
				if key, ok := values[i].(string); ok {
					result[key] = values[i+1]
				}
			}
			return result
		},
	}

	for k, v := range defaultFuncs {
		if _, exists := tr.funcMap[k]; !exists {
			tr.funcMap[k] = v
		}
	}
}

// LoadTemplates 从文件系统加载模板文件
func (tr *templateRenderer) LoadTemplates(templateFS fs.FS, patterns ...string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	tr.templateFS = templateFS
	tr.patterns = patterns

	// 初始化默认函数
	tr.initDefaultFuncs()

	// 创建根模板
	tmpl := template.New("root").Funcs(tr.funcMap)

	// 解析所有匹配的模板文件
	for _, pattern := range patterns {
		// 获取匹配的文件列表
		matches, err := fs.Glob(templateFS, pattern)
		if err != nil {
			return err
		}

		// 解析每个模板文件
		for _, filePath := range matches {
			// 跳过目录
			info, err := fs.Stat(templateFS, filePath)
			if err != nil || info.IsDir() {
				continue
			}

			// 读取文件内容
			content, err := fs.ReadFile(templateFS, filePath)
			if err != nil {
				return err
			}

			// 使用文件名（不含扩展名）作为模板名
			name := filepath.Base(filePath)
			name = strings.TrimSuffix(name, filepath.Ext(name))

			// 解析模板内容
			var currentTmpl *template.Template
			if name == tmpl.Name() {
				currentTmpl = tmpl
			} else {
				currentTmpl = tmpl.New(name)
			}

			_, err = currentTmpl.Parse(string(content))
			if err != nil {
				return err
			}
		}
	}

	tr.templates = tmpl
	return nil
}

// Execute 渲染指定名称的模板
func (tr *templateRenderer) Execute(w io.Writer, name string, data interface{}) error {

	if tr.templates == nil {
		log.Fatalf("没有模板\n")
	}

	return tr.templates.ExecuteTemplate(w, name, data)
}

// ExecuteTemplate 支持布局模板的渲染
func (tr *templateRenderer) ExecuteTemplate(w io.Writer, name string, data interface{}, layouts ...string) error {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	if tr.templates == nil {
		if tr.templates == nil {
			log.Fatalf("没有模板\n")
		}
	}

	// 如果没有指定布局，直接渲染模板
	if len(layouts) == 0 {
		return tr.templates.ExecuteTemplate(w, name, data)
	}

	// 克隆模板以避免修改原始模板
	clonedTmpl, err := tr.templates.Clone()
	if err != nil {
		return err
	}

	// 添加布局模板
	for _, layout := range layouts {
		layoutTmpl := clonedTmpl.Lookup(layout)
		if layoutTmpl == nil {
			continue
		}

		// 在布局模板中定义内容块
		_, err = clonedTmpl.New("content").Parse("{{template \"" + name + "\" .}}")
		if err != nil {
			return err
		}

		return clonedTmpl.ExecuteTemplate(w, layout, data)
	}

	// 回退到直接渲染
	return tr.templates.ExecuteTemplate(w, name, data)
}

// AddFunc 添加自定义模板函数[4](@ref)
func (tr *templateRenderer) AddFunc(name string, fn interface{}) TemplateRenderer {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	tr.funcMap[name] = fn

	// 如果模板已加载，重新加载以应用新函数
	if tr.templates != nil && tr.reloadable {
		tr.reloadTemplates()
	}

	return tr
}

// AddFuncMap 批量添加模板函数[4](@ref)
func (tr *templateRenderer) AddFuncMap(funcMap template.FuncMap) TemplateRenderer {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	for k, v := range funcMap {
		tr.funcMap[k] = v
	}

	// 如果模板已加载，重新加载以应用新函数
	if tr.templates != nil && tr.reloadable {
		tr.reloadTemplates()
	}

	return tr
}

// Reload 重新加载模板（开发模式使用）
func (tr *templateRenderer) Reload() error {
	if !tr.reloadable {
		return nil
	}

	return tr.reloadTemplates()
}

// reloadTemplates 内部重新加载实现
func (tr *templateRenderer) reloadTemplates() error {
	if tr.templateFS == nil || len(tr.patterns) == 0 {
		return nil
	}

	return tr.LoadTemplates(tr.templateFS, tr.patterns...)
}
