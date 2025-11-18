package app

import (
	"fmt"

	"github.com/palagend/slowmade/internal/view"
)

// 钱包管理命令处理函数
func (r *REPL) handleWalletCreate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: wallet.create <password>")
	}

	password := args[0]

	// 显示创建中状态
	fmt.Println(r.template.Info("Creating new HD wallet..."))

	_, err := r.walletMgr.CreateNewWallet(password)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %v", err)
	}

	// 显示助记词（重要安全信息）
	mnemonic, err := r.walletMgr.ExportMnemonic(password)
	if err == nil && mnemonic != "" {
		fmt.Printf("\n%sMnemonic Phrase:%s\n", view.ColorYellow, view.ColorReset)
		fmt.Printf("%s%s%s\n\n", view.ColorGreen, mnemonic, view.ColorReset)
		fmt.Println(r.template.Warning("SAVE THIS MNEMONIC PHRASE IN A SECURE LOCATION!"))
		fmt.Println(r.template.Warning("It can be used to restore your wallet."))
		fmt.Println(r.template.Separator())
	}

	fmt.Println(r.template.WalletCreated("locked"))
	return nil
}

func (r *REPL) handleWalletRestore(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: wallet.restore <mnemonic> <password>")
	}

	mnemonic := args[0]
	password := args[1]

	fmt.Println(r.template.Info("Restoring wallet from mnemonic..."))

	_, err := r.walletMgr.RestoreWalletFromMnemonic(mnemonic, password)
	if err != nil {
		return fmt.Errorf("failed to restore wallet: %v", err)
	}

	fmt.Println(r.template.WalletRestored("locked"))
	return nil
}

func (r *REPL) handleWalletUnlock(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: wallet.unlock <password>")
	}

	password := args[0]

	err := r.walletMgr.UnlockWallet(password)
	if err != nil {
		return fmt.Errorf("failed to unlock wallet: %v", err)
	}

	fmt.Println(r.template.WalletUnlocked())
	return nil
}

func (r *REPL) handleWalletLock(args []string) error {
	r.walletMgr.LockWallet()
	fmt.Println(r.template.WalletLocked())
	return nil
}

func (r *REPL) handleWalletStatus(args []string) error {
	status := "locked"
	if r.walletMgr.IsUnlocked() {
		status = "unlocked"
	}
	fmt.Println(r.template.WalletStatus(status))
	return nil
}

// 简化的账户管理命令
func (r *REPL) handleAccountCreate(args []string) error {
	r.logger.Info("TODO handleAccountCreate...")
	return nil
}

func (r *REPL) handleAccountList(args []string) error {
	r.logger.Info("TODO handleAccountList...")
	return nil
}

func (r *REPL) handleAddressDerive(args []string) error {
	r.logger.Info("TODO handleAddressDerive...")
	return nil
}

func (r *REPL) handleAddressList(args []string) error {
	r.logger.Info("TODO handleAddressList...")
	return nil
}

// 基础命令处理函数
func (r *REPL) handleExit(args []string) error {
	r.running = false
	fmt.Println(r.template.Goodbye())
	return ErrExitRequested
}

func (r *REPL) handleHelp(args []string) error {
	fmt.Println(r.template.Help())
	return nil
}

func (r *REPL) handleClear(args []string) error {
	fmt.Print("\033[H\033[2J")
	return nil
}

func (r *REPL) handleHistory(args []string) error {
	history := r.history.GetLast(10) // 减少显示数量
	if len(history) == 0 {
		fmt.Println(r.template.Info("No command history"))
		return nil
	}

	fmt.Println(r.template.HistoryHeader())
	for i, cmd := range history {
		fmt.Println(r.template.HistoryItem(i, cmd))
	}
	return nil
}

func (r *REPL) handleVersion(args []string) error {
	fmt.Println(r.template.Version())
	return nil
}
