package flow

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/yahao333/GoManus/internal/agent"
    "github.com/yahao333/GoManus/internal/logger"
    "github.com/yahao333/GoManus/internal/schema"
    "go.uber.org/zap"
)

// Flow 工作流接口
type Flow interface {
	Execute(ctx context.Context, input string) (string, error)
	GetStatus() FlowStatus
	GetAgents() []agent.BaseAgent
}

// FlowStatus 工作流状态
type FlowStatus string

const (
	FlowStatusIdle    FlowStatus = "IDLE"
	FlowStatusRunning FlowStatus = "RUNNING"
	FlowStatusPaused  FlowStatus = "PAUSED"
	FlowStatusFinished FlowStatus = "FINISHED"
	FlowStatusError   FlowStatus = "ERROR"
)

// BaseFlow 基础工作流
type BaseFlow struct {
	ID          string
	Name        string
	Description string
	Status      FlowStatus
	Agents      []agent.BaseAgent
	CurrentStep int
	MaxSteps    int
	
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewBaseFlow 创建基础工作流
func NewBaseFlow(name, description string) *BaseFlow {
	return &BaseFlow{
		ID:          generateFlowID(),
		Name:        name,
		Description: description,
		Status:      FlowStatusIdle,
		Agents:      make([]agent.BaseAgent, 0),
		CurrentStep: 0,
		MaxSteps:    10,
	}
}

// AddAgent 添加智能体
func (f *BaseFlow) AddAgent(ag agent.BaseAgent) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Agents = append(f.Agents, ag)
}

// RemoveAgent 移除智能体
func (f *BaseFlow) RemoveAgent(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	for i, ag := range f.Agents {
		if ag.GetName() == name {
			f.Agents = append(f.Agents[:i], f.Agents[i+1:]...)
			break
		}
	}
}

// GetStatus 获取状态
func (f *BaseFlow) GetStatus() FlowStatus {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.Status
}

// GetAgents 获取智能体列表
func (f *BaseFlow) GetAgents() []agent.BaseAgent {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.Agents
}

// SetStatus 设置状态
func (f *BaseFlow) SetStatus(status FlowStatus) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Status = status
	logger.Info("工作流状态变更", 
		zap.String("flow", f.Name),
		zap.String("status", string(status)))
}

// Initialize 初始化工作流
func (f *BaseFlow) Initialize(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.ctx != nil {
		return fmt.Errorf("工作流已经初始化")
	}

	f.ctx, f.cancel = context.WithCancel(ctx)
	f.Status = FlowStatusIdle

	// 初始化所有智能体
	for _, ag := range f.Agents {
		if err := ag.Initialize(f.ctx); err != nil {
			return fmt.Errorf("初始化智能体 %s 失败: %w", ag.GetName(), err)
		}
	}

	logger.Info("工作流初始化完成", zap.String("flow", f.Name))
	return nil
}

// Cleanup 清理工作流
func (f *BaseFlow) Cleanup() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.cancel != nil {
		f.cancel()
	}

	// 清理所有智能体
	for _, ag := range f.Agents {
		if err := ag.Cleanup(f.ctx); err != nil {
			logger.Error("清理智能体失败", 
				zap.String("agent", ag.GetName()),
				zap.Error(err))
		}
	}

	f.Status = FlowStatusIdle
	logger.Info("工作流清理完成", zap.String("flow", f.Name))
	return nil
}

// PlanningFlow 规划工作流
type PlanningFlow struct {
	*BaseFlow
	PlanningAgent agent.BaseAgent
	ExecutionAgent agent.BaseAgent
}

// NewPlanningFlow 创建规划工作流
func NewPlanningFlow() *PlanningFlow {
	baseFlow := NewBaseFlow("PlanningFlow", "规划工作流")
	
	// 创建规划智能体
	planningAgent, _ := agent.NewAgent(
		"Planner",
		"规划智能体",
		"你是一个任务规划专家，负责将复杂任务分解为可执行的步骤。",
		"确定下一步应该执行什么。",
	)
	
	// 创建执行智能体
	executionAgent, _ := agent.NewAgent(
		"Executor",
		"执行智能体",
		"你是一个任务执行专家，负责执行具体的任务步骤。",
		"完成当前任务步骤。",
	)
	
	flow := &PlanningFlow{
		BaseFlow:       baseFlow,
		PlanningAgent:  planningAgent,
		ExecutionAgent: executionAgent,
	}
	
	flow.AddAgent(planningAgent)
	flow.AddAgent(executionAgent)
	
	return flow
}

// Execute 执行工作流
func (f *PlanningFlow) Execute(ctx context.Context, input string) (string, error) {
	if err := f.Initialize(ctx); err != nil {
		return "", fmt.Errorf("初始化工作流失败: %w", err)
	}
	defer f.Cleanup()

	f.SetStatus(FlowStatusRunning)
	defer f.SetStatus(FlowStatusFinished)

	logger.Info("开始执行规划工作流", zap.String("input", input))

	// 步骤1: 规划阶段
	planMessage := schema.NewUserMessage(fmt.Sprintf("请为以下任务创建详细的执行计划: %s", input))
	planResponse, err := f.PlanningAgent.ProcessMessage(ctx, planMessage)
	if err != nil {
		f.SetStatus(FlowStatusError)
		return "", fmt.Errorf("规划阶段失败: %w", err)
	}

	plan := ""
	if planResponse.Content != nil {
		plan = *planResponse.Content
	}

	logger.Info("规划完成", zap.String("plan", plan))

	// 步骤2: 执行阶段
	executionMessage := schema.NewUserMessage(fmt.Sprintf("请按照以下计划执行任务: %s", plan))
	executionResponse, err := f.ExecutionAgent.ProcessMessage(ctx, executionMessage)
	if err != nil {
		f.SetStatus(FlowStatusError)
		return "", fmt.Errorf("执行阶段失败: %w", err)
	}

	result := ""
	if executionResponse.Content != nil {
		result = *executionResponse.Content
	}

	logger.Info("执行完成", zap.String("result", result))

	return result, nil
}

// MultiAgentFlow 多智能体工作流
type MultiAgentFlow struct {
	*BaseFlow
	Coordinator agent.BaseAgent
}

// NewMultiAgentFlow 创建多智能体工作流
func NewMultiAgentFlow() *MultiAgentFlow {
	baseFlow := NewBaseFlow("MultiAgentFlow", "多智能体工作流")
	
	// 创建协调智能体
	coordinator, _ := agent.NewAgent(
		"Coordinator",
		"协调智能体",
		"你是一个任务协调专家，负责协调多个智能体完成任务。",
		"确定哪个智能体应该执行下一步。",
	)
	
	flow := &MultiAgentFlow{
		BaseFlow:    baseFlow,
		Coordinator: coordinator,
	}
	
	flow.AddAgent(coordinator)
	
	return flow
}

// AddSpecializedAgent 添加专业智能体
func (f *MultiAgentFlow) AddSpecializedAgent(agent agent.BaseAgent) {
	f.AddAgent(agent)
}

// Execute 执行多智能体工作流
func (f *MultiAgentFlow) Execute(ctx context.Context, input string) (string, error) {
	if err := f.Initialize(ctx); err != nil {
		return "", fmt.Errorf("初始化工作流失败: %w", err)
	}
	defer f.Cleanup()

	f.SetStatus(FlowStatusRunning)
	defer f.SetStatus(FlowStatusFinished)

	logger.Info("开始执行多智能体工作流", zap.String("input", input))

	// 协调智能体分析任务
	coordinationMessage := schema.NewUserMessage(fmt.Sprintf("分析以下任务并确定最佳执行策略: %s", input))
	coordinationResponse, err := f.Coordinator.ProcessMessage(ctx, coordinationMessage)
	if err != nil {
		f.SetStatus(FlowStatusError)
		return "", fmt.Errorf("协调阶段失败: %w", err)
	}

	strategy := ""
	if coordinationResponse.Content != nil {
		strategy = *coordinationResponse.Content
	}

	logger.Info("协调策略", zap.String("strategy", strategy))

	// 根据策略分配任务给专业智能体
	var results []string
	for _, ag := range f.Agents {
		if ag.GetName() == "Coordinator" {
			continue // 跳过协调智能体
		}

		taskMessage := schema.NewUserMessage(fmt.Sprintf("根据策略 '%s' 执行任务: %s", strategy, input))
		response, err := ag.ProcessMessage(ctx, taskMessage)
		if err != nil {
			logger.Error("智能体执行任务失败", 
				zap.String("agent", ag.GetName()),
				zap.Error(err))
			continue
		}

		if response.Content != nil {
			results = append(results, *response.Content)
		}
	}

	// 协调智能体汇总结果
	finalMessage := schema.NewUserMessage(fmt.Sprintf("汇总以下结果: %v", results))
	finalResponse, err := f.Coordinator.ProcessMessage(ctx, finalMessage)
	if err != nil {
		f.SetStatus(FlowStatusError)
		return "", fmt.Errorf("汇总阶段失败: %w", err)
	}

	finalResult := ""
	if finalResponse.Content != nil {
		finalResult = *finalResponse.Content
	}

	logger.Info("多智能体工作流完成", zap.String("result", finalResult))

	return finalResult, nil
}

// generateFlowID 生成工作流ID
func generateFlowID() string {
	return fmt.Sprintf("flow_%d", time.Now().UnixNano())
}
