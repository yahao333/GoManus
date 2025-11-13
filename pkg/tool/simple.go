package tool

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/yahao333/GoManus/pkg/logger"
    "go.uber.org/zap"
)

// SimpleBrowser 简化浏览器工具
type SimpleBrowser struct {
	BaseTool
	client *http.Client
}

// NewSimpleBrowser 创建简化浏览器工具
func NewSimpleBrowser() *SimpleBrowser {
	return &SimpleBrowser{
		BaseTool: BaseTool{
			Name:        "SimpleBrowser",
			Description: "简单的HTTP浏览器工具",
			Parameters: map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "要访问的URL",
				},
				"method": map[string]interface{}{
					"type":        "string",
					"description": "HTTP方法: GET, POST",
					"enum":        []string{"GET", "POST"},
					"default":     "GET",
				},
				"headers": map[string]interface{}{
					"type":        "object",
					"description": "HTTP请求头",
					"default":     map[string]string{},
				},
				"body": map[string]interface{}{
					"type":        "string",
					"description": "请求体（POST方法时使用）",
				},
			},
			Required: []string{"url"},
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Execute 执行浏览器操作
func (s *SimpleBrowser) Execute(ctx context.Context, arguments string) (interface{}, error) {
	args, err := parseArguments(arguments)
	if err != nil {
		return nil, err
	}

	if err := validateArguments(args, s.Required); err != nil {
		return nil, err
	}

	url, _ := args["url"].(string)
	method := "GET"
	if methodArg, ok := args["method"].(string); ok {
		method = methodArg
	}

	logger.Info("执行浏览器请求", 
		zap.String("url", url),
		zap.String("method", method))

	// 创建请求
	var req *http.Request
	var reqErr error

	if method == "POST" {
		body := ""
		if bodyArg, ok := args["body"].(string); ok {
			body = bodyArg
		}
		req, reqErr = http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	} else {
		req, reqErr = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if reqErr != nil {
		return nil, fmt.Errorf("创建请求失败: %w", reqErr)
	}

	// 设置请求头
	if headers, ok := args["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}
	}

	// 执行请求
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	var result strings.Builder
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			result.Write(buffer[:n])
		}
		if err != nil {
			break
		}
	}

	// 截断内容（避免太长）
	content := result.String()
	if len(content) > 5000 {
		content = content[:5000] + "..."
	}

	return map[string]interface{}{
		"url":        url,
		"method":     method,
		"status_code": resp.StatusCode,
		"status":     resp.Status,
		"headers":    resp.Header,
		"content":    content,
		"length":     len(content),
	}, nil
}

// SimpleSearch 简化搜索工具
type SimpleSearch struct {
	BaseTool
}

// NewSimpleSearch 创建简化搜索工具
func NewSimpleSearch() *SimpleSearch {
	return &SimpleSearch{
		BaseTool: BaseTool{
			Name:        "SimpleSearch",
			Description: "简单的网络搜索工具",
			Parameters: map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "搜索查询",
				},
				"engine": map[string]interface{}{
					"type":        "string",
					"description": "搜索引擎: google, bing, duckduckgo",
					"enum":        []string{"google", "bing", "duckduckgo"},
					"default":     "duckduckgo",
				},
				"num_results": map[string]interface{}{
					"type":        "integer",
					"description": "结果数量",
					"default":     5,
				},
			},
			Required: []string{"query"},
		},
	}
}

// Execute 执行搜索
func (s *SimpleSearch) Execute(ctx context.Context, arguments string) (interface{}, error) {
	args, err := parseArguments(arguments)
	if err != nil {
		return nil, err
	}

	if err := validateArguments(args, s.Required); err != nil {
		return nil, err
	}

	query, _ := args["query"].(string)
	engine := "duckduckgo"
	if engineArg, ok := args["engine"].(string); ok {
		engine = engineArg
	}

	numResults := 5
	if numArg, ok := args["num_results"].(float64); ok {
		numResults = int(numArg)
	}

	logger.Info("执行搜索", 
		zap.String("query", query),
		zap.String("engine", engine),
		zap.Int("num_results", numResults))

	// 构建搜索URL
	var searchURL string
	switch engine {
	case "google":
		searchURL = fmt.Sprintf("https://www.google.com/search?q=%s&num=%d", 
			strings.ReplaceAll(query, " ", "+"), numResults)
	case "bing":
		searchURL = fmt.Sprintf("https://www.bing.com/search?q=%s&count=%d", 
			strings.ReplaceAll(query, " ", "+"), numResults)
	default: // duckduckgo
		searchURL = fmt.Sprintf("https://duckduckgo.com/?q=%s&kl=us-en", 
			strings.ReplaceAll(query, " ", "+"))
	}

	// 使用浏览器工具获取搜索结果
	browser := NewSimpleBrowser()
	browserArgs, _ := json.Marshal(map[string]interface{}{
		"url": searchURL,
	})

	_, err = browser.Execute(ctx, string(browserArgs))
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	// 简化搜索结果（实际实现中需要解析HTML）
	return map[string]interface{}{
		"query":        query,
		"engine":       engine,
		"search_url":   searchURL,
		"results":      "模拟搜索结果",
		"num_results":  numResults,
		"note":         "这是简化的搜索结果，实际实现需要解析HTML",
	}, nil
}
