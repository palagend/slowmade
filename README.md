# 项目结构

```bash
cipherkey/
├── cmd/                 # 控制器层
│   ├── root.go         # 根命令
│   ├── wallet.go       # 钱包管理
│   ├── address.go      # 地址管理
│   ├── qrcode.go       # 二维码生成
│   ├── transaction.go  # 交易操作
│   └── keys.go         # 密钥工具
├── internal/
│   ├── model/          # 模型层
│   │   ├── wallet.go
│   │   ├── address.go
│   │   └── coin/       # 币种工厂模式
│   │       ├── interface.go
│   │       ├── bitcoin.go
│   │       ├── ethereum.go
│   │       ├── litecoin.go
│   │       └── factory.go
│   ├── view/           # 视图层
│   │   ├── renderer.go
│   │   ├── qr_renderer.go
│   │   └── templates/
│   ├── service/        # 服务层
│   │   ├── wallet_service.go
│   │   ├── key_service.go
│   │   ├── qr_service.go
│   │   ├── security.go      # 安全内存管理
│   │   └── keystore.go      # 加密存储
│   └── crypto/         # 加密模块
│       ├── securemem.go
│       └── encryption.go
├── config/
│   └── config.go
├── templates/          # 用户模板目录
├── go.mod
├── main.go
└── README.md
```