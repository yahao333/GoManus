package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

// LLMSettings LLM配置
type LLMSettings struct {
	Model          string  `mapstructure:"model"`
	BaseURL        string  `mapstructure:"base_url"`
	APIKey         string  `mapstructure:"api_key"`
	MaxTokens      int     `mapstructure:"max_tokens"`
	MaxInputTokens *int    `mapstructure:"max_input_tokens"`
	Temperature    float64 `mapstructure:"temperature"`
	APIType        string  `mapstructure:"api_type"`
	APIVersion     string  `mapstructure:"api_version"`
}

// ProxySettings 代理配置
type ProxySettings struct {
	Server   string `mapstructure:"server"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// SearchSettings 搜索配置
type SearchSettings struct {
	Engine          string   `mapstructure:"engine"`
	FallbackEngines []string `mapstructure:"fallback_engines"`
	RetryDelay      int      `mapstructure:"retry_delay"`
	MaxRetries      int      `mapstructure:"max_retries"`
	Lang            string   `mapstructure:"lang"`
	Country         string   `mapstructure:"country"`
}

// BrowserSettings 浏览器配置
type BrowserSettings struct {
	Headless            bool          `mapstructure:"headless"`
	DisableSecurity     bool          `mapstructure:"disable_security"`
	ExtraChromiumArgs   []string      `mapstructure:"extra_chromium_args"`
	ChromeInstancePath  string        `mapstructure:"chrome_instance_path"`
	WssURL              string        `mapstructure:"wss_url"`
	CDPURL              string        `mapstructure:"cdp_url"`
	Proxy               *ProxySettings  `mapstructure:"proxy"`
	MaxContentLength    int           `mapstructure:"max_content_length"`
}

// SandboxSettings 沙盒配置
type SandboxSettings struct {
	UseSandbox     bool   `mapstructure:"use_sandbox"`
	Image          string `mapstructure:"image"`
	WorkDir        string `mapstructure:"work_dir"`
	MemoryLimit    string `mapstructure:"memory_limit"`
	CPULimit       float64 `mapstructure:"cpu_limit"`
	Timeout        int    `mapstructure:"timeout"`
	NetworkEnabled bool   `mapstructure:"network_enabled"`
}

// DaytonaSettings Daytona配置
type DaytonaSettings struct {
	Enabled            bool   `mapstructure:"enabled"`
	DaytonaAPIKey      string `mapstructure:"daytona_api_key"`
	DaytonaServerURL   string `mapstructure:"daytona_server_url"`
	DaytonaTarget      string `mapstructure:"daytona_target"`
	SandboxImageName   string `mapstructure:"sandbox_image_name"`
	SandboxEntrypoint  string `mapstructure:"sandbox_entrypoint"`
	VNCPassword        string `mapstructure:"vnc_password"`
}

// MCPServerConfig MCP服务器配置
type MCPServerConfig struct {
	Type    string   `mapstructure:"type"`
	URL     string   `mapstructure:"url"`
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}

// MCPSettings MCP配置
type MCPSettings struct {
	ServerReference string                    `mapstructure:"server_reference"`
	Servers         map[string]MCPServerConfig  `mapstructure:"servers"`
}

// RunflowSettings 工作流配置
type RunflowSettings struct {
	UseDataAnalysisAgent bool `mapstructure:"use_data_analysis_agent"`
}

// AppConfig 应用配置
type AppConfig struct {
	LLM          map[string]LLMSettings  `mapstructure:"llm"`
	Sandbox      *SandboxSettings        `mapstructure:"sandbox"`
	BrowserConfig *BrowserSettings       `mapstructure:"browser"`
	SearchConfig *SearchSettings         `mapstructure:"search"`
	MCPConfig    *MCPSettings            `mapstructure:"mcp"`
	RunflowConfig *RunflowSettings       `mapstructure:"runflow"`
	DaytonaConfig *DaytonaSettings       `mapstructure:"daytona"`
}

// Config 全局配置单例
type Config struct {
	viper   *viper.Viper
	config  *AppConfig
	mu      sync.RWMutex
}

var (
	instance *Config
	once     sync.Once
)

// GetConfig 获取配置实例
func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			viper: viper.New(),
		}
		instance.init()
	})
	return instance
}

// init 初始化配置
func (c *Config) init() {
	// 设置配置文件名和路径
	c.viper.SetConfigName("config")
	c.viper.SetConfigType("toml")
	
	// 添加配置路径
	c.viper.AddConfigPath("./config")
	c.viper.AddConfigPath("../config")
	c.viper.AddConfigPath(".")
	
	// 设置环境变量前缀
	c.viper.SetEnvPrefix("GOMANUS")
	c.viper.AutomaticEnv()
	
	// 读取配置文件
	if err := c.viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，尝试读取示例配置
		c.viper.SetConfigName("config.example")
		if err := c.viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("无法读取配置文件: %w", err))
		}
	}
	
	// 解析配置
	c.parseConfig()
}

// parseConfig 解析配置
func (c *Config) parseConfig() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	var appConfig AppConfig
	if err := c.viper.Unmarshal(&appConfig); err != nil {
		panic(fmt.Errorf("无法解析配置文件: %w", err))
	}
	
	c.config = &appConfig
}

// Reload 重新加载配置
func (c *Config) Reload() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if err := c.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("重新加载配置文件失败: %w", err)
	}
	
	var appConfig AppConfig
	if err := c.viper.Unmarshal(&appConfig); err != nil {
		return fmt.Errorf("重新解析配置文件失败: %w", err)
	}
	
	c.config = &appConfig
	return nil
}

// GetLLMSettings 获取LLM配置
func (c *Config) GetLLMSettings(name string) (LLMSettings, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.config == nil || c.config.LLM == nil {
		return LLMSettings{}, false
	}
	
	settings, ok := c.config.LLM[name]
	return settings, ok
}

// GetDefaultLLMSettings 获取默认LLM配置
func (c *Config) GetDefaultLLMSettings() LLMSettings {
	settings, ok := c.GetLLMSettings("default")
	if !ok {
		// 返回默认配置
		return LLMSettings{
			Model:       "gpt-4o",
			BaseURL:     "https://api.openai.com/v1",
			APIKey:      "",
			MaxTokens:   4096,
			Temperature: 0.7,
			APIType:     "openai",
			APIVersion:  "",
		}
	}
	return settings
}

// GetSandboxSettings 获取沙盒配置
func (c *Config) GetSandboxSettings() *SandboxSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.config == nil {
		return nil
	}
	return c.config.Sandbox
}

// GetBrowserSettings 获取浏览器配置
func (c *Config) GetBrowserSettings() *BrowserSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.config == nil {
		return nil
	}
	return c.config.BrowserConfig
}

// GetSearchSettings 获取搜索配置
func (c *Config) GetSearchSettings() *SearchSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.config == nil {
		return nil
	}
	return c.config.SearchConfig
}

// GetMCPSettings 获取MCP配置
func (c *Config) GetMCPSettings() *MCPSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.config == nil {
		return nil
	}
	return c.config.MCPConfig
}

// GetRunflowSettings 获取工作流配置
func (c *Config) GetRunflowSettings() *RunflowSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.config == nil {
		return nil
	}
	return c.config.RunflowConfig
}

// GetDaytonaSettings 获取Daytona配置
func (c *Config) GetDaytonaSettings() *DaytonaSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.config == nil {
		return nil
	}
	return c.config.DaytonaConfig
}

// GetWorkspaceRoot 获取工作空间根目录
func (c *Config) GetWorkspaceRoot() string {
	execPath, err := os.Getwd()
	if err != nil {
		return "./workspace"
	}
	return filepath.Join(execPath, "workspace")
}

// GetProjectRoot 获取项目根目录
func (c *Config) GetProjectRoot() string {
	execPath, err := os.Getwd()
	if err != nil {
		return "."
	}
	return execPath
}