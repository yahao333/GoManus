package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yahao333/GoManus/internal/agent"
	"github.com/yahao333/GoManus/internal/logger"
	"github.com/yahao333/GoManus/internal/plugin"
	"github.com/yahao333/GoManus/pkg/config"
	"go.uber.org/zap"
)

const (
	// Version 应用版本
	Version = "0.2.0"
	// BuildTime 构建时间
	BuildTime = "2025-11-13"
	// GitCommit Git提交哈希
	GitCommit = "unknown"
)

var (
	cfgFile string
	verbose bool

	rootCmd = &cobra.Command{
		Use:   "gomanus",
		Short: "GoManus - AI Agent 框架",
		Long: `GoManus 是一个基于 Go 的多智能体 AI 框架，
支持工具调用、任务规划和本地执行。`,
		Version: fmt.Sprintf("%s (build: %s, commit: %s)", Version, BuildTime, GitCommit),
	}

	runCmd = &cobra.Command{
		Use:   "run [prompt]",
		Short: "运行智能体任务",
		Long:  `运行一个智能体任务，可以指定提示词或通过交互方式输入`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  runAgent,
	}

	configCmd = &cobra.Command{
		Use:   "config",
		Short: "配置管理命令",
		Long:  `管理 GoManus 的配置文件`,
	}

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "初始化配置文件",
		Long:  `创建默认的配置文件到 ~/.gomanus/config.toml`,
		RunE:  initConfig,
	}

	validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "验证配置文件",
		Long:  `验证配置文件的正确性`,
		RunE:  validateConfig,
	}

	pluginCmd = &cobra.Command{
		Use:   "plugin",
		Short: "插件管理命令",
		Long:  `管理 GoManus 的插件系统`,
	}

	pluginListCmd = &cobra.Command{
		Use:   "list",
		Short: "列出已加载的插件",
		Long:  `显示当前已加载的所有插件信息`,
		RunE:  listPlugins,
	}

	pluginLoadCmd = &cobra.Command{
		Use:   "load [plugin_path]",
		Short: "加载插件",
		Long:  `加载指定的插件文件`,
		Args:  cobra.ExactArgs(1),
		RunE:  loadPlugin,
	}

	pluginUnloadCmd = &cobra.Command{
		Use:   "unload [plugin_name]",
		Short: "卸载插件",
		Long:  `卸载指定的插件`,
		Args:  cobra.ExactArgs(1),
		RunE:  unloadPlugin,
	}

	// 简化的直接运行模式（兼容旧版本）
	directCmd = &cobra.Command{
		Use:   "direct [prompt]",
		Short: "直接运行模式",
		Long:  `使用标志位参数直接运行，兼容旧版本使用方式`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  runDirect,
	}
)

func init() {
	cobra.OnInitialize(initLogger)

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: ~/.gomanus/config.toml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细日志输出")

	// 添加子命令
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(directCmd)
	rootCmd.AddCommand(pluginCmd)

	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(validateCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginLoadCmd)
	pluginCmd.AddCommand(pluginUnloadCmd)
}

func main() {
	// 如果没有参数，显示帮助信息
	if len(os.Args) == 1 {
		rootCmd.Help()
		return
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initLogger() {
	logLevel := zap.InfoLevel
	if verbose {
		logLevel = zap.DebugLevel
	}

	if err := logger.InitLogger("", logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
}

// runAgent 运行智能体任务（新的命令模式）
func runAgent(cmd *cobra.Command, args []string) error {
	var prompt string

	if len(args) > 0 {
		prompt = args[0]
	} else {
		fmt.Print("请输入您的任务描述: ")
		if _, err := fmt.Scanln(&prompt); err != nil {
			return fmt.Errorf("读取输入失败: %w", err)
		}
	}

	if prompt == "" {
		return fmt.Errorf("任务描述不能为空")
	}

	logger.Info("开始运行智能体任务", zap.String("prompt", prompt))

	// 创建Manus智能体
	manus, err := agent.NewManus()
	if err != nil {
		return fmt.Errorf("创建智能体失败: %w", err)
	}

	// 运行任务
	ctx := cmd.Context()
	if err := manus.Run(ctx, prompt); err != nil {
		return fmt.Errorf("运行任务失败: %w", err)
	}

	logger.Info("任务完成")
	return nil
}

// runDirect 直接运行模式（兼容旧版本）
func runDirect(cmd *cobra.Command, args []string) error {
	var prompt string

	if len(args) > 0 {
		prompt = args[0]
	} else {
		fmt.Print("请输入您的提示: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			prompt = scanner.Text()
		} else {
			return fmt.Errorf("读取用户输入失败: %v", scanner.Err())
		}
	}

	if prompt == "" {
		return fmt.Errorf("空提示提供")
	}

	// 加载配置
	configPath := getConfigPath()
	if _, err := config.LoadConfig(configPath); err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号 - 使用更大的缓冲区避免信号丢失
	sigChan := make(chan os.Signal, 10)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 设置信号处理goroutine
	go func() {
		sigCount := 0
		for {
			select {
			case sig := <-sigChan:
				sigCount++
				logger.Info("收到中断信号，正在关闭...",
					zap.String("signal", sig.String()),
					zap.Int("count", sigCount))
				cancel()
				// 第一次中断后，第二次强制退出
				if sigCount >= 2 {
					logger.Info("收到多次中断信号，强制退出")
					os.Exit(1)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	logger.Info("GoManus 启动")
	logger.Info("处理您的请求...")

	// 创建Manus智能体
	manus, err := agent.NewManus()
	if err != nil {
		return fmt.Errorf("创建Manus智能体失败: %w", err)
	}

	// 运行智能体
	if err := manus.Run(ctx, prompt); err != nil {
		return fmt.Errorf("运行智能体失败: %w", err)
	}

	logger.Info("请求处理完成")
	return nil
}

func initConfig(cmd *cobra.Command, args []string) error {
	configPath := getConfigPath()

	// 检查配置文件是否已存在
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("配置文件已存在: %s", configPath)
	}

	// 创建配置目录
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 生成默认配置
	defaultConfig := `# GoManus 配置文件

[llm]
default = "openai"

[llm.providers.openai]
api_type = "openai"
api_key = "your-openai-api-key"
base_url = "https://api.openai.com/v1"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7

[llm.providers.azure]
api_type = "azure"
api_key = "your-azure-api-key"
base_url = "https://your-resource.openai.azure.com"
api_version = "2024-02-15-preview"
model = "gpt-4"
max_tokens = 2000
temperature = 0.7

[agent]
max_steps = 50
max_observe = 10000

[tools.python]
enabled = true
timeout = "30s"

[tools.browser]
enabled = true
headless = true

[tools.search]
enabled = true
engine = "duckduckgo"

[plugins]
enabled = true
directories = ["./plugins", "~/.gomanus/plugins"]
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

[mcp]
enabled = false

# MCP服务器配置示例
# [mcp.servers.weather]
# type = "sse"
# url = "http://localhost:8080/weather/sse"
#
# [mcp.servers.filesystem]
# type = "stdio"
# command = "mcp-server-filesystem"
# args = ["--path", "/home/user/documents"]
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("配置文件已创建: %s\n", configPath)
	fmt.Println("请编辑配置文件并设置您的API密钥")

	return nil
}

func validateConfig(cmd *cobra.Command, args []string) error {
	configPath := getConfigPath()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("配置文件验证失败: %w", err)
	}

	fmt.Println("配置文件验证通过")
	return nil
}

func getConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "config.toml"
	}

	return filepath.Join(home, ".gomanus", "config.toml")
}

// listPlugins 列出已加载的插件
func listPlugins(cmd *cobra.Command, args []string) error {
	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}

	if !cfg.Plugins.Enabled {
		fmt.Println("插件系统已禁用")
		return nil
	}

	// 创建插件管理器
	pluginManager := plugin.NewManager()
	loader := plugin.NewLoader(pluginManager, cfg.Plugins.Directories, cfg.Plugins.ManifestFile)

	// 加载所有插件
	if err := loader.LoadAllPlugins(); err != nil {
		fmt.Printf("加载插件失败: %v\n", err)
	}

	// 列出插件
	plugins := loader.GetLoadedPlugins()
	if len(plugins) == 0 {
		fmt.Println("当前没有加载任何插件")
		return nil
	}

	fmt.Println("已加载的插件:")
	for _, p := range plugins {
		fmt.Printf("  - %s (v%s): %s\n", p.Name, p.Version, p.Description)
	}

	return nil
}

// loadPlugin 加载插件
func loadPlugin(cmd *cobra.Command, args []string) error {
	pluginPath := args[0]

	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}

	if !cfg.Plugins.Enabled {
		return fmt.Errorf("插件系统已禁用")
	}

	// 创建插件管理器
	pluginManager := plugin.NewManager()

	// 加载插件
	if err := pluginManager.LoadPlugin(pluginPath); err != nil {
		return fmt.Errorf("加载插件失败: %w", err)
	}

	fmt.Printf("插件加载成功: %s\n", pluginPath)
	return nil
}

// unloadPlugin 卸载插件
func unloadPlugin(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}

	if !cfg.Plugins.Enabled {
		return fmt.Errorf("插件系统已禁用")
	}

	// 创建插件管理器
	pluginManager := plugin.NewManager()

	// 卸载插件
	if err := pluginManager.UnloadPlugin(pluginName); err != nil {
		return fmt.Errorf("卸载插件失败: %w", err)
	}

	fmt.Printf("插件卸载成功: %s\n", pluginName)
	return nil
}
