package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"go.uber.org/zap"
)

// SSESession SSE传输的MCP会话
type SSESession struct {
	serverURL string
	client    *http.Client
	logger    *zap.Logger
	mu        sync.Mutex
}

// NewSSESession 创建新的SSE会话
func NewSSESession(serverURL string, logger *zap.Logger) *SSESession {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SSESession{
		serverURL: serverURL,
		client:    &http.Client{},
		logger:    logger,
	}
}

// Initialize 初始化会话
func (s *SSESession) Initialize(ctx context.Context) error {
	// 验证服务器URL
	u, err := url.Parse(s.serverURL)
	if err != nil {
		return fmt.Errorf("无效的SSE服务器URL: %w", err)
	}

	// 测试连接
	endpoint := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	if u.RawQuery != "" {
		endpoint = fmt.Sprintf("%s?%s", endpoint, u.RawQuery)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("连接SSE服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE服务器返回错误状态: %d", resp.StatusCode)
	}

	// 验证响应头
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !isSSEContentType(contentType) {
		s.logger.Warn("SSE服务器返回非标准Content-Type", zap.String("content_type", contentType))
	}

	s.logger.Info("SSE会话初始化成功", zap.String("url", s.serverURL))
	return nil
}

// ListTools 列出可用工具
func (s *SSESession) ListTools(ctx context.Context) (*ListToolsResult, error) {
	// 构建请求URL
	u, err := url.Parse(s.serverURL)
	if err != nil {
		return nil, fmt.Errorf("解析URL失败: %w", err)
	}

	// 移除路径中的查询参数，添加/tools路径
	u.RawQuery = ""
	toolsURL := fmt.Sprintf("%s://%s/tools", u.Scheme, u.Host)

	req, err := http.NewRequestWithContext(ctx, "GET", toolsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求工具列表失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("服务器返回错误: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result ListToolsResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	s.logger.Info("获取工具列表成功", zap.Int("tool_count", len(result.Tools)))
	return &result, nil
}

// CallTool 调用工具
func (s *SSESession) CallTool(ctx context.Context, name string, args map[string]interface{}) (*CallToolResult, error) {
	// 构建请求
	u, err := url.Parse(s.serverURL)
	if err != nil {
		return nil, fmt.Errorf("解析URL失败: %w", err)
	}

	// 构建调用URL
	callURL := fmt.Sprintf("%s://%s/tools/call", u.Scheme, u.Host)

	requestBody := map[string]interface{}{
		"tool": name,
		"arguments": args,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", callURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(jsonBody))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用工具失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("工具调用失败: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result CallToolResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	s.logger.Info("工具调用成功", zap.String("tool", name))
	return &result, nil
}

// Close 关闭会话
func (s *SSESession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// HTTP客户端不需要特殊清理
	s.logger.Info("SSE会话关闭", zap.String("url", s.serverURL))
	return nil
}

// isSSEContentType 检查是否为SSE内容类型
func isSSEContentType(contentType string) bool {
	return contentType == "text/event-stream" || 
		   contentType == "application/x-ndjson" ||
		   contentType == "text/plain"
}