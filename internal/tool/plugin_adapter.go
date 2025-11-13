package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yahao333/GoManus/internal/plugin"
	"github.com/yahao333/GoManus/internal/schema"
)

// PluginToolAdapter 插件工具适配器
type PluginToolAdapter struct {
	BaseTool
	pluginManager *plugin.Manager
	pluginName    string
	toolName      string
}

// NewPluginToolAdapter 创建插件工具适配器
func NewPluginToolAdapter(pluginManager *plugin.Manager, pluginName, toolName string, definition schema.ToolDefinition) *PluginToolAdapter {
	return &PluginToolAdapter{
		BaseTool: BaseTool{
			Name:        definition.Name,
			Description: definition.Description,
			Parameters:  definition.Parameters,
			Required:    definition.Required,
		},
		pluginManager: pluginManager,
		pluginName:    pluginName,
		toolName:      toolName,
	}
}

// Execute 执行插件工具
func (p *PluginToolAdapter) Execute(ctx context.Context, arguments string) (interface{}, error) {
	// 解析参数
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return nil, fmt.Errorf("解析参数失败: %w", err)
	}

	// 执行插件工具
	result, err := p.pluginManager.ExecuteTool(p.pluginName, p.toolName, args)
	if err != nil {
		return nil, fmt.Errorf("执行插件工具失败: %w", err)
	}

	return result, nil
}

// PluginToolManager 插件工具管理器
type PluginToolManager struct {
	pluginManager *plugin.Manager
	toolCollection *ToolCollection
}

// NewPluginToolManager 创建插件工具管理器
func NewPluginToolManager(pluginManager *plugin.Manager, toolCollection *ToolCollection) *PluginToolManager {
	return &PluginToolManager{
		pluginManager:  pluginManager,
		toolCollection: toolCollection,
	}
}

// LoadPluginTools 加载插件工具
func (p *PluginToolManager) LoadPluginTools() error {
	// 获取所有插件
	plugins := p.pluginManager.ListPlugins()

	for _, pluginMetadata := range plugins {
		plugin, err := p.pluginManager.GetPlugin(pluginMetadata.Name)
		if err != nil {
			return fmt.Errorf("获取插件失败: %w", err)
		}

		// 获取插件的所有工具
		tools := plugin.GetTools()
		for _, tool := range tools {
			// 创建工具适配器
			adapter := NewPluginToolAdapter(p.pluginManager, pluginMetadata.Name, tool.Name, tool)

			// 添加到工具集合
			p.toolCollection.AddTool(adapter)
		}
	}

	return nil
}

// getRequiredParameters 从参数定义中获取必需参数
func getRequiredParameters(parameters map[string]interface{}) []string {
	if parameters == nil {
		return []string{}
	}

	// 检查是否有required字段
	if required, ok := parameters["required"].([]interface{}); ok {
		result := make([]string, len(required))
		for i, param := range required {
			if paramStr, ok := param.(string); ok {
				result[i] = paramStr
			}
		}
		return result
	}

	return []string{}
}

// UnloadPluginTools 卸载插件工具
func (p *PluginToolManager) UnloadPluginTools(pluginName string) error {
	// 获取插件的所有工具
	plugin, err := p.pluginManager.GetPlugin(pluginName)
	if err != nil {
		return fmt.Errorf("获取插件失败: %w", err)
	}

	tools := plugin.GetTools()
	for _, tool := range tools {
		// 从工具集合中移除
		toolName := fmt.Sprintf("%s_%s", pluginName, tool.Name)
		p.toolCollection.RemoveTool(toolName)
	}

	return nil
}