# GoManus

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> 🚀 **GoManus** - 高性能的多智能体 AI 系统

GoManus 是一个用 Go 语言构建的多智能体 AI 系统，集成了多种工具和 LLM 提供商，可以执行复杂的任务和工作流。

## ✨ 核心特性

- 🤖 **多智能体架构** - 支持多种智能体类型
- 🛠️ **丰富工具生态** - 集成 Python、浏览器、文件编辑等工具
- 🧠 **多 LLM 支持** - 支持 OpenAI、Azure、Ollama 等主流 LLM
- 🐳 **安全沙盒** - Docker 容器化的安全代码执行环境
- ⚡ **高性能** - Go 语言的并发特性带来出色性能

## 🚀 快速开始

### 前置要求

- **Go 1.21+** - [安装 Go](https://golang.org/dl/)
- **Python 3.8+** - 用于 Python 代码执行（可选）
- **Docker** - 用于沙盒环境（可选）

### 安装

```bash
# 克隆仓库
git clone https://github.com/yahao333/GoManus.git
cd GoManus

# 安装依赖
go mod download

# 创建配置文件
cp config/config.example.toml config/config.toml

# 配置你的 API 密钥
vim config/config.toml

# 运行
go run main.go
```

### 配置

编辑 `config/config.toml` 文件：

```toml
[llm.default]
model = "gpt-4o"
base_url = "https://api.openai.com/v1"
api_key = "your-api-key-here"  # 替换为你的 API 密钥
max_tokens = 4096
temperature = 0.7
api_type = "openai"
```

## 📖 使用方法

### 基础使用

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yahao333/GoManus/pkg/agent"
)

func main() {
    // 创建 Manus 智能体
    manus, err := agent.NewManus()
    if err != nil {
        log.Fatal(err)
    }

    // 运行任务
    err = manus.Run(context.Background(), "帮我创建一个简单的计算器网页")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("✅ 任务完成！")
}
```

### 命令行使用

```bash
# 交互模式
go run main.go

# 直接提供提示
go run main.go --prompt "分析深圳周末亲子游的热门景点"
```

## 🏗️ 架构

```
┌─────────────────────────────────────────┐
│              GoManus 架构             │
├─────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌───────┐ │
│  │  Agent  │  │  Tool   │  │  LLM   │ │
│  │  系统   │  │  生态   │  │  集成   │ │
│  └─────────┘  └─────────┘  └───────┘ │
│      │            │           │      │
│ ┌────────────┐ ┌──────────┐ ┌──────┐│
│ │   Manus    │ │PythonExc │ │OpenAI ││
│ │ToolCallAgent│ │BrowserUse│ │ Azure ││
│ │  BaseAgent │ │StrReplace│ │Ollama││
│ └────────────┘ └──────────┘ └──────┘│
└─────────────────────────────────────────┘
```

## 🛠️ 内置工具

| 工具名称 | 功能描述 |
|---------|---------|
| **PythonExecute** | Python 代码执行 |
| **BrowserUseTool** | 浏览器自动化 |
| **StrReplaceEditor** | 文件编辑 |
| **AskHuman** | 用户交互 |
| **SimpleSearch** | 网络搜索 |

## 📁 项目结构

```
GoManus/
├── main.go              # 主程序入口
├── config/              # 配置文件
│   ├── config.example.toml  # 配置模板
│   └── config.toml         # 实际配置
├── pkg/                 # 核心代码
│   ├── agent/         # 智能体实现
│   ├── tool/          # 工具实现
│   ├── llm/           # LLM 集成
│   └── logger/        # 日志系统
└── examples/            # 示例代码
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🤝 贡献

欢迎所有形式的贡献！请遵循以下步骤：

1. Fork 仓库
2. 创建分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m '✨ Add amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 **MIT 许可证**。

## 🙏 致谢

- [OpenManus](https://github.com/FoundationAgents/OpenManus) - 原始 Python 实现和灵感来源
- [go-openai](https://github.com/sashabaranov/go-openai) - OpenAI Go SDK

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给我们一个 Star！**

Made with ❤️ by GoManus Team

</div>
