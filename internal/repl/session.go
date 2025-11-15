package repl

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/palagend/slowmade/internal/core"
	"github.com/palagend/slowmade/pkg/i18n"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/spf13/viper"
)

type SessionState int

const (
	StateInitial SessionState = iota
	StateWalletUnlocked
	StateWalletLocked
)

type Session struct {
	walletManager *core.WalletManager
	state         SessionState
	language      string
	history       []string
	mu            sync.RWMutex
	cancelFunc    context.CancelFunc
}

func NewSession(wm *core.WalletManager) *Session {
	return &Session{
		walletManager: wm,
		state:         StateInitial,
		language:      viper.GetString("ui.lang"),
		history:       make([]string, 0),
	}
}

func (s *Session) Start(ctx context.Context) error {
	// 设置取消函数
	ctx, s.cancelFunc = context.WithCancel(ctx)

	// 监听上下文取消
	go func() {
		<-ctx.Done()
		logging.Get().Info("Context cancelled, exiting session")
		s.Exit()
	}()

	// 设置语言
	i18n.SetLanguage(s.language)

	// 检查现有钱包
	if s.hasExistingWallets() {
		return s.handleExistingWallet(ctx)
	}
	return s.handleNewWallet(ctx)
}

func (s *Session) hasExistingWallets() bool {
	// 实际实现应检查keystore目录
	// 这里简化返回false
	return false
}

func (s *Session) handleExistingWallet(ctx context.Context) error {
	fmt.Println(i18n.Tr("MSG_WELCOME_BACK"))
	// 简化实现：直接进入新钱包创建
	return s.handleNewWallet(ctx)
}

func (s *Session) handleNewWallet(ctx context.Context) error {
	fmt.Println(i18n.Tr("MSG_WELCOME_NEW"))

	password, err := s.promptPassword()
	if err != nil {
		return fmt.Errorf("password prompt failed: %w", err)
	}

	mnemonic, err := s.walletManager.CreateWallet(password, "")
	if err != nil {
		return fmt.Errorf(i18n.Tr("ERR_CREATE_WALLET_FAILED"), err)
	}

	// 显示助记词并要求备份
	fmt.Printf("\n%s\n\n", i18n.Tr("MSG_BACKUP_MNEMONIC"))
	fmt.Printf("Mnemonic: %s\n\n", mnemonic)
	fmt.Println(i18n.Tr("MSG_MNEMONIC_WARNING"))

	if !s.promptConfirmation(i18n.Tr("PROMPT_CONFIRM_BACKUP")) {
		fmt.Println(i18n.Tr("MSG_BACKUP_REQUIRED"))
		return fmt.Errorf("backup confirmation required")
	}

	s.state = StateWalletUnlocked
	fmt.Println(i18n.Tr("MSG_WALLET_READY"))

	return s.startREPL(ctx)
}

func (s *Session) promptPassword() (string, error) {
	// 实际实现应使用term.ReadPassword等安全输入方式
	fmt.Print(i18n.Tr("PROMPT_ENTER_PASSWORD") + ": ")
	return "testpassword123", nil // 简化实现
}

func (s *Session) promptConfirmation(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

func (s *Session) startREPL(ctx context.Context) error {
	prompt := NewPrompt(s)

	// 启动REPL，如果正常退出则返回nil
	err := prompt.Run(ctx)
	if err != nil {
		return err
	}

	// REPL正常退出，返回nil让上层知道可以结束程序
	return nil
}

func (s *Session) Lock() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == StateWalletUnlocked {
		s.walletManager.LockWallet()
		s.state = StateWalletLocked
		logging.Get().Info("Wallet locked")
	}
}

func (s *Session) Exit() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == StateWalletUnlocked {
		s.walletManager.LockWallet()
		s.state = StateWalletLocked
		logging.Get().Info("Wallet locked")
	}

	logging.Get().Info("Session terminated")
}

func (s *Session) GetState() SessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *Session) GetWalletManager() *core.WalletManager {
	return s.walletManager
}

func Cleanup() {
	logging.Get().Info("Performing global cleanup")
	// 执行全局资源清理
}
