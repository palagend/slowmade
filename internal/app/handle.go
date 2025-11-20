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
		fmt.Printf("\n%s\n", view.Yellow("Mnemonic Phrase:"))
		fmt.Printf("%s\n\n", view.Green(mnemonic))
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

// 修改 handleHistory 函数使用会话历史记录
func (r *REPL) handleHistory(args []string) error {
	limit := 50 // 默认显示最近50条记录

	if len(args) > 0 {
		// 解析可选的限制参数
		if n, err := fmt.Sscanf(args[0], "%d", &limit); n != 1 || err != nil {
			return fmt.Errorf("invalid limit: %s. Usage: history [limit]", args[0])
		}
		if limit <= 0 {
			return fmt.Errorf("limit must be positive")
		}
	}

	if len(r.sessionHistory) == 0 {
		fmt.Println("No command history found in current session.")
		return nil
	}

	// 计算显示的起始索引
	start := 0
	if limit < len(r.sessionHistory) {
		start = len(r.sessionHistory) - limit
	}

	fmt.Printf("Command history (showing last %d of %d commands from current session):\n",
		len(r.sessionHistory)-start, len(r.sessionHistory))
	for i := start; i < len(r.sessionHistory); i++ {
		fmt.Printf("%5d: %s\n", i+1, r.sessionHistory[i])
	}
	return nil
}

func (r *REPL) handleVersion(args []string) error {
	fmt.Println(r.template.Version())
	return nil
}

func (r *REPL) handleAddressDerive(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("用法: address derive <账户ID> <找零地址/收款地址> [地址索引]")
	}

	accountID := args[0]
	changeType := uint32(1)
	if args[1] == "change" {
		changeType = 0
	}
	startIndex := uint32(0)
	if len(args) > 2 {
		if _, err := fmt.Sscanf(args[2], "%d", &startIndex); err != nil {
			return fmt.Errorf("无效的起始索引参数: %s", args[2])
		}
		if startIndex < 0 {
			return fmt.Errorf("起始索引不能为负数")
		}
	}

	// 检查钱包是否已解锁
	if r.walletMgr.IsLocked() {
		return fmt.Errorf("钱包已锁定，请先解锁钱包")
	}

	fmt.Println(r.template.Info(fmt.Sprintf("正在从账户 %s... 派生地址...", accountID[5:13])))

	// 派生地址
	addr, err := r.accountMgr.DeriveAddress(accountID, changeType, startIndex)
	if err != nil {
		return fmt.Errorf("派生地址失败: %v", err)
	}

	// 显示派生结果
	if addr.ChangeType == uint32(0) {
		fmt.Printf("%s (地址索引: %d，币种：%s， 类型： 收款地址)\n", addr.Address, startIndex, addr.CoinSymbol)
	}
	if addr.ChangeType == uint32(1) {
		fmt.Printf("%s (地址索引: %d，币种：%s， 类型： 找零地址)\n", addr.Address, startIndex, addr.CoinSymbol)
	}

	return nil
}

func (r *REPL) handleAddressList(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("用法: address list <账户ID> [显示数量]")
	}

	accountID := args[0]

	// 检查钱包是否已解锁
	if r.walletMgr.IsLocked() {
		return fmt.Errorf("钱包已锁定，请先解锁钱包")
	}

	fmt.Println(r.template.Info(fmt.Sprintf("正在获取账户 %s 的地址列表...", accountID)))

	// 获取地址列表
	addresses, err := r.accountMgr.GetAddresses(accountID)
	if err != nil {
		return fmt.Errorf("获取地址列表失败: %v", err)
	}

	if len(addresses) == 0 {
		fmt.Println("该账户尚未派生任何地址")
		return nil
	}

	// 显示地址列表
	fmt.Println(r.template.AddressList(addresses))
	return nil
}
