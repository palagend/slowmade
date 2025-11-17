// internal/app/repl.go
package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/palagend/slowmade/internal/core"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/peterh/liner"
	"go.uber.org/zap"
)

// REPL 表示一个交互式读取-求值-打印循环环境
type REPL struct {
	line      *liner.State
	history   *HistoryManager
	running   bool
	commands  map[string]CommandHandler
	logger    *zap.Logger
	walletMgr core.WalletManager // 改为接口类型而非指针
}

// CommandHandler 定义命令处理函数类型
type CommandHandler func(args []string) error

// NewREPL 创建并初始化一个新的 REPL 实例
func NewREPL(walletMgr core.WalletManager) (*REPL, error) {
	line := liner.NewLiner()
	line.SetCtrlCAborts(true)
	line.SetTabCompletionStyle(liner.TabCircular)

	// 设置增强的 Tab 补全
	line.SetCompleter(func(line string) []string {
		commands := []string{
			"exit", "quit", "help", "clear", "history", "version",
			"wallet.create", "wallet.restore", "wallet.unlock", "wallet.lock", "wallet.status",
			"account.create", "account.list", "address.derive", "address.list",
		}
		return commands
	})

	repl := &REPL{
		line:      line,
		history:   NewHistoryManager(),
		running:   true,
		logger:    logging.Get(),
		commands:  make(map[string]CommandHandler),
		walletMgr: walletMgr,
	}

	// 注册内置命令
	repl.registerCommands()

	return repl, nil
}

// registerCommands 注册所有内置命令
func (r *REPL) registerCommands() {
	r.commands = map[string]CommandHandler{
		// 基础命令
		"exit":    r.handleExit,
		"quit":    r.handleExit,
		"help":    r.handleHelp,
		"clear":   r.handleClear,
		"history": r.handleHistory,
		"version": r.handleVersion,

		// 钱包管理命令（已根据接口调整）
		"wallet.create":  r.handleWalletCreate,
		"wallet.restore": r.handleWalletRestore,
		"wallet.unlock":  r.handleWalletUnlock,
		"wallet.lock":    r.handleWalletLock,
		"wallet.status":  r.handleWalletStatus,

		// 账户管理命令
		"account.create": r.handleAccountCreate,
		"account.list":   r.handleAccountList,
		"address.derive": r.handleAddressDerive,
		"address.list":   r.handleAddressList,
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

// getPrompt 生成带钱包状态的提示符
func (r *REPL) getPrompt() string {
	if r.walletMgr.IsUnlocked() {
		return "slowmade(unlocked)> "
	}
	return "slowmade(locked)> "
}

// 钱包管理命令处理函数（根据接口更新）
func (r *REPL) handleWalletCreate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: wallet.create <password>")
	}

	password := args[0]

	_, err := r.walletMgr.CreateNewWallet(password)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %v", err)
	}

	fmt.Printf("✅ Wallet created successfully!\n")
	fmt.Printf("   Status: %s\n", "Locked") // 新创建的钱包默认锁定
	return nil
}

func (r *REPL) handleWalletRestore(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: wallet.restore <mnemonic> <password>")
	}

	mnemonic := args[0]
	password := args[1]

	_, err := r.walletMgr.RestoreWalletFromMnemonic(mnemonic, password)
	if err != nil {
		return fmt.Errorf("failed to restore wallet: %v", err)
	}

	fmt.Printf("✅ Wallet restored successfully!\n")
	fmt.Printf("   Status: %s\n", "Locked")

	return nil
}

func (r *REPL) handleWalletUnlock(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: wallet.unlock <encryptedMnemonic> <password>")
	}

	// 注意：这里需要根据实际实现获取加密的助记词
	encryptedMnemonic := []byte(args[0])
	password := args[1]

	err := r.walletMgr.UnlockWallet(encryptedMnemonic, password)
	if err != nil {
		return fmt.Errorf("failed to unlock wallet: %v", err)
	}

	fmt.Printf("✅ Wallet unlocked successfully!\n")
	return nil
}

func (r *REPL) handleWalletLock(args []string) error {
	r.walletMgr.LockWallet()
	fmt.Printf("✅ Wallet locked successfully!\n")
	return nil
}

func (r *REPL) handleWalletStatus(args []string) error {
	status := "locked"
	if r.walletMgr.IsUnlocked() {
		status = "unlocked"
	}

	fmt.Printf("Wallet Status: %s\n", status)
	return nil
}

// 账户管理命令处理函数
func (r *REPL) handleAccountCreate(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: account.create <coinType> <accountIndex> <password> <accountName>")
	}

	// 这里需要将参数转换为适当类型
	// coinTypeStr := args[0]
	// accountIndexStr := args[1]
	// password := args[2]
	// accountName := args[3]

	// 实际实现需要类型转换和错误处理
	fmt.Printf("Account creation command received\n")
	return nil
}

func (r *REPL) handleAccountList(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: account.list <coinType>")
	}

	// coinTypeStr := args[0]
	fmt.Printf("Account list command received\n")
	return nil
}

func (r *REPL) handleAddressDerive(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: address.derive <accountID> <changeType> <addressIndex> <password>")
	}

	// accountID := args[0]
	// changeTypeStr := args[1]
	// addressIndexStr := args[2]
	// password := args[3]

	fmt.Printf("Address derivation command received\n")
	return nil
}

func (r *REPL) handleAddressList(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: address.list <accountID>")
	}

	// accountID := args[0]
	fmt.Printf("Address list command received\n")
	return nil
}

// 更新help命令以反映新的命令结构
func (r *REPL) handleHelp(args []string) error {
	helpText := `
Available commands:

Basic Commands:
  exit, quit    - Exit the REPL
  help          - Show this help message
  clear         - Clear the screen
  history       - Show command history
  version       - Show version information

Wallet Management (BIP44 Compatible):
  wallet.create <password>                    - Create a new HD wallet
  wallet.restore <mnemonic> <password>        - Restore wallet from mnemonic
  wallet.unlock <encryptedMnemonic> <password> - Unlock wallet with password
  wallet.lock                                 - Lock wallet (clear sensitive data)
  wallet.status                               - Check wallet lock status

Account Management:
  account.create <coinType> <accountIndex> <password> <name> - Create new account
  account.list <coinType>                   - List accounts for coin type
  address.derive <accountID> <change> <index> <password> - Derive new address
  address.list <accountID>                  - List all addresses for account

Expressions:
  Enter any expression to evaluate it
  Use \ at the end of a line for multi-line input

Shortcuts:
  Ctrl+D        - Exit the REPL
  Ctrl+C        - Abort current input
  Up/Down arrows - Navigate command history
  Tab           - Auto-completion

Examples:
  wallet.create mySecurePassword123
  wallet.restore "word1 word2 ... word12" myPassword123
  wallet.unlock <encryptedData> myPassword123
  account.create 60 0 myPassword123 "My ETH Account"
`
	fmt.Println(helpText)
	return nil
}

// 原有的其他方法保持不变（handleExit, handleClear, handleHistory, handleVersion等）
func (r *REPL) handleExit(args []string) error {
	r.running = false
	fmt.Println("Goodbye!")
	return ErrExitRequested
}

func (r *REPL) handleClear(args []string) error {
	fmt.Print("\033[H\033[2J")
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
	fmt.Println("Slowmade REPL v1.1.0 with BIP44 Wallet Management")
	return nil
}

func (r *REPL) printWelcome() {
	welcome := `
Welcome to Slowmade REPL with BIP44 Wallet Management!
Type 'help' for available commands, 'exit' to quit.

Features:
• BIP44-compliant HD wallet creation and management
• Multi-currency account support (BTC, ETH, SOL, BNB, SUI)
• Secure address derivation
• Wallet locking/unlocking with password protection
• Mnemonic-based wallet restoration
`
	fmt.Println(welcome)
}

// processInput, readInput, evaluateExpression 等方法保持不变
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

	return r.handleExpression(input)
}

func (r *REPL) handleExpression(expr string) error {
	result, err := r.evaluateExpression(expr)
	if err != nil {
		return err
	}
	if result != "" {
		fmt.Println(result)
	}
	return nil
}

func (r *REPL) evaluateExpression(expr string) (string, error) {
	return fmt.Sprintf("Expression: %s (evaluation would go here)", expr), nil
}

func (r *REPL) readInput() (string, error) {
	prompt := r.getPrompt()
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

func (r *REPL) Close() {
	if r.line != nil {
		if err := r.history.Save(); err != nil {
			r.logger.Warn("Failed to save history on close", zap.Error(err))
		}
		r.line.Close()
	}
}

// 辅助函数
func askForConfirmation(prompt string) bool {
	fmt.Printf("%s (y/N): ", prompt)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}
