package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yahao333/GoManus/pkg/agent"
	"github.com/yahao333/GoManus/pkg/flow"
	"github.com/yahao333/GoManus/pkg/logger"
	"github.com/yahao333/GoManus/pkg/schema"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	if err := logger.InitLogger("logs/gomanus.log", zap.InfoLevel); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	logger.Info("GoManus 示例程序启动")

	// 示例1: 基础智能体
	fmt.Println("=== 示例1: 基础智能体 ===")
	basicExample()

	// 示例2: 工具调用智能体
	fmt.Println("\n=== 示例2: 工具调用智能体 ===")
	toolExample()

	// 示例3: Manus 智能体
	fmt.Println("\n=== 示例3: Manus 智能体 ===")
	manusExample()

	// 示例4: 工作流
	fmt.Println("\n=== 示例4: 规划工作流 ===")
	flowExample()

	logger.Info("所有示例执行完成")
}

// basicExample 基础智能体示例
func basicExample() {
	// 创建基础智能体
	basicAgent, err := agent.NewAgent(
		"BasicAgent",
		"基础智能体",
		"你是一个有帮助的AI助手。",
		"请回答用户的问题。",
	)
	if err != nil {
		logger.Error("创建基础智能体失败", zap.Error(err))
		return
	}

	ctx := context.Background()
	
	// 初始化智能体
	if err := basicAgent.Initialize(ctx); err != nil {
		logger.Error("初始化基础智能体失败", zap.Error(err))
		return
	}
	defer basicAgent.Cleanup(ctx)

	// 处理消息
	message := "你好，请介绍一下你自己。"
	response, err := basicAgent.ProcessMessage(ctx, schema.NewUserMessage(message))
	if err != nil {
		logger.Error("处理消息失败", zap.Error(err))
		return
	}

	if response.Content != nil {
		fmt.Printf("用户: %s\n", message)
		fmt.Printf("智能体: %s\n", *response.Content)
	}
}

// toolExample 工具调用智能体示例
func toolExample() {
	// 创建工具调用智能体
	toolAgent, err := agent.NewToolCallAgent(
		"ToolAgent",
		"工具调用智能体",
		"你是一个可以使用工具的AI助手。",
		"选择合适的工具来完成任务。",
	)
	if err != nil {
		logger.Error("创建工具调用智能体失败", zap.Error(err))
		return
	}

	ctx := context.Background()
	
	// 初始化智能体
	if err := toolAgent.Initialize(ctx); err != nil {
		logger.Error("初始化工具调用智能体失败", zap.Error(err))
		return
	}
	defer toolAgent.Cleanup(ctx)

	// 添加一些工具（这里只是示例，实际使用时需要实现具体的工具）
	fmt.Println("工具调用智能体已创建，可以执行各种工具操作")
}

// manusExample Manus 智能体示例
func manusExample() {
	// 创建 Manus 智能体
	manus, err := agent.NewManus()
	if err != nil {
		logger.Error("创建 Manus 智能体失败", zap.Error(err))
		return
	}

	ctx := context.Background()
	
	// 运行任务（这里使用简单的任务作为示例）
	task := "创建一个简单的 Python 脚本来计算斐波那契数列"
	fmt.Printf("任务: %s\n", task)
	
	if err := manus.Run(ctx, task); err != nil {
		logger.Error("运行 Manus 智能体失败", zap.Error(err))
		return
	}

	fmt.Println("Manus 智能体完成任务")
}

// flowExample 工作流示例
func flowExample() {
	// 创建规划工作流
	planningFlow := flow.NewPlanningFlow()
	
	ctx := context.Background()
	
	// 执行任务
	task := "创建一个简单的网页应用"
	fmt.Printf("工作流任务: %s\n", task)
	
	result, err := planningFlow.Execute(ctx, task)
	if err != nil {
		logger.Error("执行工作流失败", zap.Error(err))
		return
	}

	fmt.Printf("工作流结果: %s\n", result)
}