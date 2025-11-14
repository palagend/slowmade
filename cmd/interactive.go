package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start interactive wallet session",
	Long:  `Start an interactive command-line session for wallet management`,
	Run: func(cmd *cobra.Command, args []string) {
		startInteractiveSession()
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

// 交互式会话状态
type interactiveSession struct {
	currentWalletID string
	scanner         *bufio.Scanner
}

func startInteractiveSession() {
	if walletController == nil {
		fmt.Println("[X] Wallet controller not initialized. Please check configuration.")
		return
	}

	session := &interactiveSession{
		scanner: bufio.NewScanner(os.Stdin),
	}

	fmt.Println("[O] Crypto Wallet CLI - Interactive Mode")
	fmt.Println("Type 'help' for available commands, 'exit' to quit")

	for {
		fmt.Print("\nwallet> ")
		if !session.scanner.Scan() {
			break
		}

		input := strings.TrimSpace(session.scanner.Text())
		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		session.handleCommand(input)
	}
}

func (s *interactiveSession) handleCommand(input string) {
	args := strings.Fields(input)
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "help", "?":
		s.showHelp()
	case "list":
		s.handleListCommand(args[1:])
	case "new": // 使用 "new" 替代 "create" 避免冲突
		s.handleNewCommand(args[1:])
	case "use":
		s.handleUseCommand(args[1:])
	case "info":
		s.handleInfoCommand(args[1:])
	case "restore":
		s.handleRestoreCommand(args[1:])
	case "backup":
		s.handleBackupCommand(args[1:])
	case "clear":
		fmt.Print("\033[H\033[2J") // 清屏
	default:
		fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", args[0])
	}
}

func (s *interactiveSession) showHelp() {
	helpText := `
Available Commands:

Wallet Management:
  new [name]           - Create a new wallet (interactive)
  list                 - List all wallets
  use <wallet_id>      - Select a wallet to work with
  info                 - Show current wallet info
  restore              - Restore wallet from mnemonic
  backup               - Backup current wallet data

General:
  help                 - Show this help message
  clear                - Clear screen
  exit                 - Exit interactive mode

Examples:
  new mywallet
  list
  use wallet_123abc
  info
  restore
  backup
`
	fmt.Println(helpText)
}

func (s *interactiveSession) handleNewCommand(args []string) {
	var walletName string
	if len(args) > 0 {
		walletName = args[0]
	} else {
		fmt.Print("Enter wallet name: ")
		if s.scanner.Scan() {
			walletName = strings.TrimSpace(s.scanner.Text())
		}
		if walletName == "" {
			fmt.Println("[X] Wallet name cannot be empty")
			return
		}
	}

	password, err := s.readPasswordWithConfirmation()
	if err != nil {
		fmt.Printf("[X] Error reading password: %v\n", err)
		return
	}

	// 使用 WalletController 创建钱包
	result, err := walletController.CreateWallet(walletName, password, cloak)
	if err != nil {
		fmt.Printf("[X] Error creating wallet: %v\n", err)
		return
	}

	fmt.Printf("[V] Wallet '%s' created successfully!\n", walletName)
	fmt.Println(result)
}

func (s *interactiveSession) handleListCommand(args []string) {
	result, err := walletController.ListWallets()
	if err != nil {
		fmt.Printf("[X] Error listing wallets: %v\n", err)
		return
	}

	if result == "" || result == "[]" {
		fmt.Println("[i] No wallets available. Use 'new' to make a new wallet.")
		return
	}

	fmt.Println(result)
}

func (s *interactiveSession) handleUseCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("[X] Please specify wallet ID. Usage: use <wallet_id>")
		return
	}

	walletID := args[0]

	// 验证钱包是否存在
	_, err := walletController.GetWalletInfo(walletID)
	if err != nil {
		fmt.Printf("[X] Wallet with ID '%s' not found or error: %v\n", walletID, err)
		return
	}

	s.currentWalletID = walletID
	fmt.Printf("[V] Selected wallet: %s\n", walletID)
}

func (s *interactiveSession) handleInfoCommand(args []string) {
	if s.currentWalletID == "" {
		fmt.Println("[X] No wallet selected. Use 'use <wallet_id>' first.")
		return
	}

	result, err := walletController.GetWalletInfo(s.currentWalletID)
	if err != nil {
		fmt.Printf("[X] Error getting wallet info: %v\n", err)
		return
	}

	fmt.Println(result)
}

func (s *interactiveSession) handleRestoreCommand(args []string) {
	var mnemonic, walletName string

	fmt.Print("Enter mnemonic phrase: ")
	if s.scanner.Scan() {
		mnemonic = strings.TrimSpace(s.scanner.Text())
	}
	if mnemonic == "" {
		fmt.Println("[X] Mnemonic cannot be empty")
		return
	}

	fmt.Print("Enter wallet name: ")
	if s.scanner.Scan() {
		walletName = strings.TrimSpace(s.scanner.Text())
	}
	if walletName == "" {
		fmt.Println("[X] Wallet name cannot be empty")
		return
	}

	password, err := s.readPasswordWithConfirmation()
	if err != nil {
		fmt.Printf("[X] Error reading password: %v\n", err)
		return
	}

	// 使用 WalletController 恢复钱包
	err = walletController.RestoreWallet(mnemonic, walletName, password)
	if err != nil {
		fmt.Printf("[X] Error restoring wallet: %v\n", err)
		return
	}

	fmt.Printf("[V] Wallet '%s' restored successfully!\n", walletName)
}

func (s *interactiveSession) handleBackupCommand(args []string) {
	if s.currentWalletID == "" {
		fmt.Println("[X] No wallet selected. Use 'use <wallet_id>' first.")
		return
	}

	walletInfo, err := walletController.GetWalletInfo(s.currentWalletID)
	if err != nil {
		fmt.Printf("[X] Error getting wallet info for backup: %v\n", err)
		return
	}

	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("wallet_backup_%s_%d.json", s.currentWalletID, timestamp)

	err = os.WriteFile(filename, []byte(walletInfo), 0600)
	if err != nil {
		fmt.Printf("[X] Error writing backup file: %v\n", err)
		return
	}

	fmt.Printf("[V] Wallet backed up to: %s\n", filename)
	fmt.Println("[!!!]  Keep this file secure! It contains sensitive information.")
}

func (s *interactiveSession) readPasswordWithConfirmation() (string, error) {
	fmt.Print("Set wallet password: ")
	password, err := s.readPassword()
	if err != nil {
		return "", err
	}

	fmt.Print("Confirm wallet password: ")
	confirmPassword, err := s.readPassword()
	if err != nil {
		return "", err
	}

	if password != confirmPassword {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

func (s *interactiveSession) readPassword() (string, error) {
	if s.scanner.Scan() {
		return strings.TrimSpace(s.scanner.Text()), nil
	}
	return "", s.scanner.Err()
}
