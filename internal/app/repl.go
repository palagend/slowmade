package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/palagend/slowmade/pkg/logging"
	"github.com/peterh/liner"
	"go.uber.org/zap"
)

// REPL 表示一个交互式读取-求值-打印循环环境
type REPL struct {
	line     *liner.State
	history  *HistoryManager
	running  bool
	commands map[string]CommandHandler
	logger   *zap.Logger
}

// CommandHandler 定义命令处理函数类型
type CommandHandler func(args []string) error

// NewREPL 创建并初始化一个新的 REPL 实例
func NewREPL() (*REPL, error) {
	line := liner.NewLiner()
	line.SetCtrlCAborts(true)
	line.SetTabCompletionStyle(liner.TabCircular)

	// 设置基本的 Tab 补全
	line.SetCompleter(func(line string) []string {
		return []string{"exit", "help", "clear", "history", "version"}
	})

	repl := &REPL{
		line:     line,
		history:  NewHistoryManager(),
		running:  true,
		logger:   logging.Get(),
		commands: make(map[string]CommandHandler),
	}

	// 注册内置命令
	repl.registerCommands()

	return repl, nil
}

// registerCommands 注册所有内置命令
func (r *REPL) registerCommands() {
	r.commands = map[string]CommandHandler{
		"exit":    r.handleExit,
		"quit":    r.handleExit,
		"help":    r.handleHelp,
		"clear":   r.handleClear,
		"history": r.handleHistory,
		"version": r.handleVersion,
	}
}

// Run 启动 REPL 主循环
func (r *REPL) Run() {
	defer r.Close()

	r.printWelcome()

	for r.running {
		input, err := r.readInput()
		if err != nil {
			if err == liner.ErrPromptAborted || err == io.EOF {
				fmt.Println("\nGoodbye!")
				break
			}
			r.logger.Error("Error reading input", zap.Error(err))
			continue
		}

		if input == "" {
			continue
		}

		// 添加到历史记录
		r.history.Add(input)
		if err := r.history.Save(); err != nil {
			r.logger.Warn("Failed to save history", zap.Error(err))
		}

		// 处理输入
		if err := r.processInput(input); err != nil {
			if err == ErrExitRequested {
				break
			}
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// readInput 读取用户输入，支持多行输入
func (r *REPL) readInput() (string, error) {
	prompt := "slowmade> "
	var input strings.Builder
	var isMultiline bool

	for {
		var line string
		var err error

		if isMultiline {
			line, err = r.line.Prompt("... ")
		} else {
			line, err = r.line.Prompt(prompt)
		}

		if err != nil {
			return "", err
		}

		// 检查是否是多行输入（以 \ 结尾）
		if strings.HasSuffix(line, "\\") {
			input.WriteString(strings.TrimSuffix(line, "\\"))
			input.WriteString(" ")
			isMultiline = true
			continue
		}

		input.WriteString(line)
		break
	}

	return strings.TrimSpace(input.String()), nil
}

// processInput 处理用户输入
func (r *REPL) processInput(input string) error {
	// 去除前后空格
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	// 分割命令和参数
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	// 检查是否是内置命令
	if handler, exists := r.commands[command]; exists {
		return handler(args)
	}

	// 如果不是内置命令，作为表达式求值
	return r.handleExpression(input)
}

// handleExpression 处理表达式求值
func (r *REPL) handleExpression(expr string) error {
	// 这里可以实现您的业务逻辑表达式求值
	// 目前先简单回显
	result, err := r.evaluateExpression(expr)
	if err != nil {
		return err
	}

	if result != "" {
		fmt.Println(result)
	}
	return nil
}

// evaluateExpression 表达式求值核心逻辑
func (r *REPL) evaluateExpression(expr string) (string, error) {
	// 这里可以集成 gomacro 或其他求值引擎
	// 目前先实现一个简单的示例

	// 检查是否是简单的数学表达式
	if strings.ContainsAny(expr, "+-*/") {
		// 简单的数学表达式求值（示例）
		// 在实际项目中，您应该使用更安全的求值方式
		return fmt.Sprintf("Expression: %s (evaluation would go here)", expr), nil
	}

	// 默认处理
	return fmt.Sprintf("Executed: %s", expr), nil
}

// 内置命令处理函数
func (r *REPL) handleExit(args []string) error {
	r.running = false
	fmt.Println("Goodbye!")
	return ErrExitRequested
}

func (r *REPL) handleHelp(args []string) error {
	helpText := `
Available commands:
  exit, quit    - Exit the REPL
  help          - Show this help message
  clear         - Clear the screen
  history       - Show command history
  version       - Show version information

Expressions:
  Enter any expression to evaluate it
  Use \ at the end of a line for multi-line input

Shortcuts:
  Ctrl+D        - Exit the REPL
  Ctrl+C        - Abort current input
  Up/Down arrows - Navigate command history
  Tab           - Auto-completion
`
	fmt.Println(helpText)
	return nil
}

func (r *REPL) handleClear(args []string) error {
	fmt.Print("\033[H\033[2J") // ANSI escape code to clear screen
	return nil
}

func (r *REPL) handleHistory(args []string) error {
	history := r.history.GetLast(20)
	for i, cmd := range history {
		fmt.Printf("%4d  %s\n", i+1, cmd)
	}
	return nil
}

func (r *REPL) handleVersion(args []string) error {
	fmt.Println("Slowmade REPL v1.0.0")
	return nil
}

// printWelcome 显示欢迎信息
func (r *REPL) printWelcome() {
	welcome := `
Welcome to Slowmade REPL!
Type 'help' for available commands, 'exit' to quit.
`
	fmt.Println(welcome)
}

// Close 清理资源
func (r *REPL) Close() {
	if r.line != nil {
		if err := r.history.Save(); err != nil {
			r.logger.Warn("Failed to save history on close", zap.Error(err))
		}
		r.line.Close()
	}
}
