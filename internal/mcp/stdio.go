package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"go.uber.org/zap"
)

// StdioSession stdio传输的MCP会话
type StdioSession struct {
	command string
	args    []string
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	logger  *zap.Logger
	mu      sync.Mutex
	closed  bool
}

// NewStdioSession 创建新的stdio会话
func NewStdioSession(command string, args []string, logger *zap.Logger) *StdioSession {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &StdioSession{
		command: command,
		args:    args,
		logger:  logger,
	}
}

// Initialize 初始化会话
func (s *StdioSession) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("会话已关闭")
	}

	// 创建命令
	s.cmd = exec.CommandContext(ctx, s.command, s.args...)

	// 获取标准输入输出
	stdin, err := s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("创建stdin管道失败: %w", err)
	}

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %w", err)
	}

	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建stderr管道失败: %w", err)
	}

	s.stdin = stdin
	s.stdout = stdout
	s.stderr = stderr

	// 启动命令
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("启动命令失败: %w", err)
	}

	// 启动错误监控
	go s.monitorErrors()

	s.logger.Info("stdio会话初始化成功", zap.String("command", s.command), zap.Strings("args", s.args))
	return nil
}

// monitorErrors 监控错误输出
func (s *StdioSession) monitorErrors() {
	scanner := bufio.NewScanner(s.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		s.logger.Warn("MCP服务器错误输出", zap.String("stderr", line))
	}
	if err := scanner.Err(); err != nil {
		s.logger.Error("读取stderr失败", zap.Error(err))
	}
}

// sendRequest 发送请求并等待响应
func (s *StdioSession) sendRequest(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, fmt.Errorf("会话已关闭")
	}

	// 序列化请求
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 发送请求（添加换行符作为分隔符）
	_, err = fmt.Fprintf(s.stdin, "%s\n", requestData)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	// 创建带超时的读取器
	done := make(chan []byte, 1)
	errChan := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(s.stdout)
		if scanner.Scan() {
			done <- scanner.Bytes()
		} else {
			if err := scanner.Err(); err != nil {
				errChan <- fmt.Errorf("读取响应失败: %w", err)
			} else {
				errChan <- fmt.Errorf("未收到响应")
			}
		}
	}()

	// 等待响应或上下文取消
	select {
	case responseData := <-done:
		var response map[string]interface{}
		if err := json.Unmarshal(responseData, &response); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}
		return response, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("请求被取消: %w", ctx.Err())
	}
}

// ListTools 列出可用工具
func (s *StdioSession) ListTools(ctx context.Context) (*ListToolsResult, error) {
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/list",
		"id":      generateRequestID(),
	}

	response, err := s.sendRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("列出工具失败: %w", err)
	}

	// 检查错误
	if errMsg, exists := response["error"]; exists {
		return nil, fmt.Errorf("服务器返回错误: %v", errMsg)
	}

	// 解析结果
	resultData, err := json.Marshal(response["result"])
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %w", err)
	}

	var result ListToolsResult
	if err := json.Unmarshal(resultData, &result); err != nil {
		return nil, fmt.Errorf("解析结果失败: %w", err)
	}

	s.logger.Info("获取工具列表成功", zap.Int("tool_count", len(result.Tools)))
	return &result, nil
}

// CallTool 调用工具
func (s *StdioSession) CallTool(ctx context.Context, name string, args map[string]interface{}) (*CallToolResult, error) {
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"id":      generateRequestID(),
		"params": map[string]interface{}{
			"name":      name,
			"arguments": args,
		},
	}

	response, err := s.sendRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("调用工具失败: %w", err)
	}

	// 检查错误
	if errMsg, exists := response["error"]; exists {
		return nil, fmt.Errorf("服务器返回错误: %v", errMsg)
	}

	// 解析结果
	resultData, err := json.Marshal(response["result"])
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %w", err)
	}

	var result CallToolResult
	if err := json.Unmarshal(resultData, &result); err != nil {
		return nil, fmt.Errorf("解析结果失败: %w", err)
	}

	s.logger.Info("工具调用成功", zap.String("tool", name))
	return &result, nil
}

// Close 关闭会话
func (s *StdioSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// 关闭标准输入
	if s.stdin != nil {
		s.stdin.Close()
	}

	// 等待进程结束或强制终止
	if s.cmd != nil && s.cmd.Process != nil {
		// 先尝试优雅终止
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case err := <-done:
			if err != nil {
				s.logger.Warn("进程退出时发生错误", zap.Error(err))
			}
		case <-time.After(5 * time.Second):
			// 超时后强制终止
			if err := s.cmd.Process.Kill(); err != nil {
				s.logger.Error("强制终止进程失败", zap.Error(err))
			} else {
				s.logger.Info("强制终止进程")
			}
		}
	}

	s.logger.Info("stdio会话关闭", zap.String("command", s.command))
	return nil
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
