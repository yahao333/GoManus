package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	globalConfig *Config
	configOnce   sync.Once
	configMutex  sync.RWMutex
)

// Config 主配置结构
type Config struct {
	LLM     LLMConfig     `mapstructure:"llm"`
	Agent   AgentConfig   `mapstructure:"agent"`
	Tools   ToolsConfig   `mapstructure:"tools"`
	Logging LoggingConfig `mapstructure:"logging"`
	Plugins PluginConfig  `mapstructure:"plugins"`
	Memory  MemoryConfig  `mapstructure:"memory"`
	MCP     MCPConfig     `mapstructure:"mcp"`
}

// LLMConfig LLM配置
type LLMConfig struct {
	Default   string                  `mapstructure:"default"`
	Providers map[string]LLMProvider `mapstructure:"providers"`
}

// LLMProvider LLM提供者配置
type LLMProvider struct {
	APIType     string  `mapstructure:"api_type"`
	APIKey      string  `mapstructure:"api_key"`
	BaseURL     string  `mapstructure:"base_url"`
	APIVersion  string  `mapstructure:"api_version"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// AgentConfig 智能体配置
type AgentConfig struct {
	MaxSteps  int `mapstructure:"max_steps"`
	MaxObserve int `mapstructure:"max_observe"`
}

// ToolsConfig 工具配置
type ToolsConfig struct {
	Python PythonToolConfig `mapstructure:"python"`
	Browser BrowserToolConfig `mapstructure:"browser"`
	Search  SearchToolConfig `mapstructure:"search"`
}

// PythonToolConfig Python工具配置
type PythonToolConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Timeout string `mapstructure:"timeout"`
}

// BrowserToolConfig 浏览器工具配置
type BrowserToolConfig struct {
	Enabled  bool `mapstructure:"enabled"`
	Headless bool `mapstructure:"headless"`
}

// SearchToolConfig 搜索工具配置
type SearchToolConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Engine  string `mapstructure:"engine"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	Directories  []string `mapstructure:"directories"`
	ManifestFile string   `mapstructure:"manifest_file"`
	AutoLoad     bool     `mapstructure:"auto_load"`
}

// MemoryConfig 内存配置
type MemoryConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Type        string `mapstructure:"type"`
	Path        string `mapstructure:"path"`
	MaxMessages int    `mapstructure:"max_messages"`
}

// MCPServerConfig MCP服务器配置
type MCPServerConfig struct {
	Type    string   `mapstructure:"type"`    // 服务器连接类型 (sse 或 stdio)
	URL     string   `mapstructure:"url"`     // SSE连接的URL
	Command string   `mapstructure:"command"` // stdio连接的命令
	Args    []string `mapstructure:"args"`    // stdio命令的参数
}

// MCPConfig MCP配置
type MCPConfig struct {
	Enabled bool                    `mapstructure:"enabled"`
	Servers map[string]MCPServerConfig `mapstructure:"servers"`
}

// LLMSettings LLM设置 (兼容旧接口)
type LLMSettings = LLMProvider

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置全局配置
	configMutex.Lock()
	globalConfig = &config
	configMutex.Unlock()

	return &config, nil
}

// GetConfig 获取全局配置实例
func GetConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.LLM.Default == "" {
		return fmt.Errorf("未设置默认LLM提供者")
	}

	if _, ok := c.LLM.Providers[c.LLM.Default]; !ok {
		return fmt.Errorf("默认LLM提供者 %s 未配置", c.LLM.Default)
	}

	// 验证每个提供者
	for name, provider := range c.LLM.Providers {
		if provider.APIKey == "" {
			return fmt.Errorf("LLM提供者 %s 未设置API密钥", name)
		}
		if provider.Model == "" {
			return fmt.Errorf("LLM提供者 %s 未设置模型", name)
		}
	}

	return nil
}

// GetDefaultLLMProvider 获取默认LLM提供者
func (c *Config) GetDefaultLLMProvider() LLMProvider {
	return c.LLM.Providers[c.LLM.Default]
}

// GetLLMProvider 获取指定LLM提供者
func (c *Config) GetLLMProvider(name string) (LLMProvider, bool) {
	provider, ok := c.LLM.Providers[name]
	return provider, ok
}

// GetLLMSettings 获取LLM设置 (兼容旧接口)
func (c *Config) GetLLMSettings(name string) (LLMProvider, bool) {
	return c.GetLLMProvider(name)
}

// GetDefaultLLMSettings 获取默认LLM设置 (兼容旧接口)
func (c *Config) GetDefaultLLMSettings() LLMProvider {
	return c.GetDefaultLLMProvider()
}

// GetWorkspaceRoot 获取工作空间根目录 (兼容旧接口)
func (c *Config) GetWorkspaceRoot() string {
	return "."
}

// setDefaults 设置默认配置
func setDefaults() {
	viper.SetDefault("llm.default", "openai")
	viper.SetDefault("llm.providers.openai.api_type", "openai")
	viper.SetDefault("llm.providers.openai.base_url", "https://api.openai.com/v1")
	viper.SetDefault("llm.providers.openai.model", "gpt-4")
	viper.SetDefault("llm.providers.openai.max_tokens", 2000)
	viper.SetDefault("llm.providers.openai.temperature", 0.7)

	viper.SetDefault("agent.max_steps", 50)
	viper.SetDefault("agent.max_observe", 10000)

	viper.SetDefault("tools.python.enabled", true)
	viper.SetDefault("tools.python.timeout", "30s")
	viper.SetDefault("tools.browser.enabled", true)
	viper.SetDefault("tools.browser.headless", true)
	viper.SetDefault("tools.search.enabled", true)
	viper.SetDefault("tools.search.engine", "duckduckgo")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "~/.gomanus/logs/gomanus.log")

	viper.SetDefault("plugins.enabled", true)
	viper.SetDefault("plugins.directories", []string{"./plugins", "~/.gomanus/plugins"})
	viper.SetDefault("plugins.manifest_file", "./plugins/manifest.json")
	viper.SetDefault("plugins.auto_load", true)

	viper.SetDefault("memory.enabled", true)
	viper.SetDefault("memory.type", "sqlite")
	viper.SetDefault("memory.path", "~/.gomanus/memory.db")
	viper.SetDefault("memory.max_messages", 10000)

	viper.SetDefault("mcp.enabled", false)
	viper.SetDefault("mcp.servers", map[string]interface{}{})
}