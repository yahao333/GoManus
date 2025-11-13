package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/yahao333/GoManus/internal/llm"
	"github.com/yahao333/GoManus/internal/logger"
	"github.com/yahao333/GoManus/internal/schema"
	"github.com/yahao333/GoManus/internal/tool"
	"go.uber.org/zap"
)

// BaseAgent 基础智能体接口
type BaseAgent interface {
	GetName() string
	GetDescription() string
	GetState() schema.AgentState
	SetState(state schema.AgentState)
	GetMemory() *schema.Memory
	GetLLM() *llm.LLM
	GetAvailableTools() *tool.ToolCollection
	
	// 核心方法
	Initialize(ctx context.Context) error
	ProcessMessage(ctx context.Context, message schema.Message) (*schema.Message, error)
	Run(ctx context.Context, prompt string) error
	Cleanup(ctx context.Context) error
	
	// 状态管理
	UpdateMemory(role schema.Role, content string, base64Image ...string) error
	GetSystemPrompt() string
	GetNextStepPrompt() string
}

// Agent 基础智能体实现
type Agent struct {
	ID               string
	Name             string
	Description      string
	SystemPrompt     string
	NextStepPrompt   string
	State            schema.AgentState
	Memory           *schema.Memory
	LLM              *llm.LLM
	AvailableTools   *tool.ToolCollection
	MaxSteps         int
	CurrentStep      int
	DuplicateThreshold int
	
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewAgent 创建新的基础智能体
func NewAgent(name, description, systemPrompt, nextStepPrompt string) (*Agent, error) {
	// 创建LLM客户端
	llmClient, err := llm.NewLLM(name)
	if err != nil {
		return nil, fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	return &Agent{
		ID:               uuid.New().String(),
		Name:             name,
		Description:      description,
		SystemPrompt:     systemPrompt,
		NextStepPrompt:   nextStepPrompt,
		State:            schema.AgentStateIdle,
		Memory:           schema.NewMemory(100),
		LLM:              llmClient,
		AvailableTools:   tool.NewToolCollection(),
		MaxSteps:         10,
		CurrentStep:      0,
		DuplicateThreshold: 2,
	}, nil
}

// GetName 获取智能体名称
func (a *Agent) GetName() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Name
}

// GetDescription 获取智能体描述
func (a *Agent) GetDescription() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Description
}

// GetState 获取智能体状态
func (a *Agent) GetState() schema.AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.State
}

// SetState 设置智能体状态
func (a *Agent) SetState(state schema.AgentState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.State = state
	logger.Info("智能体状态变更", 
		zap.String("agent", a.Name),
		zap.String("state", string(state)))
}

// GetMemory 获取内存
func (a *Agent) GetMemory() *schema.Memory {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Memory
}

// GetLLM 获取LLM客户端
func (a *Agent) GetLLM() *llm.LLM {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.LLM
}

// GetAvailableTools 获取可用工具
func (a *Agent) GetAvailableTools() *tool.ToolCollection {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.AvailableTools
}

// Initialize 初始化智能体
func (a *Agent) Initialize(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ctx != nil {
		return fmt.Errorf("智能体已经初始化")
	}

	a.ctx, a.cancel = context.WithCancel(ctx)
	a.State = schema.AgentStateIdle

	// 添加系统消息
	if a.SystemPrompt != "" {
		a.Memory.AddMessage(schema.NewSystemMessage(a.SystemPrompt))
	}

	logger.Info("智能体初始化完成", zap.String("agent", a.Name))
	return nil
}

// ProcessMessage 处理消息
func (a *Agent) ProcessMessage(ctx context.Context, message schema.Message) (*schema.Message, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.State != schema.AgentStateIdle && a.State != schema.AgentStateRunning {
		return nil, fmt.Errorf("智能体状态不正确: %s", a.State)
	}

	// 添加消息到内存
	a.Memory.AddMessage(message)

	// 生成响应
	response, err := a.generateResponse(ctx)
	if err != nil {
		a.State = schema.AgentStateError
		return nil, fmt.Errorf("生成响应失败: %w", err)
	}

	// 添加响应到内存
	a.Memory.AddMessage(*response)

	return response, nil
}

// Run 运行智能体
func (a *Agent) Run(ctx context.Context, prompt string) error {
	if err := a.Initialize(ctx); err != nil {
		return fmt.Errorf("初始化智能体失败: %w", err)
	}
	defer a.Cleanup(ctx)

	// 设置运行状态
	a.SetState(schema.AgentStateRunning)
	defer a.SetState(schema.AgentStateFinished)

	// 添加用户消息
	userMessage := schema.NewUserMessage(prompt)
	a.Memory.AddMessage(userMessage)

	logger.Info("开始运行智能体", 
		zap.String("agent", a.Name),
		zap.String("prompt", prompt))

	// 执行步骤循环
	for a.CurrentStep < a.MaxSteps {
		select {
		case <-a.ctx.Done():
			return fmt.Errorf("智能体运行被取消")
		case <-ctx.Done():
			return fmt.Errorf("上下文被取消")
		default:
		}

		a.CurrentStep++
		logger.Info("执行步骤", 
			zap.String("agent", a.Name),
			zap.Int("step", a.CurrentStep),
			zap.Int("max_steps", a.MaxSteps))

		// 生成响应
		response, err := a.generateResponse(ctx)
		if err != nil {
			a.SetState(schema.AgentStateError)
			return fmt.Errorf("生成响应失败: %w", err)
		}

		// 添加响应到内存
		a.Memory.AddMessage(*response)

		// 检查是否完成任务
		if a.isTaskComplete(response) {
			logger.Info("任务完成", zap.String("agent", a.Name))
			break
		}

		// 检查重复响应
		if a.isDuplicateResponse(response) {
			logger.Warn("检测到重复响应", zap.String("agent", a.Name))
			break
		}
	}

	if a.CurrentStep >= a.MaxSteps {
		logger.Warn("达到最大步骤限制", 
			zap.String("agent", a.Name),
			zap.Int("max_steps", a.MaxSteps))
	}

	return nil
}

// Cleanup 清理资源
func (a *Agent) Cleanup(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}

	a.State = schema.AgentStateIdle
	logger.Info("智能体清理完成", zap.String("agent", a.Name))
	return nil
}

// UpdateMemory 更新内存
func (a *Agent) UpdateMemory(role schema.Role, content string, base64Image ...string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	message := schema.NewUserMessage(content)
	message.Role = role
	if len(base64Image) > 0 {
		message.Base64Image = &base64Image[0]
	}

	a.Memory.AddMessage(message)
	return nil
}

// GetSystemPrompt 获取系统提示
func (a *Agent) GetSystemPrompt() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.SystemPrompt
}

// GetNextStepPrompt 获取下一步提示
func (a *Agent) GetNextStepPrompt() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.NextStepPrompt
}

// generateResponse 生成响应
func (a *Agent) generateResponse(ctx context.Context) (*schema.Message, error) {
	// 获取工具定义
	toolDefs := a.AvailableTools.GetDefinitions()

	// 生成响应
	response, err := a.LLM.GenerateResponse(ctx, a.Memory.GetRecentMessages(20), toolDefs)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// isTaskComplete 检查任务是否完成
func (a *Agent) isTaskComplete(response *schema.Message) bool {
	if response.Content != nil {
		content := *response.Content
		// 简单的完成检测逻辑
		if contains(content, "任务完成") || contains(content, "task completed") ||
		   contains(content, "完成") || contains(content, "completed") {
			return true
		}
	}
	return false
}

// isDuplicateResponse 检查重复响应
func (a *Agent) isDuplicateResponse(response *schema.Message) bool {
	if response.Content == nil {
		return false
	}

	currentContent := *response.Content
	recentMessages := a.Memory.GetRecentMessages(5)

	duplicateCount := 0
	for _, msg := range recentMessages {
		if msg.Content != nil && *msg.Content == currentContent {
			duplicateCount++
		}
	}

	return duplicateCount >= a.DuplicateThreshold
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

// containsSubstring 检查字符串是否包含子字符串（大小写不敏感）
func containsSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	
	// 简单的子字符串搜索
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}