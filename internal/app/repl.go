// internal/app/repl.go
package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/palagend/slowmade/internal/core"
	"github.com/palagend/slowmade/internal/security"
	"github.com/palagend/slowmade/internal/view"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/peterh/liner"
	"go.uber.org/zap"
)

// REPL 表示一个交互式读取-求值-打印循环环境
type REPL struct {
	line           *liner.State
	history        *HistoryManager
	running        bool
	commands       map[string]CommandHandler
	logger         *zap.Logger
	walletMgr      core.WalletManager
	accountMgr     core.AccountManager
	template       view.DisplayTemplate
	cachedPassword []byte
	passwordMgr    *security.PasswordManager
}

// CommandHandler 定义命令处理函数类型
type CommandHandler func(args []string) error

// NewREPL 创建并初始化一个新的 REPL 实例
func NewREPL(walletMgr core.WalletManager, accountMgr core.AccountManager) (*REPL, error) {
	return NewREPLWithTemplate(walletMgr, accountMgr, view.NewDefaultTemplate())
}

// NewREPLWithTemplate 使用自定义模板创建 REPL 实例
func NewREPLWithTemplate(walletMgr core.WalletManager, accountMgr core.AccountManager, template view.DisplayTemplate) (*REPL, error) {
	line := liner.NewLiner()
	line.SetCtrlCAborts(true)
	line.SetTabCompletionStyle(liner.TabCircular)

	// 简化的命令补全
	line.SetCompleter(func(line string) []string {
		return []string{
			"exit", "quit", "help", "clear", "history", "version",
			"wallet.create", "wallet.restore", "wallet.unlock", "wallet.lock", "wallet.status",
			"account.create", "account.list", "address.derive", "address.list",
		}
	})

	repl := &REPL{
		line:        line,
		history:     NewHistoryManager(),
		running:     true,
		logger:      logging.Get(),
		commands:    make(map[string]CommandHandler),
		walletMgr:   walletMgr,
		accountMgr:  accountMgr,
		template:    template,
		passwordMgr: security.GetPasswordManager(),
	}

	repl.registerCommands()
	return repl, nil
}

// registerCommands 注册所有命令
func (r *REPL) registerCommands() {
	r.commands = map[string]CommandHandler{
		// 基础命令
		"exit":    r.handleExit,
		"quit":    r.handleExit,
		"help":    r.handleHelp,
		"clear":   r.handleClear,
		"history": r.handleHistory,
		"version": r.handleVersion,

		// 钱包管理命令
		"wallet.create":  r.handleWalletCreate,
		"wallet.restore": r.handleWalletRestore,
		"wallet.unlock":  r.handleWalletUnlock,
		"wallet.lock":    r.handleWalletLock,
		"wallet.status":  r.handleWalletStatus,

		// 账户管理命令（简化参数）
		"account.create": r.handleAccountCreate,
		"account.list":   r.handleAccountList,
		"address.derive": r.handleAddressDerive,
		"address.list":   r.handleAddressList,
	}
}

// getPrompt 使用模板生成提示符
func (r *REPL) getPrompt() string {
	return r.template.Prompt(r.walletMgr.IsLocked())
}

// printWelcome 显示欢迎信息
func (r *REPL) printWelcome() {
	fmt.Println(r.template.Welcome())
}

// Run 启动 REPL 主循环（增强错误处理）
func (r *REPL) Run() {
	defer r.Close()
	r.printWelcome()

	for r.running {
		input, err := r.readInput()
		if err != nil {
			if err == liner.ErrPromptAborted {
				r.logger.Debug("Prompt aborted by user")
				fmt.Println(r.template.Goodbye())
				break
			}
			if err == io.EOF {
				r.logger.Debug("EOF received, exiting")
				fmt.Println(r.template.Goodbye())
				break
			}

			// 更详细的错误处理
			r.logger.Error("Error reading input",
				zap.Error(err),
				zap.String("error_type", fmt.Sprintf("%T", err)))

			// 检查是否是终端相关错误
			if strings.Contains(err.Error(), "invalid prompt") {
				r.logger.Warn("Invalid prompt detected, using fallback prompt")
				// 使用回退提示符重试
				if fallbackInput, fallbackErr := r.readInputWithFallback(); fallbackErr == nil {
					input = fallbackInput
				} else {
					r.logger.Error("Fallback prompt also failed, exiting", zap.Error(fallbackErr))
					break
				}
			} else {
				// 其他错误继续循环
				continue
			}
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
			fmt.Println(r.template.Error(err.Error()))
		}
	}
}

// readInputWithFallback 使用回退提示符读取输入
func (r *REPL) readInputWithFallback() (string, error) {
	// 使用简单的回退提示符
	fallbackPrompt := "slowmade> "

	r.logger.Info("Using fallback prompt", zap.String("prompt", fallbackPrompt))

	line, err := r.line.Prompt(fallbackPrompt)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

// processInput 处理用户输入
func (r *REPL) processInput(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	if handler, exists := r.commands[command]; exists {
		return handler(args)
	}

	// 移除表达式求值功能，简化 REPL
	return fmt.Errorf("unknown command: %s. Type 'help' for available commands", command)
}

// readInput 读取用户输入（支持多行）
func (r *REPL) readInput() (string, error) {
	// 获取并验证提示符
	prompt := r.getPrompt()

	// 添加调试日志（生产环境可移除）
	// r.logger.Debug("Prompt generated",
	// zap.String("prompt", fmt.Sprintf("%q", prompt)),
	// zap.Int("length", len(prompt)))

	line, err := r.line.Prompt(prompt)
	if err == liner.ErrPromptAborted || err == io.EOF {
		return "", err
	}
	if err != nil {
		r.logger.Error("Prompt failed",
			zap.Error(err),
			zap.String("prompt", fmt.Sprintf("%q", prompt)))
		return "", err
	}

	return strings.TrimSpace(line), nil
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
