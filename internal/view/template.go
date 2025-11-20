package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/palagend/slowmade/internal/core"
	"github.com/spf13/viper"
)

// DisplayTemplate 定义显示模板接口
type DisplayTemplate interface {
	Welcome() string
	Prompt(isLocked bool) string
	WalletCreated(status string) string
	AccountList(accounts []*core.CoinAccount) string
	WalletRestored(status string) string
	WalletUnlocked() string
	WalletLocked() string
	WalletStatus(status string) string
	Help() string
	Goodbye() string
	Error(message string) string
	Info(message string) string
	Success(message string) string
	Warning(message string) string
	HistoryHeader() string
	HistoryItem(index int, command string) string
	Version() string
	Separator() string
}

// DefaultTemplate 使用 lipgloss 的现代化模板
type DefaultTemplate struct {
	styles *Styles
}

// Styles 集中管理所有样式
type Styles struct {
	Title     lipgloss.Style
	Header    lipgloss.Style
	Success   lipgloss.Style
	Error     lipgloss.Style
	Warning   lipgloss.Style
	Info      lipgloss.Style
	Highlight lipgloss.Style
	Muted     lipgloss.Style
	Accent    lipgloss.Style
	Border    lipgloss.Style
}

// ASCII 图标定义
const (
	IconSuccess  = "[+]"
	IconError    = "[X]"
	IconWarning  = "[*]"
	IconInfo     = "[i]"
	IconLock     = "[L]"
	IconOpen     = "[O]"
	IconUnlock   = "[U]"
	IconWallet   = "[W]"
	IconAccount  = "[A]"
	IconArrow    = "=>"
	IconDot      = "•"
	IconStar     = "★"
	IconCircle   = "○"
	IconSquare   = "■"
	IconTriangle = "▶"
)

// NewDefaultTemplate 创建新的模板实例
func NewDefaultTemplate() *DefaultTemplate {
	return &DefaultTemplate{
		styles: createStyles(),
	}
}

// createStyles 创建统一的样式定义
func createStyles() *Styles {
	return &Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Align(lipgloss.Center).
			Padding(0, 1),

		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginTop(1).
			MarginBottom(1),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Italic(true),

		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("201")).
			Bold(true),

		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Faint(true),

		Accent: lipgloss.NewStyle().
			Foreground(lipgloss.Color("93")),

		Border: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")),
	}
}

// 辅助函数
func (t *DefaultTemplate) banner(title string) string {
	width := 70
	border := strings.Repeat("=", width-2)

	titleLine := lipgloss.PlaceHorizontal(
		width-4,
		lipgloss.Center,
		title,
		lipgloss.WithWhitespaceChars(" "),
	)

	return fmt.Sprintf("+%s+\n| %s |\n+%s+",
		border, titleLine, border)
}

func (t *DefaultTemplate) section(title string, content string) string {
	return fmt.Sprintf("%s\n\n%s",
		t.styles.Header.Render(title),
		content)
}

func (t *DefaultTemplate) statusIcon(status string) string {
	if status == "unlocked" {
		return IconUnlock
	}
	return IconLock
}

func (t *DefaultTemplate) statusStyle(status string) lipgloss.Style {
	if status == "unlocked" {
		return t.styles.Success
	}
	return t.styles.Error
}

// 实现接口方法
func (t *DefaultTemplate) Welcome() string {
	features := []string{
		"HD Wallet Creation & Restoration",
		"Multi-Currency Support (BTC, ETH, SOL, BNB, SUI)",
		"Secure Hierarchical Deterministic Key Derivation",
		"Encrypted Wallet Storage",
		"Mnemonic Phrase Backup & Recovery",
	}

	featureItems := ""
	for _, feature := range features {
		featureItems += fmt.Sprintf("  %s %s\n", IconDot, feature)
	}

	return fmt.Sprintf(`%s

%s

Type '%s' for available commands, '%s' to quit.`,
		t.banner("SLOWMADE WALLET REPL"),
		t.section("FEATURES", featureItems),
		t.styles.Highlight.Render("help"),
		t.styles.Highlight.Render("exit"),
	)
}

func (t *DefaultTemplate) Prompt(isLocked bool) string {
	statusIcon := IconLock
	if !isLocked {
		statusIcon = IconOpen
	}
	return fmt.Sprintf("%s(%s) > ", statusIcon, viper.GetString("storage.base_dir"))

}

func (t *DefaultTemplate) WalletCreated(status string) string {
	content := fmt.Sprintf(`%s Wallet created successfully!
   Status: %s %s

%s IMPORTANT:
  %s Save your mnemonic phrase in a secure location
  %s Never share your private keys or mnemonic phrase  
  %s Backup your wallet regularly`,
		IconSuccess,
		t.statusStyle(status).Render(status),
		t.statusIcon(status),
		IconWarning,
		IconDot,
		IconDot,
		IconDot,
	)

	return fmt.Sprintf("%s\n\n%s", t.banner("WALLET CREATED"), content)
}

func (t *DefaultTemplate) AccountList(accounts []*core.CoinAccount) string {
	if len(accounts) == 0 {
		return fmt.Sprintf("%s\n\n%s No accounts found",
			t.banner("ACCOUNT LIST"),
			IconInfo)
	}

	// 使用简洁的列表显示
	var accountList strings.Builder
	accountList.WriteString(fmt.Sprintf("%s Found %s accounts:\n\n",
		IconSuccess,
		t.styles.Highlight.Render(fmt.Sprintf("%d", len(accounts)))))

	for i, account := range accounts {
		keyPreview := "[ENCRYPTED]"
		if len(account.EncryptedAccountPrivateKey) > 16 {
			keyPreview = account.EncryptedAccountPrivateKey[:8] + "..." +
				account.EncryptedAccountPrivateKey[len(account.EncryptedAccountPrivateKey)-8:]
		}

		accountList.WriteString(fmt.Sprintf(`%s Account #%d
  %s ID:       %s
  %s Coin:     %s
  %s Path:     %s
  %s Key:      %s
`,
			IconSquare, i+1,
			IconArrow, account.ID,
			IconArrow, t.styles.Highlight.Render(account.CoinSymbol),
			IconArrow, account.DerivationPath,
			IconArrow, t.styles.Muted.Render(keyPreview),
		))
	}

	return fmt.Sprintf("%s\n\n%s\n\n%s Each account has a unique derivation path",
		t.banner("ACCOUNT LIST"),
		accountList.String(),
		IconInfo,
	)
}

func (t *DefaultTemplate) WalletRestored(status string) string {
	return fmt.Sprintf(`%s

%s Wallet restored from mnemonic successfully!
   Status: %s %s`,
		t.banner("WALLET RESTORED"),
		IconSuccess,
		t.statusStyle(status).Render(status),
		t.statusIcon(status),
	)
}

func (t *DefaultTemplate) WalletUnlocked() string {
	return fmt.Sprintf(`%s

%s Wallet unlocked successfully!
   %s You can now perform account operations`,
		t.banner("WALLET UNLOCKED"),
		IconSuccess,
		IconArrow,
	)
}

func (t *DefaultTemplate) WalletLocked() string {
	return fmt.Sprintf(`%s

%s Wallet locked successfully!
   %s All sensitive data has been cleared from memory`,
		t.banner("WALLET LOCKED"),
		IconSuccess,
		IconArrow,
	)
}

func (t *DefaultTemplate) WalletStatus(status string) string {
	return fmt.Sprintf("Wallet Status: %s %s",
		t.statusStyle(status).Render(status),
		t.statusIcon(status))
}

func (t *DefaultTemplate) Help() string {
	commands := map[string][]string{
		"WALLET MANAGEMENT": {
			"wallet.create [password]        " + IconArrow + " Create a new HD wallet",
			"wallet.restore <mnemonic> <password> " + IconArrow + " Restore wallet from mnemonic",
			"wallet.unlock <password>        " + IconArrow + " Unlock wallet with password",
			"wallet.lock                   " + IconArrow + " Lock wallet",
			"wallet.status                 " + IconArrow + " Check wallet status",
		},
		"ACCOUNT MANAGEMENT": {
			"account.create <derivationPath> " + IconArrow + " Create new account",
			"account.list <CoinSymbol>       " + IconArrow + " List accounts",
			"address.derive <accountID> <password> " + IconArrow + " Derive new address",
			"address.list <accountID>        " + IconArrow + " List addresses",
		},
		"BASIC COMMANDS": {
			"exit, quit    " + IconArrow + " Exit the REPL",
			"help        " + IconArrow + " Show help",
			"clear       " + IconArrow + " Clear screen",
			"history     " + IconArrow + " Show history",
			"version     " + IconArrow + " Show version",
		},
	}

	var helpText strings.Builder
	helpText.WriteString(t.banner("AVAILABLE COMMANDS") + "\n\n")

	for category, cmds := range commands {
		helpText.WriteString(t.styles.Header.Render(category) + "\n")
		for _, cmd := range cmds {
			// 分割命令和描述
			parts := strings.SplitN(cmd, IconArrow, 2)
			if len(parts) == 2 {
				helpText.WriteString(fmt.Sprintf("  %s %s %s\n",
					t.styles.Highlight.Render(strings.TrimSpace(parts[0])),
					IconArrow,
					strings.TrimSpace(parts[1]),
				))
			} else {
				helpText.WriteString("  " + cmd + "\n")
			}
		}
		helpText.WriteString("\n")
	}

	// 添加快捷键说明
	helpText.WriteString(t.styles.Header.Render("SHORTCUTS") + "\n")
	helpText.WriteString(fmt.Sprintf("  Ctrl+D, Ctrl+C  %s Exit immediately\n", IconArrow))
	helpText.WriteString(fmt.Sprintf("  Tab            %s Auto-completion\n", IconArrow))

	return helpText.String()
}

// 简化通用消息方法
func (t *DefaultTemplate) Error(message string) string {
	return fmt.Sprintf("%s %s", IconError, t.styles.Error.Render(message))
}

func (t *DefaultTemplate) Info(message string) string {
	return fmt.Sprintf("%s %s", IconInfo, t.styles.Info.Render(message))
}

func (t *DefaultTemplate) Success(message string) string {
	return fmt.Sprintf("%s %s", IconSuccess, t.styles.Success.Render(message))
}

func (t *DefaultTemplate) Warning(message string) string {
	return fmt.Sprintf("%s %s", IconWarning, t.styles.Warning.Render(message))
}

func (t *DefaultTemplate) Goodbye() string {
	return t.banner("GOODBYE! Thank you for using Slowmade")
}

func (t *DefaultTemplate) HistoryHeader() string {
	return t.styles.Header.Render("Command History:")
}

func (t *DefaultTemplate) HistoryItem(index int, command string) string {
	return fmt.Sprintf("  %s%2d.%s %s",
		t.styles.Muted.Render(">"),
		index+1,
		lipgloss.NewStyle(),
		command)
}

func (t *DefaultTemplate) Version() string {
	return t.styles.Info.Render("Slowmade REPL v1.1.0 - BIP44 HD Wallet Management")
}

func (t *DefaultTemplate) Separator() string {
	return t.styles.Border.Render(strings.Repeat("-", 60))
}
