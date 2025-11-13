package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yahao333/GoManus/pkg/agent"
	"github.com/yahao333/GoManus/pkg/logger"
	"go.uber.org/zap"
)

const (
	// Version 应用版本
	Version = "0.1.0"
	// BuildTime 构建时间
	BuildTime = "2025-11-13"
	// GitCommit Git提交哈希
	GitCommit = "unknown"
)

func main() {
	// 解析命令行参数
	var (
		prompt   string
		showVer  bool
	)
	flag.StringVar(&prompt, "prompt", "", "输入提示")
	flag.BoolVar(&showVer, "version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息
	if showVer {
		fmt.Printf("GoManus v%s\n", Version)
		fmt.Printf("构建时间: %s\n", BuildTime)
		fmt.Printf("Git提交: %s\n", GitCommit)
		os.Exit(0)
	}

	// 初始化日志
	if err := logger.InitLogger("logs/gomanus.log", zap.InfoLevel); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("GoManus 启动")

	// 获取用户输入
	if prompt == "" {
		fmt.Print("请输入您的提示: ")
		if _, err := fmt.Scanln(&prompt); err != nil {
			logger.Error("读取用户输入失败", zap.Error(err))
			os.Exit(1)
		}
	}

	if prompt == "" {
		logger.Warn("空提示提供")
		os.Exit(0)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("收到中断信号，正在关闭...")
		cancel()
	}()

	// 创建Manus智能体
	manus, err := agent.NewManus()
	if err != nil {
		logger.Error("创建Manus智能体失败", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("处理您的请求...")

	// 运行智能体
	if err := manus.Run(ctx, prompt); err != nil {
		logger.Error("运行智能体失败", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("请求处理完成")
}