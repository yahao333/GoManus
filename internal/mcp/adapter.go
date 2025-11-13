package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yahao333/GoManus/internal/schema"
	"github.com/yahao333/GoManus/internal/tool"
)

// MCPToolAdapter MCP工具适配器，适配到GoManus工具系统
type MCPToolAdapter struct {
	mcpTool *MCPTool
}

// NewMCPToolAdapter 创建MCP工具适配器
func NewMCPToolAdapter(mcpTool *MCPTool) *MCPToolAdapter {
	return &MCPToolAdapter{
		mcpTool: mcpTool,
	}
}

// Name 获取工具名称
func (a *MCPToolAdapter) Name() string {
	return a.mcpTool.Name()
}

// Description 获取工具描述
func (a *MCPToolAdapter) Description() string {
	return a.mcpTool.Description()
}

// GetName 获取工具名称
func (a *MCPToolAdapter) GetName() string {
	return a.mcpTool.Name()
}

// GetDescription 获取工具描述
func (a *MCPToolAdapter) GetDescription() string {
	return a.mcpTool.Description()
}

// GetParameters 获取工具参数
func (a *MCPToolAdapter) GetParameters() map[string]interface{} {
	return a.mcpTool.Parameters()
}

// GetRequired 获取必需参数
func (a *MCPToolAdapter) GetRequired() []string {
	// MCP工具通过JSON Schema定义必需参数
	if schema, ok := a.mcpTool.inputSchema["required"].([]interface{}); ok {
		required := make([]string, len(schema))
		for i, item := range schema {
			if str, ok := item.(string); ok {
				required[i] = str
			}
		}
		return required
	}
	return []string{}
}

// Execute 执行工具
func (a *MCPToolAdapter) Execute(ctx context.Context, arguments string) (interface{}, error) {
	// 解析JSON参数
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return nil, fmt.Errorf("解析工具参数失败: %w", err)
	}
	
	return a.mcpTool.Execute(ctx, args)
}

// ToToolDefinition 转换为工具定义
func (a *MCPToolAdapter) ToToolDefinition() schema.ToolDefinition {
	return a.mcpTool.ToToolDefinition()
}

// MCPToolCollection MCP工具集合
type MCPToolCollection struct {
	tools map[string]*MCPToolAdapter
}

// NewMCPToolCollection 创建MCP工具集合
func NewMCPToolCollection() *MCPToolCollection {
	return &MCPToolCollection{
		tools: make(map[string]*MCPToolAdapter),
	}
}

// AddTool 添加MCP工具
func (c *MCPToolCollection) AddTool(mcpTool *MCPTool) {
	adapter := NewMCPToolAdapter(mcpTool)
	c.tools[adapter.Name()] = adapter
}

// GetTool 获取工具
func (c *MCPToolCollection) GetTool(name string) (tool.Tool, bool) {
	adapter, exists := c.tools[name]
	if !exists {
		return nil, false
	}
	return adapter, true
}

// GetTools 获取所有工具
func (c *MCPToolCollection) GetTools() map[string]tool.Tool {
	tools := make(map[string]tool.Tool)
	for name, adapter := range c.tools {
		tools[name] = adapter
	}
	return tools
}

// RemoveTool 移除工具
func (c *MCPToolCollection) RemoveTool(name string) {
	delete(c.tools, name)
}

// Clear 清空所有工具
func (c *MCPToolCollection) Clear() {
	c.tools = make(map[string]*MCPToolAdapter)
}

// Size 获取工具数量
func (c *MCPToolCollection) Size() int {
	return len(c.tools)
}

// ConvertToToolDefinitions 转换为工具定义列表
func (c *MCPToolCollection) ConvertToToolDefinitions() []schema.ToolDefinition {
	definitions := make([]schema.ToolDefinition, 0, len(c.tools))
	for _, adapter := range c.tools {
		definitions = append(definitions, adapter.ToToolDefinition())
	}
	return definitions
}

// ConvertToToolCollection 转换为标准工具集合
func (c *MCPToolCollection) ConvertToToolCollection() *tool.ToolCollection {
	collection := tool.NewToolCollection()
	for _, adapter := range c.tools {
		collection.AddTool(adapter)
	}
	return collection
}

// MergeWithToolCollection 与现有工具集合合并
func (c *MCPToolCollection) MergeWithToolCollection(existingCollection *tool.ToolCollection) *tool.ToolCollection {
	// 创建新的工具集合
	merged := tool.NewToolCollection()
	
	// 添加现有工具
	existingTools := existingCollection.GetAllTools()
	for _, t := range existingTools {
		merged.AddTool(t)
	}
	
	// 添加MCP工具
	for name, adapter := range c.tools {
		if _, err := merged.GetTool(name); err != nil {
			merged.AddTool(adapter)
		}
	}
	
	return merged
}