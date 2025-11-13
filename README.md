# 项目结构

```bash
slowmade/
├── main.go                 # 程序入口
├── cmd/                   # Cobra命令定义
│   ├── root.go           # 根命令
│   ├── interactive.go    # 交互式模式
│   ├── create.go         # 创建钱包
│   ├── restore.go        # 恢复钱包
│   ├── list.go           # 列出钱包
│   ├── info.go           # 查看详情
│   ├── addcoin.go        # 添加币种
│   ├── address.go        # 生成地址
│   └── transfer.go       # 转账操作
├── internal/
│   ├── mvc/              # MVC架构实现
│   │   ├── controllers/  # 控制器层
│   │   ├── models/       # 模型层(充血模型)
│   │   ├── services/     # 服务层
│   │   └── views/        # 视图层(模板渲染)
│   │            ├── templates/           # 嵌入式模板目录
│   │            │   ├── wallet_created.tmpl
│   │            │   ├── wallet_list.tmpl
│   │            │   ├── wallet_info.tmpl
│   │            │   ├── address_qr.tmpl
│   │            │   └── transaction.tmpl
│   │            ├── template_renderer.go # 模板引擎
│   │            └── defaults.go          # 默认模板
│   ├── storage/                          # 存储模块
│   │
│   └── config/           
│             └── manager.go              # 配置模块
├── config/                               # 配置文件
└── templates/                            # 文本模板
```
