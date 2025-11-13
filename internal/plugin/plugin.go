package plugin

import (
	"fmt"
	"plugin"
	"sync"

	"github.com/yahao333/GoManus/internal/schema"
)

// Plugin 定义插件接口
type Plugin interface {
	// 插件信息
	Name() string
	Version() string
	Description() string

	// 插件生命周期
	Init(config map[string]interface{}) error
	Start() error
	Stop() error

	// 插件功能
	GetTools() []schema.ToolDefinition
	ExecuteTool(name string, args map[string]interface{}) (interface{}, error)
	GetAgents() []AgentInfo
}

// AgentInfo 定义插件提供的Agent信息
type AgentInfo struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
}

// PluginMetadata 插件元数据
type PluginMetadata struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Config      map[string]interface{} `json:"config"`
	EntryPoint  string                 `json:"entry_point"`
}

// Manager 插件管理器
type Manager struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewManager 创建插件管理器
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
	}
}

// LoadPlugin 加载插件
func (m *Manager) LoadPlugin(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 打开插件文件
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// 查找插件构造函数
	symbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("plugin missing NewPlugin function: %w", err)
	}

	// 类型断言
	newPlugin, ok := symbol.(func() Plugin)
	if !ok {
		return fmt.Errorf("NewPlugin must be func() Plugin")
	}

	// 创建插件实例
	pluginInstance := newPlugin()

	// 初始化插件
	if err := pluginInstance.Init(nil); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// 启动插件
	if err := pluginInstance.Start(); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	// 添加到管理器
	m.plugins[pluginInstance.Name()] = pluginInstance

	return nil
}

// UnloadPlugin 卸载插件
func (m *Manager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// 停止插件
	if err := plugin.Stop(); err != nil {
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	// 从管理器中移除
	delete(m.plugins, name)

	return nil
}

// GetPlugin 获取插件
func (m *Manager) GetPlugin(name string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (m *Manager) ListPlugins() []PluginMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var metadata []PluginMetadata
	for _, p := range m.plugins {
		metadata = append(metadata, PluginMetadata{
			Name:        p.Name(),
			Version:     p.Version(),
			Description: p.Description(),
		})
	}

	return metadata
}

// GetAllTools 获取所有插件的工具
func (m *Manager) GetAllTools() []schema.ToolDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allTools []schema.ToolDefinition
	for _, p := range m.plugins {
		tools := p.GetTools()
		allTools = append(allTools, tools...)
	}

	return allTools
}

// ExecuteTool 执行插件工具
func (m *Manager) ExecuteTool(pluginName, toolName string, args map[string]interface{}) (interface{}, error) {
	plugin, err := m.GetPlugin(pluginName)
	if err != nil {
		return nil, err
	}

	return plugin.ExecuteTool(toolName, args)
}

// GetAllAgents 获取所有插件的Agent
func (m *Manager) GetAllAgents() []AgentInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allAgents []AgentInfo
	for _, p := range m.plugins {
		agents := p.GetAgents()
		allAgents = append(allAgents, agents...)
	}

	return allAgents
}