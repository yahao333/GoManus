package schema

import (
	"encoding/json"
	"time"
)

// Role 消息角色类型
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ToolChoice 工具选择类型
type ToolChoice string

const (
	ToolChoiceNone     ToolChoice = "none"
	ToolChoiceAuto     ToolChoice = "auto"
	ToolChoiceRequired ToolChoice = "required"
)

// AgentState 智能体状态
type AgentState string

const (
	AgentStateIdle    AgentState = "IDLE"
	AgentStateRunning AgentState = "RUNNING"
	AgentStateFinished AgentState = "FINISHED"
	AgentStateError   AgentState = "ERROR"
)

// Function 函数定义
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Message 消息结构
type Message struct {
	Role        Role      `json:"role"`
	Content     *string   `json:"content,omitempty"`
	ToolCalls   []ToolCall `json:"tool_calls,omitempty"`
	Name        *string   `json:"name,omitempty"`
	ToolCallID  *string   `json:"tool_call_id,omitempty"`
	Base64Image *string   `json:"base64_image,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewUserMessage 创建用户消息
func NewUserMessage(content string, base64Image ...string) Message {
	msg := Message{
		Role:      RoleUser,
		Content:   &content,
		Timestamp: time.Now(),
	}
	if len(base64Image) > 0 {
		msg.Base64Image = &base64Image[0]
	}
	return msg
}

// NewSystemMessage 创建系统消息
func NewSystemMessage(content string) Message {
	return Message{
		Role:      RoleSystem,
		Content:   &content,
		Timestamp: time.Now(),
	}
}

// NewAssistantMessage 创建助手消息
func NewAssistantMessage(content string, base64Image ...string) Message {
	msg := Message{
		Role:      RoleAssistant,
		Content:   &content,
		Timestamp: time.Now(),
	}
	if len(base64Image) > 0 {
		msg.Base64Image = &base64Image[0]
	}
	return msg
}

// NewToolMessage 创建工具消息
func NewToolMessage(content, name, toolCallID string, base64Image ...string) Message {
	msg := Message{
		Role:       RoleTool,
		Content:    &content,
		Name:       &name,
		ToolCallID: &toolCallID,
		Timestamp:  time.Now(),
	}
	if len(base64Image) > 0 {
		msg.Base64Image = &base64Image[0]
	}
	return msg
}

// ToDict 将消息转换为字典
func (m Message) ToDict() map[string]interface{} {
	result := make(map[string]interface{})
	result["role"] = m.Role
	if m.Content != nil {
		result["content"] = *m.Content
	}
	if m.ToolCalls != nil {
		toolCalls := make([]map[string]interface{}, len(m.ToolCalls))
		for i, tc := range m.ToolCalls {
			toolCalls[i] = map[string]interface{}{
				"id":   tc.ID,
				"type": tc.Type,
				"function": map[string]interface{}{
					"name":      tc.Function.Name,
					"arguments": tc.Function.Arguments,
				},
			}
		}
		result["tool_calls"] = toolCalls
	}
	if m.Name != nil {
		result["name"] = *m.Name
	}
	if m.ToolCallID != nil {
		result["tool_call_id"] = *m.ToolCallID
	}
	if m.Base64Image != nil {
		result["base64_image"] = *m.Base64Image
	}
	return result
}

// Memory 内存结构
type Memory struct {
	Messages     []Message `json:"messages"`
	MaxMessages  int       `json:"max_messages"`
}

// NewMemory 创建新内存
func NewMemory(maxMessages int) *Memory {
	if maxMessages <= 0 {
		maxMessages = 100
	}
	return &Memory{
		Messages:    make([]Message, 0),
		MaxMessages: maxMessages,
	}
}

// AddMessage 添加消息到内存
func (m *Memory) AddMessage(message Message) {
	m.Messages = append(m.Messages, message)
	if len(m.Messages) > m.MaxMessages {
		m.Messages = m.Messages[len(m.Messages)-m.MaxMessages:]
	}
}

// AddMessages 添加多条消息到内存
func (m *Memory) AddMessages(messages []Message) {
	m.Messages = append(m.Messages, messages...)
	if len(m.Messages) > m.MaxMessages {
		m.Messages = m.Messages[len(m.Messages)-m.MaxMessages:]
	}
}

// Clear 清空内存
func (m *Memory) Clear() {
	m.Messages = make([]Message, 0)
}

// GetRecentMessages 获取最近的消息
func (m *Memory) GetRecentMessages(n int) []Message {
	if n <= 0 || n > len(m.Messages) {
		n = len(m.Messages)
	}
	return m.Messages[len(m.Messages)-n:]
}

// ToDictList 将消息列表转换为字典列表
func (m *Memory) ToDictList() []map[string]interface{} {
	result := make([]map[string]interface{}, len(m.Messages))
	for i, msg := range m.Messages {
		result[i] = msg.ToDict()
	}
	return result
}

// ToolResult 工具执行结果
type ToolResult struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// AgentMetadata 智能体元数据
type AgentMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	Tags        []string          `json:"tags"`
	Capabilities []string         `json:"capabilities"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// LLMConfig LLM配置
type LLMConfig struct {
	Model       string  `json:"model"`
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	APIType     string  `json:"api_type"`
	APIVersion  string  `json:"api_version"`
	MaxInputTokens *int `json:"max_input_tokens,omitempty"`
}

// ToolDefinition 工具定义
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Required    []string               `json:"required"`
}

// String 实现Stringer接口
func (r Role) String() string {
	return string(r)
}

// String 实现Stringer接口
func (tc ToolChoice) String() string {
	return string(tc)
}

// String 实现Stringer接口
func (as AgentState) String() string {
	return string(as)
}

// MarshalJSON 自定义JSON序列化
func (m Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(m),
	})
}

// UnmarshalJSON 自定义JSON反序列化
func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}
	return nil
}