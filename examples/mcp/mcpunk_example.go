package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/GoManus/pkg/config"
	"github.com/yourusername/GoManus/internal/mcp"
)

// MCPunk get_a_joke 工具使用示例
func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("config/mcpunk.toml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 创建MCP客户端
	mcpClient := mcp.NewMCPClients(logger.GetLogger())
	defer mcpClient.Cleanup()

	// 3. 连接到MCPunk服务器
	ctx := context.Background()
	if err := mcpClient.ConnectStdio(ctx, "mcpunk", cfg.MCP.Servers["mcpunk"]); err != nil {
		log.Fatalf("连接MCPunk服务器失败: %v", err)
	}

	// 4. 获取可用工具列表
	tools, err := mcpClient.ListTools(ctx, "mcpunk")
	if err != nil {
		log.Fatalf("获取工具列表失败: %v", err)
	}

	fmt.Println("可用工具:")
	for name, tool := range tools {
		fmt.Printf("- %s: %s\n", name, tool.Description)
	}

	// 5. 调用get_a_joke工具
	result, err := mcpClient.CallTool(ctx, "mcpunk", "get_a_joke", map[string]interface{}{
		// get_a_joke工具通常不需要参数
	})
	if err != nil {
		log.Fatalf("调用get_a_joke工具失败: %v", err)
	}

	fmt.Printf("\n笑话结果: %s\n", result)
}