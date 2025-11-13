package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/yahao333/GoManus/internal/schema"
	"github.com/yahao333/GoManus/pkg/config"
	"go.uber.org/zap"
)

// ClientSession MCP客户端会话接口
type ClientSession interface {
	Initialize(ctx context.Context) error
	ListTools(ctx context.Context) (*ListToolsResult, error)
	CallTool(ctx context.Context, name string, args map[string]interface{}) (*CallToolResult, error)
	Close() error
}

// ListToolsResult 工具列表结果
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// Tool MCP工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// CallToolResult 工具调用结果
type CallToolResult struct {
	Content []Content `json:"content"`
}

// Content 内容项
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MCPTool MCP工具适配器
type MCPTool struct {
	name         string
	description  string
	inputSchema  map[string]interface{}
	session      ClientSession
	serverID     string
	originalName string
	logger       *zap.Logger
}

// NewMCPTool 创建新的MCP工具
func NewMCPTool(name, description string, inputSchema map[string]interface{},
	session ClientSession, serverID, originalName string, logger *zap.Logger) *MCPTool {
	return &MCPTool{
		name:         name,
		description:  description,
		inputSchema:  inputSchema,
		session:      session,
		serverID:     serverID,
		originalName: originalName,
		logger:       logger,
	}
}

// Name 获取工具名称
func (t *MCPTool) Name() string {
	return t.name
}

// Description 获取工具描述
func (t *MCPTool) Description() string {
	return t.description
}

// Parameters 获取工具参数
func (t *MCPTool) Parameters() map[string]interface{} {
	return t.inputSchema
}

// Execute 执行工具
func (t *MCPTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	t.logger.Info("执行MCP工具", zap.String("tool", t.originalName), zap.String("server", t.serverID))

	result, err := t.session.CallTool(ctx, t.originalName, args)
	if err != nil {
		t.logger.Error("MCP工具执行失败", zap.Error(err), zap.String("tool", t.originalName))
		return nil, fmt.Errorf("MCP工具执行失败: %w", err)
	}

	// 转换结果为字符串
	var output string
	for _, content := range result.Content {
		if content.Type == "text" {
			output += content.Text + "\n"
		}
	}

	if output == "" {
		output = "工具执行完成，无返回内容"
	}

	return output, nil
}

// ToToolDefinition 转换为工具定义
func (t *MCPTool) ToToolDefinition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        t.name,
		Description: t.description,
		Parameters:  t.inputSchema,
	}
}

// MCPClients MCP客户端管理器
type MCPClients struct {
	sessions map[string]ClientSession
	tools    map[string]*MCPTool
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewMCPClients 创建新的MCP客户端管理器
func NewMCPClients(logger *zap.Logger) *MCPClients {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MCPClients{
		sessions: make(map[string]ClientSession),
		tools:    make(map[string]*MCPTool),
		logger:   logger,
	}
}

// ConnectSSE 通过SSE连接MCP服务器
func (m *MCPClients) ConnectSSE(ctx context.Context, serverURL, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if serverID == "" {
		serverID = serverURL
	}

	// 如果已存在连接，先断开
	if _, exists := m.sessions[serverID]; exists {
		m.disconnectLocked(serverID)
	}

	// 创建SSE会话
	session := NewSSESession(serverURL, m.logger)
	if err := session.Initialize(ctx); err != nil {
		return fmt.Errorf("初始化SSE会话失败: %w", err)
	}

	m.sessions[serverID] = session

	// 加载工具
	if err := m.loadToolsLocked(ctx, serverID, session); err != nil {
		session.Close()
		delete(m.sessions, serverID)
		return fmt.Errorf("加载工具失败: %w", err)
	}

	m.logger.Info("连接到MCP服务器", zap.String("server_id", serverID), zap.String("url", serverURL))
	return nil
}

// ConnectStdio 通过stdio连接MCP服务器
func (m *MCPClients) ConnectStdio(ctx context.Context, command string, args []string, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if serverID == "" {
		serverID = command
	}

	// 如果已存在连接，先断开
	if _, exists := m.sessions[serverID]; exists {
		m.disconnectLocked(serverID)
	}

	// 创建stdio会话
	session := NewStdioSession(command, args, m.logger)
	if err := session.Initialize(ctx); err != nil {
		return fmt.Errorf("初始化stdio会话失败: %w", err)
	}

	m.sessions[serverID] = session

	// 加载工具
	if err := m.loadToolsLocked(ctx, serverID, session); err != nil {
		session.Close()
		delete(m.sessions, serverID)
		return fmt.Errorf("加载工具失败: %w", err)
	}

	m.logger.Info("连接到MCP服务器", zap.String("server_id", serverID), zap.String("command", command))
	return nil
}

// loadToolsLocked 加载工具（内部使用，需要持有锁）
func (m *MCPClients) loadToolsLocked(ctx context.Context, serverID string, session ClientSession) error {
	result, err := session.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("列出工具失败: %w", err)
	}

	for _, tool := range result.Tools {
		toolName := sanitizeToolName(fmt.Sprintf("mcp_%s_%s", serverID, tool.Name))

		mcpTool := NewMCPTool(
			toolName,
			tool.Description,
			tool.InputSchema,
			session,
			serverID,
			tool.Name,
			m.logger,
		)

		m.tools[toolName] = mcpTool
		m.logger.Info("加载MCP工具", zap.String("tool", toolName), zap.String("server", serverID))
	}

	return nil
}

// Disconnect 断开连接
func (m *MCPClients) Disconnect(serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.disconnectLocked(serverID)
}

// disconnectLocked 断开连接（内部使用，需要持有锁）
func (m *MCPClients) disconnectLocked(serverID string) error {
	session, exists := m.sessions[serverID]
	if !exists {
		return nil
	}

	// 移除相关工具
	for toolName, tool := range m.tools {
		if tool.serverID == serverID {
			delete(m.tools, toolName)
		}
	}

	if err := session.Close(); err != nil {
		m.logger.Error("关闭会话失败", zap.Error(err), zap.String("server_id", serverID))
	}

	delete(m.sessions, serverID)
	m.logger.Info("断开MCP服务器连接", zap.String("server_id", serverID))
	return nil
}

// GetTools 获取所有MCP工具
func (m *MCPClients) GetTools() map[string]*MCPTool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make(map[string]*MCPTool)
	for k, v := range m.tools {
		tools[k] = v
	}
	return tools
}

// GetTool 获取指定工具
func (m *MCPClients) GetTool(name string) (*MCPTool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tool, exists := m.tools[name]
	return tool, exists
}

// Close 关闭所有连接
func (m *MCPClients) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for serverID := range m.sessions {
		if err := m.disconnectLocked(serverID); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭连接时发生错误: %v", errs)
	}
	return nil
}

// InitializeFromConfig 从配置初始化MCP客户端
func (m *MCPClients) InitializeFromConfig(ctx context.Context, mcpConfig *config.MCPConfig) error {
	if !mcpConfig.Enabled {
		m.logger.Info("MCP功能未启用")
		return nil
	}

	m.logger.Info("开始初始化MCP客户端", zap.Int("server_count", len(mcpConfig.Servers)))

	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		m.logger.Info("初始化过程被取消")
		return fmt.Errorf("初始化被取消: %w", ctx.Err())
	default:
	}

	for serverID, serverConfig := range mcpConfig.Servers {
		// 定期检查上下文是否已取消
		select {
		case <-ctx.Done():
			m.logger.Info("初始化过程被取消")
			return fmt.Errorf("初始化被取消: %w", ctx.Err())
		default:
		}

		switch serverConfig.Type {
		case "sse":
			if serverConfig.URL != "" {
				if err := m.ConnectSSE(ctx, serverConfig.URL, serverID); err != nil {
					m.logger.Error("连接SSE服务器失败",
						zap.String("server_id", serverID),
						zap.String("url", serverConfig.URL),
						zap.Error(err))
					continue
				}
			}
		case "stdio":
			if serverConfig.Command != "" {
				if err := m.ConnectStdio(ctx, serverConfig.Command, serverConfig.Args, serverID); err != nil {
					m.logger.Error("连接stdio服务器失败",
						zap.String("server_id", serverID),
						zap.String("command", serverConfig.Command),
						zap.Error(err))
					continue
				}
			}
		default:
			m.logger.Error("不支持的MCP服务器类型",
				zap.String("server_id", serverID),
				zap.String("type", serverConfig.Type))
		}
	}

	m.logger.Info("MCP客户端初始化完成", zap.Int("connected_servers", len(m.sessions)))
	return nil
}

// sanitizeToolName 工具名称清理
func sanitizeToolName(name string) string {
	// 实现工具名称清理逻辑
	if len(name) > 64 {
		name = name[:64]
	}
	return name
}
