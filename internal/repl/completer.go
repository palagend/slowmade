package repl

import (
	"strings"

	"github.com/c-bata/go-prompt"
)

type Completer struct {
	session *Session
}

func NewCompleter(session *Session) *Completer {
	return &Completer{session: session}
}

func (c *Completer) Complete(d prompt.Document) []prompt.Suggest {
	textBeforeCursor := d.TextBeforeCursor()
	words := strings.Fields(textBeforeCursor)

	// 空输入时显示所有命令
	if len(words) == 0 {
		return c.getAllCommands()
	}

	// 第一个词是命令，显示子命令或参数建议
	if len(words) == 1 {
		return c.getCommandSuggestions(words[0])
	}

	// 根据当前命令提供上下文相关建议
	switch words[0] {
	case "send":
		return c.getSendSuggestions(words, d)
	case "help":
		return c.getHelpSuggestions(words)
	}

	return []prompt.Suggest{}
}

// getAllCommands 返回所有可用命令
func (c *Completer) getAllCommands() []prompt.Suggest {
	baseCommands := []prompt.Suggest{
		{Text: "help", Description: "Show help information"},
		{Text: "balance", Description: "Show current balance"},
		{Text: "send", Description: "Send transaction"},
		{Text: "lock", Description: "Lock the wallet"},
		{Text: "exit", Description: "Exit the application"},
	}

	// 如果钱包已解锁，添加更多命令
	if c.session.GetState() == StateWalletUnlocked {
		unlockedCommands := []prompt.Suggest{
			{Text: "address", Description: "Show current wallet address"},
			{Text: "new", Description: "Create new address in HD wallet"},
		}
		baseCommands = append(baseCommands, unlockedCommands...)
	}

	return baseCommands
}

// getCommandSuggestions 根据输入前缀过滤命令
func (c *Completer) getCommandSuggestions(prefix string) []prompt.Suggest {
	allCommands := c.getAllCommands()
	suggestions := make([]prompt.Suggest, 0)

	for _, cmd := range allCommands {
		if strings.HasPrefix(cmd.Text, prefix) {
			suggestions = append(suggestions, cmd)
		}
	}

	return suggestions
}

// getSendSuggestions 为send命令提供补全建议
func (c *Completer) getSendSuggestions(words []string, d prompt.Document) []prompt.Suggest {
	// 如果正在输入地址，提供地址补全
	if len(words) == 2 && !strings.HasSuffix(d.TextBeforeCursor(), " ") {
		return c.getAddressSuggestions(words[1])
	}

	// 如果正在输入金额，提供单位建议
	if len(words) == 3 && !strings.HasSuffix(d.TextBeforeCursor(), " ") {
		return []prompt.Suggest{
			{Text: "1", Description: "Amount to send"},
			{Text: "0.1", Description: "Amount to send"},
			{Text: "0.01", Description: "Amount to send"},
		}
	}

	// 提供gas price选项
	if len(words) >= 3 && strings.HasPrefix(words[len(words)-1], "--gas-price") {
		return []prompt.Suggest{
			{Text: "--gas-price", Description: "Set custom gas price"},
			{Text: "--gas-limit", Description: "Set custom gas limit"},
		}
	}

	return []prompt.Suggest{}
}

// getHelpSuggestions 为help命令提供补全建议
func (c *Completer) getHelpSuggestions(words []string) []prompt.Suggest {
	if len(words) == 1 {
		return []prompt.Suggest{
			{Text: "help", Description: "Show general help"},
			{Text: "help send", Description: "Show send command help"},
			{Text: "help balance", Description: "Show balance command help"},
			{Text: "help lock", Description: "Show lock command help"},
		}
	}

	return []prompt.Suggest{}
}

// getAddressSuggestions 提供地址补全建议
func (c *Completer) getAddressSuggestions(prefix string) []prompt.Suggest {
	// 实际实现应该从钱包管理器获取历史地址
	// 这里简化处理，返回空数组
	return []prompt.Suggest{}
}
