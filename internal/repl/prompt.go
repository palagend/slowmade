package repl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/c-bata/go-prompt"
	"github.com/palagend/slowmade/pkg/i18n"
	"github.com/palagend/slowmade/pkg/logging"
	"go.uber.org/zap"
)

type Prompt struct {
	session *Session
	ctx     context.Context
}

func NewPrompt(session *Session) *Prompt {
	return &Prompt{session: session}
}

func (p *Prompt) Run(ctx context.Context) error {
	p.ctx = ctx
	fmt.Println(i18n.Tr("MSG_REPL_READY"))

	completer := NewCompleter(p.session)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n" + i18n.Tr("MSG_GOODBYE"))
			return nil // 正常退出
		default:
			var input string
			if p.session.GetState() == StateWalletUnlocked {
				input = prompt.Input("wallet > ", completer.Complete,
					prompt.OptionTitle("slowmade-wallet"),
					prompt.OptionPrefixTextColor(prompt.Yellow),
				)
			} else {
				input = prompt.Input("> ", completer.Complete,
					prompt.OptionTitle("slowmade"),
					prompt.OptionPrefixTextColor(prompt.Blue),
				)
			}

			if isInputEmpty(input) {
				p.session.Exit()
				fmt.Println(i18n.Tr("MSG_GOODBYE"))
				return nil // 正常退出
			}

			input = strings.TrimSpace(input)
			if input == "" {
				continue
			}

			if err := p.executeCommand(input); err != nil {
				logging.Get().Error("Command execution failed",
					zap.String("command", input),
					zap.Error(err))
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
}

func (p *Prompt) executeCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "help", "?":
		return p.helpCommand(args)
	case "balance", "bal":
		return p.balanceCommand(args)
	case "send":
		return p.sendCommand(args)
	case "lock":
		return p.lockCommand(args)
	case "exit", "quit":
		return p.exitCommand(args)
	default:
		return fmt.Errorf(i18n.Tr("ERR_UNKNOWN_COMMAND"), command)
	}
}

func (p *Prompt) helpCommand(args []string) error {
	if len(args) == 0 {
		fmt.Println(i18n.Tr("HELP_GENERAL"))
		return nil
	}

	switch args[0] {
	case "send":
		fmt.Println(i18n.Tr("HELP_SEND"))
	case "balance":
		fmt.Println(i18n.Tr("HELP_BALANCE"))
	case "lock":
		fmt.Println(i18n.Tr("HELP_LOCK"))
	default:
		fmt.Printf(i18n.Tr("HELP_UNKNOWN"), args[0])
	}
	return nil
}

func (p *Prompt) balanceCommand(args []string) error {
	if p.session.GetState() != StateWalletUnlocked {
		return fmt.Errorf(i18n.Tr("ERR_WALLET_LOCKED"))
	}

	wallet := p.session.GetWalletManager().GetCurrentWallet()
	if wallet == nil {
		return fmt.Errorf(i18n.Tr("ERR_NO_WALLET"))
	}

	// 实际实现应从区块链节点获取余额
	balance, err := wallet.GetBalance()
	if err != nil {
		return fmt.Errorf(i18n.Tr("ERR_GET_BALANCE"), err)
	}

	fmt.Printf(i18n.Tr("MSG_BALANCE"),
		balance.Currency, balance.Amount,
		"USD", balance.FiatValue)
	return nil
}

func (p *Prompt) sendCommand(args []string) error {
	if p.session.GetState() != StateWalletUnlocked {
		return fmt.Errorf(i18n.Tr("ERR_WALLET_LOCKED"))
	}

	if len(args) < 2 {
		return fmt.Errorf(i18n.Tr("ERR_INVALID_ARGS"), "send <address> <amount>")
	}

	// 验证地址格式
	if !isValidAddress(args[0]) {
		return fmt.Errorf(i18n.Tr("ERR_INVALID_ADDRESS"))
	}

	// 确认交易
	fmt.Printf(i18n.Tr("MSG_SEND_CONFIRM"),
		args[0][:6]+"..."+args[0][len(args[0])-4:],
		args[1])

	if !p.promptConfirmation(i18n.Tr("PROMPT_CONFIRM_SEND")) {
		return fmt.Errorf(i18n.Tr("MSG_TX_CANCELLED"))
	}

	// 实际发送逻辑
	txHash, err := p.session.GetWalletManager().SendTransaction(args[0], args[1])
	if err != nil {
		return fmt.Errorf(i18n.Tr("ERR_SEND_TX"), err)
	}

	fmt.Printf(i18n.Tr("MSG_TX_SENT"), txHash)
	return nil
}

func (p *Prompt) lockCommand(args []string) error {
	p.session.Lock()
	fmt.Println(i18n.Tr("MSG_WALLET_LOCKED"))
	return nil
}

func (p *Prompt) exitCommand(args []string) error {
	select {
	case <-p.ctx.Done():
		return nil
	default:
		p.session.Exit()
		fmt.Println(i18n.Tr("MSG_GOODBYE"))
		// 返回nil表示正常退出，让上层程序结束
		return nil
	}
}

func (p *Prompt) promptConfirmation(msg string) bool {
	fmt.Printf("%s [y/N]: ", msg)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

func isValidAddress(addr string) bool {
	// 简化实现，实际应验证地址格式
	return len(addr) >= 20
}

func isInputEmpty(input string) bool {
	// 检查标准输入是否是终端
	if !isTerminal(int(os.Stdin.Fd())) {
		return false
	}

	// 如果是空字符串且标准输入已关闭（EOF）
	if input == "" {
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return true // 检测到EOF
		}
	}
	return false
}

// isTerminal 检查文件描述符是否是终端
func isTerminal(fd int) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&termios)),
		0, 0, 0,
	)
	return err == 0
}
