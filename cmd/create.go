package cmd

import (
	"fmt"

	"github.com/palagend/slowmade/internal/logging"
	"github.com/palagend/slowmade/internal/mvc/controllers"
	"github.com/palagend/slowmade/internal/mvc/services"
	"github.com/palagend/slowmade/internal/mvc/views"
	"github.com/palagend/slowmade/internal/storage"

	"github.com/spf13/cobra"
)

var (
	walletController *controllers.WalletController
	walletName       string
	password         string
	cloak            string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new HD wallet",
	Long:  `Create a new hierarchical deterministic (HD) cryptocurrency wallet`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 确保配置已初始化
		if appConfig == nil {
			return fmt.Errorf("配置未初始化，请检查初始化流程")
		}
		return initializeDependencies()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := walletController.CreateWallet(walletName, password, cloak)
		if err != nil {
			logging.Error("创建钱包失败", map[string]interface{}{
				"error": err,
				"name":  walletName,
			})
			return fmt.Errorf("创建钱包失败: %v", err)
		}

		logging.Info("钱包创建成功", map[string]interface{}{
			"name": walletName,
		})
		cmd.Println(result)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&walletName, "name", "n", "", "钱包名称 (必填)")
	createCmd.Flags().StringVarP(&password, "password", "p", "", "加密密码 (必填)")
	createCmd.Flags().StringVarP(&cloak, "cloak", "k", "", "助记词魔法密码")

	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(createCmd)
}

// initializeDependencies 初始化服务依赖
func initializeDependencies() error {
	logging.Debug("开始初始化服务依赖")

	// 获取配置
	cfg := GetConfig()
	if cfg == nil {
		return fmt.Errorf("配置不可用")
	}

	// 初始化存储层
	repo := storage.NewWalletRepository(cfg)
	logging.Debug("存储仓库初始化完成")

	// 初始化服务层
	cryptoService := services.NewCryptoService()
	walletService := services.NewWalletService(repo, cryptoService)
	logging.Debug("服务层初始化完成")

	// 初始化视图层
	renderer := views.NewTemplateRenderer(&cfg.Template)
	logging.Debug("视图渲染器初始化完成")

	// 调试信息
	renderer.PrintStatus()

	// 初始化控制器
	walletController = controllers.NewWalletController(walletService, renderer)
	logging.Info("所有服务依赖初始化完成")

	return nil
}
