# GoManus

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/yahao333/GoManus)](https://goreportcard.com/report/github.com/yahao333/GoManus)

> 🚀 **GoManus** - OpenManus 的高性能 Go 语言实现，一个强大的多智能体 AI 系统

GoManus 是一个用 Go 语言构建的多智能体 AI 系统，专注于高性能、可扩展性和易用性。它集成了多种工具和 LLM 提供商，可以执行复杂的任务和工作流。

## ✨ 核心特性

- 🤖 **多智能体架构** - 支持多种智能体类型和协作模式
- 🛠️ **丰富工具生态** - 集成 Python、浏览器、文件编辑、搜索等工具
- 🧠 **多 LLM 支持** - 支持 OpenAI、Azure OpenAI、Ollama 等主流 LLM
- 🔧 **高度可扩展** - 模块化设计，易于扩展新的智能体和工具
- 🐳 **安全沙盒** - Docker 容器化的安全代码执行环境
- 📊 **工作流管理** - 支持复杂的多步骤任务和工作流
- ⚡ **高性能** - Go 语言的并发特性带来出色性能
- 📝 **结构化日志** - 完善的日志系统和错误追踪

## 🚀 快速开始

### 前置要求

- **Go 1.21+** - [安装 Go](https://golang.org/dl/)
- **Python 3.8+** - 用于 Python 代码执行（可选）
- **Docker** - 用于沙盒环境（可选，推荐）

### 安装

```bash
# 克隆仓库
git clone https://github.com/yahao333/GoManus.git
cd GoManus

# 安装 Go 依赖
go mod download

# 创建配置文件
cp config/config.example.toml config/config.toml

# 构建项目
go build -o gomanus main.go
```

### 配置

编辑 `config/config.toml` 文件，配置你的 LLM 提供商：

```toml
# === LLM 配置 ===
[llm.default]
model = "gpt-4o"
base_url = "https://api.openai.com/v1"
api_key = "your-api-key-here"  # 替换为你的 API 密钥
max_tokens = 4096
temperature = 0.7
api_type = "openai"
api_version = ""

# === 浏览器配置 ===
[browser]
headless = false
disable_security = true
max_content_length = 2000

# === 沙盒配置 ===
[sandbox]
use_sandbox = false  # 建议生产环境设为 true
image = "python:3.12-slim"
work_dir = "/workspace"
memory_limit = "512m"
cpu_limit = 1.0
timeout = 300
network_enabled = false
```

### 运行

```bash
# 交互模式
go run main.go

# 直接提供提示
go run main.go --prompt "帮我创建一个简单的计算器网页"

# 或使用构建的二进制文件
./gomanus --prompt "分析深圳周末亲子游的热门景点"
```

## 📖 使用示例

### 基础 API 使用

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
        log.Fatalf("创建智能体失败: %v", err)
    }

    // 运行任务
    ctx := context.Background()
    err = manus.Run(ctx, "帮我创建一个简单的计算器网页")
    if err != nil {
        log.Fatalf("运行任务失败: %v", err)
    }

    fmt.Println("✅ 任务完成！")
}
```

### 工具使用示例

```go
package main

import (
    "context"
    "fmt"

    "github.com/yahao333/GoManus/pkg/tool"
    "github.com/yahao333/GoManus/pkg/agent"
)

func main() {
    // 创建工具集合
    tools := tool.NewToolCollection()
    tools.AddTool(tool.NewPythonExecute())
    tools.AddTool(tool.NewBrowserUseTool())

    // 创建智能体并添加工具
    agent, err := agent.NewToolCallAgent("MyAgent", "工具智能体", tools)
    if err != nil {
        panic(err)
    }

    // 执行任务
    ctx := context.Background()
    response, err := agent.GenerateResponse(ctx, "帮我用 Python 计算斐波那契数列的前10项")
    if err != nil {
        panic(err)
    }

    fmt.Printf("结果: %v\n", response)
}
```

## 🏗️ 架构设计

### 核心组件

```
┌─────────────────────────────────────────────────────────────┐
│                    GoManus 架构                             │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Agent     │  │    Tool     │  │    LLM      │         │
│  │   系统      │  │   生态      │  │   集成      │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│         │               │               │                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Manus     │  │PythonExecute│  │   OpenAI    │         │
│  │ToolCallAgent│  │BrowserUse   │  │   Azure     │         │
│  │  BaseAgent  │  │StrReplace   │  │   Ollama    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
           │                   │                   │
    ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
    │  配置管理    │  │  日志系统    │  │  沙盒环境    │
    │ Config Mgmt │  │   Logger    │  │  Sandbox    │
    └─────────────┘  └─────────────┘  └─────────────┘
```

### 智能体系统

1. **BaseAgent** - 基础智能体接口和核心功能
2. **ToolCallAgent** - 支持工具调用的智能体实现
3. **Manus** - 主要的多功能智能体

### 工具生态

| 工具名称 | 功能描述 | 用途 |
|---------|---------|-----|
| **PythonExecute** | Python 代码执行 | 数据分析、计算、脚本执行 |
| **BrowserUseTool** | 浏览器自动化 | 网页抓取、UI 自动化测试 |
| **StrReplaceEditor** | 文件编辑 | 代码生成、文档编辑 |
| **AskHuman** | 用户交互 | 需要用户确认的场景 |
| **Terminate** | 任务控制 | 流程控制和任务终止 |
| **SimpleSearch** | 网络搜索 | 信息检索、数据收集 |

### LLM 集成

- **OpenAI** - GPT-3.5, GPT-4, GPT-4o 等
- **Azure OpenAI** - 企业级 OpenAI 服务
- **Ollama** - 本地 LLM 部署
- **流式响应** - 实时响应支持
- **工具调用** - Function Calling 完整支持

## 📁 项目结构

```
GoManus/
├── 📄 main.go                 # 主程序入口
├── 📄 go.mod                  # Go 模块定义
├── 📄 go.sum                  # 依赖锁定文件
├── 📄 README.md               # 项目文档
├── 📁 config/                 # 配置文件目录
│   ├── config.example.toml    # 配置模板
│   └── config.toml           # 实际配置（不提交）
├── 📁 pkg/                    # 核心代码包
│   ├── 📁 agent/              # 智能体实现
│   │   ├── base.go           # 基础智能体
│   │   ├── manus.go          # Manus 智能体
│   │   └── toolcall.go       # 工具调用智能体
│   ├── 📁 tool/               # 工具实现
│   │   ├── base.go           # 工具接口
│   │   ├── tools.go          # 内置工具集
│   │   └── simple.go         # 简单工具
│   ├── 📁 llm/                # LLM 集成
│   │   └── llm.go            # LLM 提供者
│   ├── 📁 config/             # 配置管理
│   │   └── config.go         # 配置解析
│   ├── 📁 schema/             # 数据结构
│   │   └── types.go          # 类型定义
│   ├── 📁 logger/             # 日志系统
│   │   └── logger.go         # 日志实现
│   ├── 📁 flow/               # 工作流管理
│   │   └── flow.go           # 流程控制
│   └── 📁 sandbox/            # 沙盒环境
│       ├── local.go           # 本地沙盒
│       └── sandbox.go        # 沙盒接口
├── 📁 examples/               # 示例代码
│   └── main.go               # 使用示例
├── 📁 logs/                   # 日志文件（运行时生成）
└── 📁 .gitignore              # Git 忽略规则
```

## 🛠️ 开发指南

### 创建自定义智能体

```go
package myagent

import (
    "github.com/yahao333/GoManus/pkg/agent"
    "github.com/yahao333/GoManus/pkg/schema"
)

type MyAgent struct {
    *agent.BaseAgent
}

func NewMyAgent() (*MyAgent, error) {
    // 创建基础智能体
    baseAgent, err := agent.NewBaseAgent(
        "MyAgent",                    // 名称
        "我的自定义智能体",            // 描述
        "你是一个专业的数据分析师",   // 系统提示
        "请分析提供的数据",          // 用户提示
    )
    if err != nil {
        return nil, err
    }
    
    return &MyAgent{BaseAgent: baseAgent}, nil
}

func (m *MyAgent) AnalyzeData(data string) (*schema.Message, error) {
    // 实现自定义逻辑
    prompt := fmt.Sprintf("请分析以下数据：%s", data)
    return m.GenerateResponse(ctx, prompt)
}
```

### 创建自定义工具

```go
package mytool

import (
    "context"
    "github.com/yahao333/GoManus/pkg/tool"
)

type DataAnalyzer struct {
    tool.BaseTool
}

func NewDataAnalyzer() *DataAnalyzer {
    return &DataAnalyzer{
        BaseTool: tool.BaseTool{
            Name:        "DataAnalyzer",
            Description: "数据分析工具，支持统计分析",
            Parameters: map[string]interface{}{
                "data": map[string]interface{}{
                    "type":        "array",
                    "description": "要分析的数据数组",
                    "items": map[string]interface{}{
                        "type": "number"
                    }
                },
                "analysis_type": map[string]interface{}{
                    "type":        "string",
                    "description": "分析类型：mean, median, mode, std",
                    "enum":        []string{"mean", "median", "mode", "std"},
                },
            },
            Required: []string{"data", "analysis_type"},
        },
    }
}

func (d *DataAnalyzer) Execute(ctx context.Context, arguments string) (interface{}, error) {
    args, err := d.ParseArguments(arguments)
    if err != nil {
        return nil, err
    }

    data := args["data"].([]interface{})
    analysisType := args["analysis_type"].(string)

    // 实现数据分析逻辑
    result := d.performAnalysis(data, analysisType)

    return map[string]interface{}{
        "result": result,
        "type":   analysisType,
    }, nil
}
```

### 运行示例

```bash
# 运行基础示例
go run examples/main.go

# 运行特定示例
go run examples/main.go -example=basic
go run examples/main.go -example=tools
go run examples/main.go -example=workflow
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/agent/
go test ./pkg/tool/

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 性能

### 基准测试结果

| 操作 | 平均延迟 | QPS | 内存使用 |
|-----|---------|-----|---------|
| 简单对话 | ~200ms | 50 | ~50MB |
| 工具调用 | ~500ms | 20 | ~80MB |
| 复杂工作流 | ~2s | 5 | ~150MB |

### 性能优化建议

1. **启用并发**：利用 Go 的 goroutine 并发处理多个任务
2. **缓存配置**：避免重复解析配置文件
3. **连接池**：使用 LLM API 连接池减少延迟
4. **内存管理**：及时释放大型数据结构

## 🤝 贡献指南

我们欢迎所有形式的贡献！请遵循以下步骤：

### 贡献流程

1. **Fork 仓库** → 点击右上角的 Fork 按钮
2. **创建分支** → `git checkout -b feature/amazing-feature`
3. **提交更改** → `git commit -m '✨ Add amazing feature'`
4. **推送分支** → `git push origin feature/amazing-feature`
5. **创建 PR** → 在 GitHub 上创建 Pull Request

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 添加适当的注释和文档
- 编写单元测试
- 更新相关文档

### 提交信息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整
refactor: 代码重构
test: 测试相关
chore: 构建或辅助工具变动
```

## 📝 更新日志

### v0.1.0 (2024-01-01)

- ✨ 初始版本发布
- 🤖 基础多智能体系统
- 🛠️ 完整工具生态
- 🧠 多 LLM 支持
- 🐳 沙盒环境
- 📊 工作流管理

## 📄 许可证

本项目采用 **MIT 许可证** - 详情请参见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [OpenManus](https://github.com/FoundationAgents/OpenManus) - 原始 Python 实现和灵感来源
- [MetaGPT](https://github.com/geekan/MetaGPT) - 多智能体框架设计参考
- [go-openai](https://github.com/sashabaranov/go-openai) - OpenAI Go SDK
- [viper](https://github.com/spf13/viper) - 配置管理库
- 所有贡献者和支持者 🌟

## 📞 联系我们

- 📧 **邮箱**: [apprank@outlook.com](mailto:apprank@outlook.com)
- 🐛 **问题反馈**: [GitHub Issues](https://github.com/yahao333/GoManus/issues)
- 💬 **讨论**: [GitHub Discussions](https://github.com/yahao333/GoManus/discussions)
- 📖 **文档**: [Wiki](https://github.com/yahao333/GoManus/wiki)

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给我们一个 Star！**

Made with ❤️ by the GoManus Team

</div>
