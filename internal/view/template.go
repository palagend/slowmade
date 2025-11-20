package view

import (
	"fmt"

	"github.com/palagend/slowmade/internal/core"
	"github.com/spf13/viper"
)

// DisplayTemplate 定义显示模板接口
type DisplayTemplate interface {
	// 欢迎信息
	Welcome() string

	// 提示符
	Prompt(isLocked bool) string

	// 命令输出模板
	WalletCreated(status string) string
	AccountList(accounts []*core.CoinAccount) string
	WalletRestored(status string) string
	WalletUnlocked() string
	WalletLocked() string
	WalletStatus(status string) string

	// 帮助信息
	Help() string

	// 通用消息
	Goodbye() string
	Error(message string) string
	Info(message string) string
	Success(message string) string
	Warning(message string) string

	// 历史记录
	HistoryHeader() string
	HistoryItem(index int, command string) string

	// 版本信息
	Version() string

	// 分隔线
	Separator() string
}

// DefaultTemplate 默认显示模板（使用 ASCII 和颜色）
type DefaultTemplate struct{}

func (t *DefaultTemplate) Welcome() string {
	return fmt.Sprintf(`%s%s
╔════════════════════════════════════════════════════════════════╗
║                        SLOWMADE WALLET REPL                    ║
║                    BIP44 HD Wallet Management                  ║
╚════════════════════════════════════════════════════════════════╝%s

%s[FEATURES]%s
  • HD Wallet Creation & Restoration
  • Multi-Currency Support (BTC, ETH, SOL, BNB, SUI)
  • Secure Hierarchical Deterministic Key Derivation
  • Encrypted Wallet Storage
  • Mnemonic Phrase Backup & Recovery

Type '%shelp%s' for available commands, '%sexit%s' to quit.
`,
		StyleBold, ColorCyan, ColorReset,
		ColorGreen, ColorReset,
		ColorYellow, ColorReset, ColorYellow, ColorReset)
}

func (t *DefaultTemplate) Prompt(isLocked bool) string {
	status := "LOCKED"
	if !isLocked {
		status = "unlocked"
	}
	return fmt.Sprintf("[%s(%s)] > ", viper.GetString("storage.base_dir"), status)
}

func (t *DefaultTemplate) WalletCreated(status string) string {
	statusColor := ColorRed
	if status == "unlocked" {
		statusColor = ColorGreen
	}

	return fmt.Sprintf(`%s
╔════════════════════════════════════════════════════════════════╗
║                          WALLET CREATED                        ║
╚════════════════════════════════════════════════════════════════╝%s
%s Wallet created successfully!
   Status: %s%s%s

%s IMPORTANT:
  • Save your mnemonic phrase in a secure location
  • Never share your private keys or mnemonic phrase
  • Backup your wallet regularly
`,
		ColorGreen, ColorReset,
		SuccessIcon(),
		statusColor, status, ColorReset,
		WarningIcon())
}

func (t *DefaultTemplate) WalletRestored(status string) string {
	statusColor := ColorRed
	if status == "unlocked" {
		statusColor = ColorGreen
	}

	return fmt.Sprintf(`%s
╔════════════════════════════════════════════════════════════════╗
║                          WALLET RESTORED                       ║
╚════════════════════════════════════════════════════════════════╝%s
%s Wallet restored from mnemonic successfully!
   Status: %s%s%s
`,
		ColorBlue, ColorReset,
		SuccessIcon(),
		statusColor, status, ColorReset)
}

func (t *DefaultTemplate) WalletUnlocked() string {
	return fmt.Sprintf(`%s
╔════════════════════════════════════════════════════════════════╗
║                          WALLET UNLOCKED                       ║
╚════════════════════════════════════════════════════════════════╝%s
%s Wallet unlocked successfully!
   You can now perform account operations
`,
		ColorGreen, ColorReset,
		SuccessIcon())
}

func (t *DefaultTemplate) WalletLocked() string {
	return fmt.Sprintf(`%s
╔════════════════════════════════════════════════════════════════╗
║                          WALLET LOCKED                         ║
╚════════════════════════════════════════════════════════════════╝%s
%s Wallet locked successfully!
   All sensitive data has been cleared from memory
`,
		ColorYellow, ColorReset,
		SuccessIcon())
}

func (t *DefaultTemplate) WalletStatus(status string) string {
	statusColor := ColorRed
	statusIcon := "[Lo]"
	if status == "unlocked" {
		statusColor = ColorGreen
		statusIcon = "[Lo]"
	}

	return fmt.Sprintf(`%sWallet Status:%s %s%s%s %s`,
		ColorCyan, ColorReset,
		statusColor, status, ColorReset, statusIcon)
}

func (t *DefaultTemplate) Help() string {
	return fmt.Sprintf(`%s
╔════════════════════════════════════════════════════════════════╗
║                         AVAILABLE COMMANDS                     ║
╚════════════════════════════════════════════════════════════════╝%s

%s[WALLET MANAGEMENT]%s
  %swallet.create [password]%s        - Create a new HD wallet
  %swallet.restore <mnemonic> <password>%s - Restore wallet from mnemonic
  %swallet.unlock <password>%s        - Unlock wallet with password
  %swallet.lock%s                   - Lock wallet (clear sensitive data)
  %swallet.status%s                 - Check wallet lock status

%s[ACCOUNT MANAGEMENT]%s
  %saccount.create <derivationPath>%s - Create new account
  %saccount.list <CoinSymbol>%s                 - List all accounts for <CoinSymbol>
  %saddress.derive <accountID> <password>%s - Derive new address
  %saddress.list <accountID>%s        - List addresses for account

%s[BASIC COMMANDS]%s
  %sexit, quit%s    - Exit the REPL
  %shelp%s        - Show this help message
  %sclear%s       - Clear the screen
  %shistory%s     - Show command history
  %sversion%s     - Show version information

%s[EXAMPLES]%s
  wallet.create mySecurePassword123
  wallet.restore "word1 word2 ... word12" myPassword123
  wallet.unlock myPassword123
  account.create  m/44'/0'/0'/0/0

%s[SHORTCUTS]%s
  Ctrl+D, Ctrl+C  - Exit immediately
  Tab            - Auto-completion
`,
		ColorCyan, ColorReset,

		ColorYellow, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,

		ColorYellow, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,

		ColorYellow, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,

		ColorYellow, ColorReset,

		ColorYellow, ColorReset)
}

func (t *DefaultTemplate) Goodbye() string {
	return fmt.Sprintf(`%s
╔════════════════════════════════════════════════════════════════╗
║                          GOODBYE!                              ║
║                 Thank you for using Slowmade                   ║
╚════════════════════════════════════════════════════════════════╝%s`,
		ColorBlue, ColorReset)
}

func (t *DefaultTemplate) Error(message string) string {
	return fmt.Sprintf("%s[ERROR]%s %s", ColorRed, ColorReset, message)
}

func (t *DefaultTemplate) Info(message string) string {
	return fmt.Sprintf("%s[INFO]%s %s", ColorBlue, ColorReset, message)
}

func (t *DefaultTemplate) Success(message string) string {
	return fmt.Sprintf("%s[SUCCESS]%s %s", ColorGreen, ColorReset, message)
}

func (t *DefaultTemplate) Warning(message string) string {
	return fmt.Sprintf("%s[WARNING]%s %s", ColorYellow, ColorReset, message)
}

func (t *DefaultTemplate) HistoryHeader() string {
	return fmt.Sprintf("%sCommand History:%s", ColorCyan, ColorReset)
}

func (t *DefaultTemplate) HistoryItem(index int, command string) string {
	return fmt.Sprintf("  %s%2d.%s %s", ColorGray, index+1, ColorReset, command)
}

func (t *DefaultTemplate) Version() string {
	return fmt.Sprintf("%sSlowmade REPL v1.1.0 - BIP44 HD Wallet Management%s", ColorCyan, ColorReset)
}

func (t *DefaultTemplate) Separator() string {
	return fmt.Sprintf("%s%s%s", ColorGray, "────────────────────────────────────────────────────────────────────────", ColorReset)
}

func (t *DefaultTemplate) AccountList(accounts []*core.CoinAccount) string {
	if len(accounts) == 0 {
		return fmt.Sprintf(`%s
╔════════════════════════════════════════════════════════════════╗
║                          ACCOUNT LIST                          ║
╚════════════════════════════════════════════════════════════════╝%s

%s No accounts found
`,
			ColorYellow, ColorReset,
			InfoIcon())
	}

	// 构建账户列表表格
	accountsTable := ""
	for i, account := range accounts {
		// 交替行颜色
		rowColor := ColorReset
		if i%2 == 0 {
			rowColor = ColorCyan
		}

		accountsTable += fmt.Sprintf(`%s
   Account #%d:
     ID:       %s
     Coin:     %s%s%s
     Path:     %s
     Key:      %s...%s
`,
			rowColor,
			i+1,
			account.ID,
			ColorYellow, account.CoinSymbol, rowColor,
			account.DerivationPath,
			ColorGray, account.EncryptedAccountPrivateKey[:8]+"..."+account.EncryptedAccountPrivateKey[len(account.EncryptedAccountPrivateKey)-8:])
	}

	return fmt.Sprintf(`%s
╔════════════════════════════════════════════════════════════════╗
║                          ACCOUNT LIST                          ║
╚════════════════════════════════════════════════════════════════╝%s

%s Found %s%d%s accounts:

%s%s
%s Account Details:
%s
%s Security Notes:
  • Each account has a unique derivation path
  • Private keys are encrypted and secure
  • Use account ID for transaction operations
`,
		ColorGreen, ColorReset,
		SuccessIcon(),
		ColorCyan, len(accounts), ColorReset,
		accountsTable,
		ColorReset,
		InfoIcon(),
		ColorReset,
		WarningIcon())
}
