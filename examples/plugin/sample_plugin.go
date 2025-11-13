// +build plugin

package main

import (
	"fmt"
	"github.com/yahao333/GoManus/internal/plugin"
	"github.com/yahao333/GoManus/internal/schema"
)

// SamplePlugin 示例插件
type SamplePlugin struct {
	config map[string]interface{}
}

// NewPlugin 创建插件实例（插件入口点）
func NewPlugin() plugin.Plugin {
	return &SamplePlugin{}
}

// Name 返回插件名称
func (p *SamplePlugin) Name() string {
	return "sample"
}

// Version 返回插件版本
func (p *SamplePlugin) Version() string {
	return "1.0.0"
}

// Description 返回插件描述
func (p *SamplePlugin) Description() string {
	return "示例插件，提供基础的计算工具"
}

// Init 初始化插件
func (p *SamplePlugin) Init(config map[string]interface{}) error {
	p.config = config
	return nil
}

// Start 启动插件
func (p *SamplePlugin) Start() error {
	fmt.Println("Sample plugin started")
	return nil
}

// Stop 停止插件
func (p *SamplePlugin) Stop() error {
	fmt.Println("Sample plugin stopped")
	return nil
}

// GetTools 返回插件提供的工具
func (p *SamplePlugin) GetTools() []schema.ToolDefinition {
	return []schema.ToolDefinition{
		{
			Name:        "calculator_add",
			Description: "加法计算",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type":        "number",
						"description": "第一个数字",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "第二个数字",
					},
				},
				"required": []string{"a", "b"},
			},
			Required: []string{"a", "b"},
		},
		{
			Name:        "calculator_multiply",
			Description: "乘法计算",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type":        "number",
						"description": "第一个数字",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "第二个数字",
					},
				},
				"required": []string{"a", "b"},
			},
			Required: []string{"a", "b"},
		},
		{
			Name:        "string_reverse",
			Description: "字符串反转",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "要反转的字符串",
					},
				},
				"required": []string{"text"},
			},
			Required: []string{"text"},
		},
	}
}

// ExecuteTool 执行工具
func (p *SamplePlugin) ExecuteTool(name string, args map[string]interface{}) (interface{}, error) {
	switch name {
	case "calculator_add":
		return p.calculatorAdd(args)
	case "calculator_multiply":
		return p.calculatorMultiply(args)
	case "string_reverse":
		return p.stringReverse(args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// GetAgents 返回插件提供的Agent
func (p *SamplePlugin) GetAgents() []plugin.AgentInfo {
	return []plugin.AgentInfo{
		{
			Name:        "math_agent",
			Type:        "calculator",
			Description: "数学计算Agent",
			Config: map[string]interface{}{
				"precision": 2,
			},
		},
	}
}

// calculatorAdd 加法计算
func (p *SamplePlugin) calculatorAdd(args map[string]interface{}) (interface{}, error) {
	a, ok := args["a"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid argument 'a': must be number")
	}

	b, ok := args["b"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid argument 'b': must be number")
	}

	return a + b, nil
}

// calculatorMultiply 乘法计算
func (p *SamplePlugin) calculatorMultiply(args map[string]interface{}) (interface{}, error) {
	a, ok := args["a"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid argument 'a': must be number")
	}

	b, ok := args["b"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid argument 'b': must be number")
	}

	return a * b, nil
}

// stringReverse 字符串反转
func (p *SamplePlugin) stringReverse(args map[string]interface{}) (interface{}, error) {
	text, ok := args["text"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument 'text': must be string")
	}

	// 字符串反转
	runes := []rune(text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes), nil
}