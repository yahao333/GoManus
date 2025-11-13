package memory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yahao333/GoManus/internal/schema"
)

// MemoryStore 内存存储接口
type MemoryStore interface {
	// 消息管理
	AddMessage(message schema.Message) error
	GetMessages(conversationID string, limit int) ([]schema.Message, error)
	DeleteConversation(conversationID string) error
	
	// 任务管理
	CreateTask(task Task) error
	UpdateTask(taskID string, updates map[string]interface{}) error
	GetTask(taskID string) (*Task, error)
	GetTasks(filter TaskFilter) ([]Task, error)
	
	// 会话管理
	CreateConversation(conv Conversation) error
	GetConversation(convID string) (*Conversation, error)
	GetConversations(filter ConversationFilter) ([]Conversation, error)
	
	// 工具调用记录
	RecordToolCall(call ToolCallRecord) error
	GetToolCalls(taskID string) ([]ToolCallRecord, error)
	
	// 生命周期管理
	Initialize() error
	Close() error
}

// Task 任务记录
type Task struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	Prompt         string    `json:"prompt"`
	Result         string    `json:"result,omitempty"`
	Error          string    `json:"error,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// TaskFilter 任务过滤条件
type TaskFilter struct {
	ConversationID string
	Type           string
	Status         string
	Limit          int
	Offset         int
}

// Conversation 会话记录
type Conversation struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ConversationFilter 会话过滤条件
type ConversationFilter struct {
	Status string
	Limit  int
	Offset int
}

// ToolCallRecord 工具调用记录
type ToolCallRecord struct {
	ID         string                 `json:"id"`
	TaskID     string                 `json:"task_id"`
	ToolName   string                 `json:"tool_name"`
	Arguments  map[string]interface{} `json:"arguments"`
	Result     interface{}            `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Duration   int64                  `json:"duration_ms"`
	CreatedAt  time.Time              `json:"created_at"`
}

// SQLiteMemoryStore SQLite内存存储实现
type SQLiteMemoryStore struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

// NewSQLiteMemoryStore 创建新的SQLite内存存储
func NewSQLiteMemoryStore(dbPath string) (*SQLiteMemoryStore, error) {
	store := &SQLiteMemoryStore{
		path: dbPath,
	}
	
	if err := store.Initialize(); err != nil {
		return nil, fmt.Errorf("初始化内存存储失败: %w", err)
	}
	
	return store, nil
}

// Initialize 初始化数据库
func (s *SQLiteMemoryStore) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建数据库目录
	dbDir := filepath.Dir(s.path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 打开数据库
	db, err := sql.Open("sqlite3", s.path)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	s.db = db

	// 创建表
	return s.createTables()
}

// createTables 创建数据库表
func (s *SQLiteMemoryStore) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL,
			prompt TEXT NOT NULL,
			result TEXT,
			error TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id)
		)`,
		
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT,
			tool_calls TEXT,
			name TEXT,
			tool_call_id TEXT,
			base64_image TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id)
		)`,
		
		`CREATE TABLE IF NOT EXISTS tool_calls (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			tool_name TEXT NOT NULL,
			arguments TEXT NOT NULL,
			result TEXT,
			error TEXT,
			duration_ms INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES tasks(id)
		)`,
		
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_conversation ON tasks(conversation_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_tool_calls_task ON tool_calls(task_id, created_at)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("创建表失败: %w", err)
		}
	}

	return nil
}

// AddMessage 添加消息
func (s *SQLiteMemoryStore) AddMessage(message schema.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 生成消息ID（如果没有）
	if message.ID == "" {
		message.ID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}

	// 序列化工具调用
	toolCallsJSON, err := json.Marshal(message.ToolCalls)
	if err != nil {
		return fmt.Errorf("序列化工具调用失败: %w", err)
	}

	query := `INSERT INTO messages (id, conversation_id, role, content, tool_calls, name, tool_call_id, base64_image, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err = s.db.Exec(query, message.ID, "default", string(message.Role), 
		message.Content, string(toolCallsJSON), message.Name, 
		message.ToolCallID, message.Base64Image, message.Timestamp)

	return err
}

// GetMessages 获取消息
func (s *SQLiteMemoryStore) GetMessages(conversationID string, limit int) ([]schema.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, role, content, tool_calls, name, tool_call_id, base64_image, created_at 
			  FROM messages WHERE conversation_id = ? ORDER BY created_at DESC`
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}
	defer rows.Close()

	var messages []schema.Message
	for rows.Next() {
		var msg schema.Message
		var toolCallsJSON string
		
		err := rows.Scan(&msg.ID, &msg.Role, &msg.Content, &toolCallsJSON, 
			&msg.Name, &msg.ToolCallID, &msg.Base64Image, &msg.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("扫描消息失败: %w", err)
		}

		// 反序列化工具调用
		if toolCallsJSON != "" {
			if err := json.Unmarshal([]byte(toolCallsJSON), &msg.ToolCalls); err != nil {
				return nil, fmt.Errorf("反序列化工具调用失败: %w", err)
			}
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// CreateTask 创建任务
func (s *SQLiteMemoryStore) CreateTask(task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO tasks (id, conversation_id, type, status, prompt, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	_, err := s.db.Exec(query, task.ID, task.ConversationID, task.Type, 
		task.Status, task.Prompt, task.CreatedAt, task.UpdatedAt)

	return err
}

// UpdateTask 更新任务
func (s *SQLiteMemoryStore) UpdateTask(taskID string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 构建更新查询
	setClause := ""
	args := []interface{}{}
	i := 0
	for key, value := range updates {
		if i > 0 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = ?", key)
		args = append(args, value)
		i++
	}
	
	if setClause == "" {
		return fmt.Errorf("没有更新字段")
	}

	// 添加更新时间
	setClause += ", updated_at = ?"
	args = append(args, time.Now())
	
	// 添加WHERE条件
	args = append(args, taskID)

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = ?", setClause)
	_, err := s.db.Exec(query, args...)

	return err
}

// GetTask 获取任务
func (s *SQLiteMemoryStore) GetTask(taskID string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, conversation_id, type, status, prompt, result, error, 
			  created_at, updated_at, completed_at FROM tasks WHERE id = ?`
	
	var task Task
	err := s.db.QueryRow(query, taskID).Scan(
		&task.ID, &task.ConversationID, &task.Type, &task.Status, 
		&task.Prompt, &task.Result, &task.Error, &task.CreatedAt, 
		&task.UpdatedAt, &task.CompletedAt)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	return &task, nil
}

// Close 关闭存储
func (s *SQLiteMemoryStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DeleteConversation 删除会话（及相关的消息和任务）
func (s *SQLiteMemoryStore) DeleteConversation(conversationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 删除相关数据
	queries := []string{
		"DELETE FROM tool_calls WHERE task_id IN (SELECT id FROM tasks WHERE conversation_id = ?)",
		"DELETE FROM messages WHERE conversation_id = ?",
		"DELETE FROM tasks WHERE conversation_id = ?",
		"DELETE FROM conversations WHERE id = ?",
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query, conversationID); err != nil {
			return fmt.Errorf("删除会话数据失败: %w", err)
		}
	}

	return nil
}

// GetTasks 获取任务列表
func (s *SQLiteMemoryStore) GetTasks(filter TaskFilter) ([]Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := "SELECT id, conversation_id, type, status, prompt, result, error, created_at, updated_at, completed_at FROM tasks WHERE 1=1"
	args := []interface{}{}

	if filter.ConversationID != "" {
		query += " AND conversation_id = ?"
		args = append(args, filter.ConversationID)
	}
	if filter.Type != "" {
		query += " AND type = ?"
		args = append(args, filter.Type)
	}
	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.ConversationID, &task.Type, &task.Status,
			&task.Prompt, &task.Result, &task.Error, &task.CreatedAt, &task.UpdatedAt, &task.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描任务失败: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// CreateConversation 创建会话
func (s *SQLiteMemoryStore) CreateConversation(conv Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO conversations (id, title, description, status, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err := s.db.Exec(query, conv.ID, conv.Title, conv.Description, 
		conv.Status, conv.CreatedAt, conv.UpdatedAt)

	return err
}

// GetConversation 获取会话
func (s *SQLiteMemoryStore) GetConversation(convID string) (*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, title, description, status, created_at, updated_at 
			  FROM conversations WHERE id = ?`
	
	var conv Conversation
	err := s.db.QueryRow(query, convID).Scan(
		&conv.ID, &conv.Title, &conv.Description, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("会话不存在: %s", convID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}

	return &conv, nil
}

// GetConversations 获取会话列表
func (s *SQLiteMemoryStore) GetConversations(filter ConversationFilter) ([]Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := "SELECT id, title, description, status, created_at, updated_at FROM conversations WHERE 1=1"
	args := []interface{}{}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	query += " ORDER BY updated_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.Title, &conv.Description, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描会话失败: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// RecordToolCall 记录工具调用
func (s *SQLiteMemoryStore) RecordToolCall(call ToolCallRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	argumentsJSON, err := json.Marshal(call.Arguments)
	if err != nil {
		return fmt.Errorf("序列化参数失败: %w", err)
	}

	resultJSON, err := json.Marshal(call.Result)
	if err != nil {
		return fmt.Errorf("序列化结果失败: %w", err)
	}

	query := `INSERT INTO tool_calls (id, task_id, tool_name, arguments, result, error, duration_ms, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err = s.db.Exec(query, call.ID, call.TaskID, call.ToolName, string(argumentsJSON),
		string(resultJSON), call.Error, call.Duration, call.CreatedAt)

	return err
}

// GetToolCalls 获取工具调用记录
func (s *SQLiteMemoryStore) GetToolCalls(taskID string) ([]ToolCallRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, task_id, tool_name, arguments, result, error, duration_ms, created_at 
			  FROM tool_calls WHERE task_id = ? ORDER BY created_at`
	
	rows, err := s.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询工具调用失败: %w", err)
	}
	defer rows.Close()

	var calls []ToolCallRecord
	for rows.Next() {
		var call ToolCallRecord
		var argumentsJSON, resultJSON string
		
		err := rows.Scan(&call.ID, &call.TaskID, &call.ToolName, &argumentsJSON,
			&resultJSON, &call.Error, &call.Duration, &call.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描工具调用失败: %w", err)
		}

		// 反序列化参数和结果
		if err := json.Unmarshal([]byte(argumentsJSON), &call.Arguments); err != nil {
			return nil, fmt.Errorf("反序列化参数失败: %w", err)
		}
		if err := json.Unmarshal([]byte(resultJSON), &call.Result); err != nil {
			return nil, fmt.Errorf("反序列化结果失败: %w", err)
		}

		calls = append(calls, call)
	}

	return calls, nil
}