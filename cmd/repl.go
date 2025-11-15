package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/palagend/slowmade/internal/core"
	"github.com/palagend/slowmade/internal/repl"
	"github.com/palagend/slowmade/pkg/i18n"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// replCmd represents the repl command
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start interactive REPL session",
	Long:  `Start the interactive Read-Eval-Print Loop session for wallet operations`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		setupSignalHandler(cancel)

		// 初始化依赖
		if err := i18n.Init("config.toml"); err != nil {
			fmt.Printf("Failed to initialize i18n: %v\n", err)
			os.Exit(1)
		}

		walletManager := core.NewWalletManager()
		session := repl.NewSession(walletManager)

		// 启动REPL会话，完成后退出程序
		if err := session.Start(ctx); err != nil {
			logging.Get().Error("REPL session failed", zap.Error(err))
			os.Exit(1)
		}

		// REPL会话正常结束，退出程序
		logging.Get().Info("REPL session completed")
		os.Exit(0)
	},
}

func setupSignalHandler(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logging.Get().Info("Received termination signal", zap.String("signal", sig.String()))
		cancel()
		// 给REPL一点时间优雅退出
		<-time.After(100 * time.Millisecond)
		os.Exit(0)
	}()
}

func init() {
	rootCmd.AddCommand(replCmd)
}
