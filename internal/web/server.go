package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/palagend/slowmade/pkg/logging"
	"go.uber.org/zap"
)

// Server 配置结构体
type Config struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// Server 表示 Web 服务器实例
type Server struct {
	config      *Config
	httpServer  *http.ServeMux
	logger      *zap.Logger
	middlewares []Middleware
}

// Middleware 定义中间件函数类型
type Middleware func(http.Handler) http.Handler

// NewServer 创建新的 Web 服务器实例
func NewServer() *Server {
	return &Server{
		config:      &Config{Host: "localhost", Port: 8080, Mode: "debug"},
		httpServer:  http.NewServeMux(),
		logger:      logging.Get(),
		middlewares: make([]Middleware, 0),
	}
}

// Host 设置服务器主机
func (s *Server) Host(host string) *Server {
	s.config.Host = host
	return s
}

// Port 设置服务器端口
func (s *Server) Port(port int) *Server {
	s.config.Port = port
	return s
}

// Mode 设置运行模式
func (s *Server) Mode(mode string) *Server {
	s.config.Mode = mode
	return s
}

// Use 添加中间件
func (s *Server) Use(middleware Middleware) *Server {
	s.middlewares = append(s.middlewares, middleware)
	return s
}

// Start 启动 Web 服务器
func (s *Server) Start() error {
	// 设置路由
	s.setupRoutes()

	// 应用中间件
	handler := s.applyMiddlewares(s.httpServer)

	// 创建 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 优雅关闭处理
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// 启动服务器
	go func() {
		s.logger.Info("Starting web server",
			zap.String("address", addr),
			zap.String("mode", s.config.Mode))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Failed to start server", zap.Error(err))
			panic(err)
		}
	}()

	<-stopChan
	s.logger.Info("Received shutdown signal, gracefully shutting down...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown failed", zap.Error(err))
		return err
	}

	s.logger.Info("Server stopped gracefully")
	return nil
}

// setupRoutes 设置路由处理
func (s *Server) setupRoutes() {
	// 健康检查端点
	s.httpServer.HandleFunc("/health", s.healthHandler)
	s.httpServer.HandleFunc("/api/v1/status", s.statusHandler)
	s.httpServer.HandleFunc("/api/v1/info", s.infoHandler)
	s.httpServer.HandleFunc("/", s.indexHandler)
}

// applyMiddlewares 应用中间件栈
func (s *Server) applyMiddlewares(handler http.Handler) http.Handler {
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		handler = s.middlewares[i](handler)
	}
	return handler
}

// 路由处理函数
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s", "service": "slowmade"}`,
		time.Now().Format(time.RFC3339))
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
        "status": "ok",
        "version": "1.0.0",
        "timestamp": "%s",
        "service": "slowmade",
        "mode": "%s"
    }`, time.Now().Format(time.RFC3339), s.config.Mode)
}

func (s *Server) infoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
        "name": "Slowmade Web Server",
        "version": "1.0.0",
        "description": "A secure cryptocurrency wallet service",
        "endpoints": [
            {"path": "/health", "method": "GET", "description": "Health check"},
            {"path": "/api/v1/status", "method": "GET", "description": "Service status"},
            {"path": "/api/v1/info", "method": "GET", "description": "Service information"}
        ]
    }`)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Slowmade Web Server</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container { 
            background: white; 
            padding: 3rem;
            border-radius: 15px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            max-width: 600px;
            text-align: center;
        }
        h1 { 
            color: #333;
            margin-bottom: 1rem;
            font-size: 2.5rem;
        }
        p { 
            color: #666;
            margin-bottom: 2rem;
            line-height: 1.6;
        }
        .endpoints { 
            text-align: left;
            margin: 2rem 0;
        }
        .endpoint { 
            background: #f8f9fa;
            padding: 1rem;
            margin: 0.5rem 0;
            border-radius: 8px;
            border-left: 4px solid #667eea;
        }
        .status { 
            color: #28a745;
            font-weight: bold;
            margin-top: 1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Slowmade Web Server</h1>
        <p>Secure cryptocurrency wallet service is running successfully!</p>
        
        <div class="endpoints">
            <div class="endpoint">
                <strong>Health Check:</strong> <a href="/health">/health</a>
            </div>
            <div class="endpoint">
                <strong>Status API:</strong> <a href="/api/v1/status">/api/v1/status</a>
            </div>
            <div class="endpoint">
                <strong>Service Info:</strong> <a href="/api/v1/info">/api/v1/info</a>
            </div>
        </div>
        
        <div class="status">Server is running on %s:%d</div>
    </div>
</body>
</html>
    `, s.config.Host, s.config.Port)
}
