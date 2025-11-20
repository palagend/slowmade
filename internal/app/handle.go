package app

import (
	"fmt"
	"syscall"

	"github.com/palagend/slowmade/internal/core"
	"github.com/palagend/slowmade/internal/view"
	"github.com/palagend/slowmade/pkg/coin"
	"github.com/palagend/slowmade/pkg/logging"
	"golang.org/x/term"
)

// 钱包管理命令处理函数
func (r *REPL) handleWalletCreate(args []string) error {
	var password string
	if len(args) > 1 {
		return fmt.Errorf("usage: wallet.create [password]")
	}
	// 如果没有提供密码参数，提示用户输入
	if len(args) < 1 {
		fmt.Print("Enter password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %v", err)
		}
		password = string(bytePassword)
		fmt.Println() // 换行，因为ReadPassword不会自动换行
	} else {
		// 保持向后兼容，支持命令行参数方式（但不推荐）
		password = args[0]
		fmt.Println("Warning: Using password from command line arguments is not secure")
	}

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
	var password string
	var err error

	// 如果已经解锁，提示用户
	if !r.walletMgr.IsLocked() {
		fmt.Println("Wallet is already unlocked")
		return nil
	}

	// 如果没有提供密码参数，提示用户输入
	if len(args) < 1 {
		fmt.Print("Enter password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %v", err)
		}
		password = string(bytePassword)
		fmt.Println() // 换行，因为ReadPassword不会自动换行
	} else {
		// 保持向后兼容，支持命令行参数方式（但不推荐）
		password = args[0]
		fmt.Println("Warning: Using password from command line arguments is not secure")
	}

	err = r.walletMgr.UnlockWallet(password)
	if err != nil {
		return fmt.Errorf("failed to unlock wallet: %v", err)
	}
	r.passwordMgr.SetPassword(password)
	fmt.Println(r.template.WalletUnlocked())
	return nil
}

func (r *REPL) handleWalletLock(args []string) error {
	// 锁定钱包
	r.walletMgr.LockWallet()
	r.passwordMgr.Clear()
	fmt.Println(r.template.WalletLocked())
	return nil
}

func (r *REPL) handleWalletStatus(args []string) error {
	status := "locked"
	if !r.walletMgr.IsLocked() {
		status = "unlocked"
	}
	fmt.Println(r.template.WalletStatus(status))
	return nil
}

// 简化的账户管理命令
func (r *REPL) handleAccountCreate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("用法: account create  <派生路径>")
	}

	derivationPath, err := core.ParseDerivationPath(args[0])
	if err != nil {
		return err
	}

	// 创建新账户
	account, err := r.accountMgr.CreateNewAccount(derivationPath)
	if err != nil {
		return fmt.Errorf("创建账户失败: %v", err)
	}

	logging.Infof("账户创建成功: ID=%s, 币种=%s, 路径=%s",
		account.ID, account.CoinSymbol, account.DerivationPath)
	return nil
}

func (r *REPL) handleAccountList(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("用法: account list  <CoinSymbol>")
	}
	coinSymbol := args[0]
	logging.Debugf("CoinSymbol is %s", coinSymbol)
	accountList, err := r.accountMgr.GetAccountsByCoin(coin.CoinType(coinSymbol, true))
	if err != nil {
		return err
	}
	fmt.Println(r.template.AccountList(accountList))
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
