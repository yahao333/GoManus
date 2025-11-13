package agent

import (
    "context"
    "fmt"

    "github.com/yahao333/GoManus/internal/logger"
    "github.com/yahao333/GoManus/internal/schema"
    "go.uber.org/zap"
)

// ToolCallAgent 工具调用智能体
type ToolCallAgent struct {
	*Agent
	MaxObserve    int
	SpecialTools  []string
}

// NewToolCallAgent 创建新的工具调用智能体
func NewToolCallAgent(name, description, systemPrompt, nextStepPrompt string) (*ToolCallAgent, error) {
	baseAgent, err := NewAgent(name, description, systemPrompt, nextStepPrompt)
	if err != nil {
		return nil, err
	}

	return &ToolCallAgent{
		Agent:        baseAgent,
		MaxObserve:   10000,
		SpecialTools: []string{},
	}, nil
}

// ProcessMessage 处理消息（重写以支持工具调用）
func (t *ToolCallAgent) ProcessMessage(ctx context.Context, message schema.Message) (*schema.Message, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State != schema.AgentStateIdle && t.State != schema.AgentStateRunning {
		return nil, fmt.Errorf("智能体状态不正确: %s", t.State)
	}

	// 添加消息到内存
	t.Memory.AddMessage(message)

	// 生成响应
	response, err := t.generateResponseWithTools(ctx)
	if err != nil {
		t.State = schema.AgentStateError
		return nil, fmt.Errorf("生成响应失败: %w", err)
	}

	// 添加响应到内存
	t.Memory.AddMessage(*response)

	// 如果有工具调用，执行工具
	if response.ToolCalls != nil && len(response.ToolCalls) > 0 {
		for _, toolCall := range response.ToolCalls {
			toolResult, err := t.executeTool(ctx, toolCall)
			if err != nil {
				logger.Error("工具执行失败", 
					zap.String("tool", toolCall.Function.Name),
					zap.Error(err))
				continue
			}

			// 添加工具结果到内存
			toolMessage := schema.NewToolMessage(
				fmt.Sprintf("%v", toolResult.Result),
				toolCall.Function.Name,
				toolCall.ID,
			)
			t.Memory.AddMessage(toolMessage)
		}
	}

	return response, nil
}

// generateResponseWithTools 生成带工具的响应
func (t *ToolCallAgent) generateResponseWithTools(ctx context.Context) (*schema.Message, error) {
	// 获取工具定义
	toolDefs := t.AvailableTools.GetDefinitions()

	// 生成响应
	response, err := t.LLM.GenerateResponse(ctx, t.Memory.GetRecentMessages(20), toolDefs)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// executeTool 执行工具
func (t *ToolCallAgent) executeTool(ctx context.Context, toolCall schema.ToolCall) (*schema.ToolResult, error) {
	toolName := toolCall.Function.Name
	toolArgs := toolCall.Function.Arguments

	logger.Info("执行工具", 
		zap.String("tool", toolName),
		zap.String("args", toolArgs))

	// 获取工具实例
	toolInstance, err := t.AvailableTools.GetTool(toolName)
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("工具未找到: %s", toolName),
		}, nil
	}

	// 执行工具
	result, err := toolInstance.Execute(ctx, toolArgs)
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 截断结果
	if len(fmt.Sprintf("%v", result)) > t.MaxObserve {
		truncated := fmt.Sprintf("%v", result)[:t.MaxObserve] + "..."
		result = truncated
	}

	return &schema.ToolResult{
		Success: true,
		Result:  result,
	}, nil
}

// isSpecialTool 检查是否为特殊工具
func (t *ToolCallAgent) isSpecialTool(toolName string) bool {
	for _, special := range t.SpecialTools {
		if special == toolName {
			return true
		}
	}
	return false
}
