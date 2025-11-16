package cmd

import (
	"github.com/palagend/slowmade/internal/web"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	serveHost string
	servePort int
	serveMode string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server",
	Long: `Start the Slowmade web server that provides HTTP API endpoints
and a web interface for interacting with the cryptocurrency wallet service.

Examples:
  # Start server with default configuration
  slowmade serve
  
  # Start server on specific port
  slowmade serve --port 9090
  
  # Start server in production mode
  slowmade serve --mode release`,
	Run: func(cmd *cobra.Command, args []string) {
		// 创建服务器实例
		server := web.NewServer()

		// 应用配置（命令行参数优先，然后是配置文件）
		if serveHost != "" {
			server.Host(serveHost)
		} else if viper.IsSet("web.host") {
			server.Host(viper.GetString("web.host"))
		}

		if servePort != 0 {
			server.Port(servePort)
		} else if viper.IsSet("web.port") {
			server.Port(viper.GetInt("web.port"))
		}

		if serveMode != "" {
			server.Mode(serveMode)
		} else if viper.IsSet("web.mode") {
			server.Mode(viper.GetString("web.mode"))
		}

		// 添加中间件
		server.Use(server.RecoveryMiddleware)
		server.Use(server.CORSMiddleware)
		server.Use(server.LoggingMiddleware)

		// 启动服务器
		if err := server.Start(); err != nil {
			logging.Get().Error("Server failed to start", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// 命令行标志
	serveCmd.Flags().StringVar(&serveHost, "host", "", "Host to bind the server to")
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 0, "Port to listen on")
	serveCmd.Flags().StringVar(&serveMode, "mode", "", "Server mode (debug|release|test)")

	// 绑定 Viper 配置键
	viper.BindPFlag("web.host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("web.port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("web.mode", serveCmd.Flags().Lookup("mode"))

	// 设置默认配置
	viper.SetDefault("web.host", "localhost")
	viper.SetDefault("web.port", 8080)
	viper.SetDefault("web.mode", "debug")
}
