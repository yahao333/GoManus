package agent

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/yahao333/GoManus/pkg/config"
    "github.com/yahao333/GoManus/internal/logger"
    "github.com/yahao333/GoManus/internal/schema"
    "github.com/yahao333/GoManus/internal/tool"
    "github.com/yahao333/GoManus/internal/mcp"
    "go.uber.org/zap"
)

// Manus 主要智能体
type Manus struct {
	*ToolCallAgent
	MaxObserve    int
	SpecialTools  []string
	MCPClients    *mcp.MCPClients
}

// NewManus 创建新的Manus智能体
func NewManus() (*Manus, error) {
	systemPrompt := fmt.Sprintf(`你是一个有用的AI助手，可以帮助用户完成各种任务。
工作目录: %s

你可以使用以下工具来完成任务：
- PythonExecute: 执行Python代码
- SimpleBrowser: 简单的HTTP浏览器
- SimpleSearch: 简单的网络搜索
- StrReplaceEditor: 编辑文件
- AskHuman: 向用户提问
- Terminate: 完成任务
- MCP工具: 通过MCP协议连接的远程工具

请根据用户的需求选择合适的工具。`, config.GetConfig().GetWorkspaceRoot())

	nextStepPrompt := "根据当前状态，确定下一步应该执行什么操作。"

	toolCallAgent, err := NewToolCallAgent(
		"Manus",
		"一个多功能的AI助手，可以使用各种工具完成任务",
		systemPrompt,
		nextStepPrompt,
	)
	if err != nil {
		return nil, err
	}

	return &Manus{
		ToolCallAgent: toolCallAgent,
		MaxObserve:    10000,
		SpecialTools:  []string{"Terminate"},
		MCPClients:    mcp.NewMCPClients(logger.GetLogger()),
	}, nil
}

// Initialize 初始化Manus智能体
func (m *Manus) Initialize(ctx context.Context) error {
	if err := m.ToolCallAgent.Initialize(ctx); err != nil {
		return err
	}

	// 添加默认工具
	m.addDefaultTools()

	// 初始化MCP客户端
	if err := m.initializeMCPClients(ctx); err != nil {
		logger.Warn("MCP客户端初始化失败", zap.Error(err))
		// 不返回错误，继续运行，MCP是可选功能
	}

	logger.Info("Manus智能体初始化完成")
	return nil
}

// Cleanup 清理资源
func (m *Manus) Cleanup(ctx context.Context) error {
	// 关闭MCP连接
	if m.MCPClients != nil {
		if err := m.MCPClients.Close(); err != nil {
			logger.Error("关闭MCP客户端失败", zap.Error(err))
		}
	}
	
	// 调用父类的清理方法
	return m.ToolCallAgent.Cleanup(ctx)
}

// initializeMCPClients 初始化MCP客户端
func (m *Manus) initializeMCPClients(ctx context.Context) error {
	mcpConfig := config.GetConfig().MCP
	if !mcpConfig.Enabled {
		logger.Info("MCP功能未启用")
		return nil
	}

	logger.Info("开始初始化MCP客户端")
	
	// 从配置初始化MCP客户端
	if err := m.MCPClients.InitializeFromConfig(ctx, &mcpConfig); err != nil {
		return fmt.Errorf("从配置初始化MCP客户端失败: %w", err)
	}

	// 获取所有MCP工具并添加到可用工具中
	mcpTools := m.MCPClients.GetTools()
	if len(mcpTools) > 0 {
		logger.Info("添加MCP工具到可用工具集", zap.Int("tool_count", len(mcpTools)))
		
		for _, mcpTool := range mcpTools {
			// 创建工具适配器
			adapter := mcp.NewMCPToolAdapter(mcpTool)
			m.AvailableTools.AddTool(adapter)
		}
	}

	logger.Info("MCP客户端初始化完成", zap.Int("tool_count", len(mcpTools)))
	return nil
}

// addDefaultTools 添加默认工具
func (m *Manus) addDefaultTools() {
	// 添加Python执行工具
	pythonTool := tool.NewPythonExecute()
	m.AvailableTools.AddTool(pythonTool)

	// 添加简化浏览器工具
	browserTool := tool.NewSimpleBrowser()
	m.AvailableTools.AddTool(browserTool)

	// 添加简化搜索工具
	searchTool := tool.NewSimpleSearch()
	m.AvailableTools.AddTool(searchTool)

	// 添加文件编辑工具
	fileTool := tool.NewStrReplaceEditor()
	m.AvailableTools.AddTool(fileTool)

	// 添加人类提问工具
	humanTool := tool.NewAskHuman()
	m.AvailableTools.AddTool(humanTool)

	// 添加终止工具
	terminateTool := tool.NewTerminate()
	m.AvailableTools.AddTool(terminateTool)
}

// Run 运行Manus智能体
func (m *Manus) Run(ctx context.Context, prompt string) error {
	logger.Info("开始运行Manus智能体", zap.String("prompt", prompt))
	
	// 创建带超时的初始化上下文（延长MCP初始化超时时间）
	initCtx, initCancel := context.WithTimeout(ctx, 60*time.Second)
	defer initCancel()
	
	// 初始化
	if err := m.Initialize(initCtx); err != nil {
		// 如果MCP初始化失败，记录警告但继续运行
		if strings.Contains(err.Error(), "MCP") {
			logger.Warn("MCP初始化失败，继续运行但MCP工具不可用", zap.Error(err))
		} else {
			return fmt.Errorf("初始化失败: %w", err)
		}
	}
	defer m.Cleanup(ctx)

	// 设置运行状态
	m.SetState(schema.AgentStateRunning)
	defer m.SetState(schema.AgentStateFinished)

	// 添加用户消息
	userMessage := schema.NewUserMessage(prompt)
	m.Memory.AddMessage(userMessage)

	// 执行主循环
	for m.CurrentStep < m.MaxSteps {
		select {
		case <-m.ctx.Done():
			return fmt.Errorf("智能体运行被取消")
		case <-ctx.Done():
			return fmt.Errorf("上下文被取消")
		default:
		}

		m.CurrentStep++
		logger.Info("执行步骤", 
			zap.Int("step", m.CurrentStep),
			zap.Int("max_steps", m.MaxSteps))

		// 处理当前状态
		response, err := m.processCurrentState(ctx)
		if err != nil {
			m.SetState(schema.AgentStateError)
			return fmt.Errorf("处理状态失败: %w", err)
		}

		// 检查是否完成任务
		if m.isTaskComplete(response) {
			logger.Info("任务完成")
			break
		}
	}

	if m.CurrentStep >= m.MaxSteps {
		logger.Warn("达到最大步骤限制", zap.Int("max_steps", m.MaxSteps))
	}

	return nil
}

// processCurrentState 处理当前状态
func (m *Manus) processCurrentState(ctx context.Context) (*schema.Message, error) {
	// 生成响应
	response, err := m.generateResponseWithTools(ctx)
	if err != nil {
		return nil, err
	}

	// 添加响应到内存
	m.Memory.AddMessage(*response)

	// 如果有工具调用，执行工具
	if response.ToolCalls != nil && len(response.ToolCalls) > 0 {
		for _, toolCall := range response.ToolCalls {
			toolResult, err := m.executeTool(ctx, toolCall)
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
			m.Memory.AddMessage(toolMessage)
		}
	}

	return response, nil
}

// isTaskComplete 检查任务是否完成
func (m *Manus) isTaskComplete(response *schema.Message) bool {
	if response.Content != nil {
		content := *response.Content
		// 检查是否包含完成标记
		if contains(content, "任务完成") || contains(content, "task completed") ||
		   contains(content, "完成") || contains(content, "completed") ||
		   contains(content, "Terminate") {
			return true
		}
	}

	// 检查工具调用
	if response.ToolCalls != nil {
		for _, tc := range response.ToolCalls {
			if tc.Function.Name == "Terminate" {
				return true
			}
		}
	}

	return false
}
