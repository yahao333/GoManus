package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Tool 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ListToolsResult 工具列表结果
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolRequest 工具调用请求
type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
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

// 示例工具
var tools = []Tool{
	{
		Name:        "hello",
		Description: "Say hello to someone",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the person to greet",
				},
			},
			"required": []string{"name"},
		},
	},
	{
		Name:        "add",
		Description: "Add two numbers",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{
					"type":        "number",
					"description": "First number",
				},
				"b": map[string]interface{}{
					"type":        "number",
					"description": "Second number",
				},
			},
			"required": []string{"a", "b"},
		},
	},
}

func main() {
	// 工具列表接口
	http.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		result := ListToolsResult{Tools: tools}
		json.NewEncoder(w).Encode(result)
	})

	// 工具调用接口
	http.HandleFunc("/tools/call", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CallToolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 执行工具
		result, err := executeTool(req.Name, req.Arguments)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// SSE接口
	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// 发送初始事件
		fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
		w.(http.Flusher).Flush()

		// 保持连接
		<-r.Context().Done()
	})

	fmt.Println("MCP测试服务器启动在 http://localhost:8080")
	fmt.Println("可用端点:")
	fmt.Println("  GET  /tools      - 获取工具列表")
	fmt.Println("  POST /tools/call - 调用工具")
	fmt.Println("  GET  /sse        - SSE连接")
	fmt.Println("")
	fmt.Println("示例工具:")
	fmt.Println("  hello - 向某人问好")
	fmt.Println("  add   - 添加两个数字")
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func executeTool(name string, args map[string]interface{}) (*CallToolResult, error) {
	switch name {
	case "hello":
		name, ok := args["name"].(string)
		if !ok {
			return nil, fmt.Errorf("name must be a string")
		}
		return &CallToolResult{
			Content: []Content{
				{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)},
			},
		}, nil

	case "add":
		a, ok := args["a"].(float64)
		if !ok {
			return nil, fmt.Errorf("a must be a number")
		}
		b, ok := args["b"].(float64)
		if !ok {
			return nil, fmt.Errorf("b must be a number")
		}
		result := a + b
		return &CallToolResult{
			Content: []Content{
				{Type: "text", Text: fmt.Sprintf("%.2f", result)},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}