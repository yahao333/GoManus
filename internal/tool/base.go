package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yahao333/GoManus/internal/schema"
)

// Tool 工具接口
type Tool interface {
	GetName() string
	GetDescription() string
	GetParameters() map[string]interface{}
	GetRequired() []string
	Execute(ctx context.Context, arguments string) (interface{}, error)
}

// BaseTool 基础工具实现
type BaseTool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Required    []string
}

// GetName 获取工具名称
func (b *BaseTool) GetName() string {
	return b.Name
}

// GetDescription 获取工具描述
func (b *BaseTool) GetDescription() string {
	return b.Description
}

// GetParameters 获取工具参数
func (b *BaseTool) GetParameters() map[string]interface{} {
	return b.Parameters
}

// GetRequired 获取必需参数
func (b *BaseTool) GetRequired() []string {
	return b.Required
}

// ToolCollection 工具集合
type ToolCollection struct {
	tools map[string]Tool
}

// NewToolCollection 创建新的工具集合
func NewToolCollection() *ToolCollection {
	return &ToolCollection{
		tools: make(map[string]Tool),
	}
}

// AddTool 添加工具
func (tc *ToolCollection) AddTool(tool Tool) {
	tc.tools[tool.GetName()] = tool
}

// GetTool 获取工具
func (tc *ToolCollection) GetTool(name string) (Tool, error) {
	tool, ok := tc.tools[name]
	if !ok {
		return nil, fmt.Errorf("工具未找到: %s", name)
	}
	return tool, nil
}

// RemoveTool 移除工具
func (tc *ToolCollection) RemoveTool(name string) {
	delete(tc.tools, name)
}

// GetAllTools 获取所有工具
func (tc *ToolCollection) GetAllTools() []Tool {
	tools := make([]Tool, 0, len(tc.tools))
	for _, tool := range tc.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetDefinitions 获取工具定义
func (tc *ToolCollection) GetDefinitions() []schema.ToolDefinition {
	tools := tc.GetAllTools()
	definitions := make([]schema.ToolDefinition, len(tools))
	
	for i, tool := range tools {
		definitions[i] = schema.ToolDefinition{
			Name:        tool.GetName(),
			Description: tool.GetDescription(),
			Parameters:  tool.GetParameters(),
			Required:    tool.GetRequired(),
		}
	}
	
	return definitions
}

// parseArguments 解析参数
func parseArguments(arguments string) (map[string]interface{}, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return nil, fmt.Errorf("解析参数失败: %w", err)
	}
	return args, nil
}

// validateArguments 验证参数
func validateArguments(args map[string]interface{}, required []string) error {
	for _, req := range required {
		if _, ok := args[req]; !ok {
			return fmt.Errorf("缺少必需参数: %s", req)
		}
	}
	return nil
}