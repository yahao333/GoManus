package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/yahao333/GoManus/internal/agent"
	"github.com/yahao333/GoManus/internal/plugin"
	"github.com/yahao333/GoManus/pkg/config"
	"github.com/yahao333/GoManus/internal/logger"
	"go.uber.org/zap"
)

// PluginIntegrationExample 演示如何集成插件系统到Agent中
func PluginIntegrationExample() {
	fmt.Println("=== 插件系统集成示例 ===")

	// 初始化日志
	if err := logger.InitLogger("", zap.InfoLevel); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 加载配置
	configPath := "config.toml"
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 检查插件系统是否启用
	if !cfg.Plugins.Enabled {
		fmt.Println("插件系统已禁用")
		return
	}

	// 创建插件管理器
	pluginManager := plugin.NewManager()
	loader := plugin.NewLoader(pluginManager, cfg.Plugins.Directories, cfg.Plugins.ManifestFile)

	// 加载所有插件
	fmt.Println("正在加载插件...")
	if err := loader.LoadAllPlugins(); err != nil {
		fmt.Printf("加载插件失败: %v\n", err)
	}

	// 列出已加载的插件
	plugins := loader.GetLoadedPlugins()
	fmt.Printf("已加载 %d 个插件:\n", len(plugins))
	for _, p := range plugins {
		fmt.Printf("  - %s (v%s): %s\n", p.Name, p.Version, p.Description)
	}

	// 获取所有插件的工具
	tools := pluginManager.GetAllTools()
	fmt.Printf("\n插件提供了 %d 个工具:\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}

	// 获取所有插件的Agent
	agents := pluginManager.GetAllAgents()
	fmt.Printf("\n插件提供了 %d 个Agent:\n", len(agents))
	for _, agent := range agents {
		fmt.Printf("  - %s (%s): %s\n", agent.Name, agent.Type, agent.Description)
	}

	// 演示执行插件工具
	if len(plugins) > 0 {
		fmt.Println("\n=== 执行插件工具示例 ===")
		
		// 查找示例插件
		for _, p := range plugins {
			if p.Name == "sample" {
				// 执行加法计算
				result, err := pluginManager.ExecuteTool("sample", "calculator_add", map[string]interface{}{
					"a": 10,
					"b": 20,
				})
				if err != nil {
					fmt.Printf("执行工具失败: %v\n", err)
				} else {
					fmt.Printf("计算结果: 10 + 20 = %v\n", result)
				}

				// 执行字符串反转
				result, err = pluginManager.ExecuteTool("sample", "string_reverse", map[string]interface{}{
					"text": "Hello GoManus!",
				})
				if err != nil {
					fmt.Printf("执行工具失败: %v\n", err)
				} else {
					fmt.Printf("字符串反转结果: %v\n", result)
				}
				break
			}
		}
	}

	fmt.Println("\n=== 将插件工具集成到Agent中 ===")

	// 创建Manus智能体
	manus, err := agent.NewManus()
	if err != nil {
		log.Fatalf("创建智能体失败: %v", err)
	}

	// 将插件工具添加到Agent的工具集中
	// 这里需要修改Agent的代码来支持动态工具加载
	// 为了演示，我们假设Agent已经支持插件工具
	ctx := context.Background()
	prompt := "使用插件工具计算 15 乘以 3，然后反转结果字符串"

	fmt.Printf("运行任务: %s\n", prompt)
	if err := manus.Run(ctx, prompt); err != nil {
		fmt.Printf("运行任务失败: %v\n", err)
	}

	fmt.Println("\n=== 插件系统集成完成 ===")
}

// CreatePluginExample 创建示例插件
func CreatePluginExample() {
	fmt.Println("=== 创建示例插件 ===")

	// 创建插件目录
	pluginDir := "plugins"
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		log.Fatalf("创建插件目录失败: %v", err)
	}

	// 创建插件清单文件
	manifestContent := `{
  "plugins": [
    {
      "name": "sample",
      "version": "1.0.0",
      "description": "示例插件，提供基础的计算工具",
      "author": "GoManus Team",
      "entry_point": "sample_plugin.so",
      "config": {
        "precision": 2,
        "timeout": 30
      },
      "dependencies": [],
      "enabled": true
    }
  ]
}`

	// 创建示例TOML配置文件
	configContent := `# GoManus 配置文件示例
[llm]
default = "openai"

[llm.providers.openai]
api_type = "openai"
api_key = "your-openai-api-key"
base_url = "https://api.openai.com/v1"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7

[agent]
max_steps = 50
max_observe = 10000

[tools.python]
enabled = true
timeout = "30s"

[plugins]
enabled = true
directories = ["./plugins"]
manifest_file = "./plugins/manifest.json"
auto_load = true

[logging]
level = "info"
file = "~/.gomanus/logs/gomanus.log"

[memory]
enabled = true
type = "sqlite"
path = "~/.gomanus/memory.db"
max_messages = 10000
`

	configPath := filepath.Join(pluginDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		log.Fatalf("创建配置文件失败: %v", err)
	}

	fmt.Printf("配置文件已创建: %s\n", configPath)

	manifestPath := filepath.Join(pluginDir, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		log.Fatalf("创建插件清单文件失败: %v", err)
	}

	fmt.Printf("插件清单文件已创建: %s\n", manifestPath)
	fmt.Println("请编译示例插件: go build -buildmode=plugin -o plugins/sample_plugin.so examples/plugin/sample_plugin.go")
	fmt.Println("然后使用配置文件运行: ./main run --config ./plugins/config.toml")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "create" {
		CreatePluginExample()
		return
	}

	PluginIntegrationExample()
}