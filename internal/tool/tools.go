package tool

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    "github.com/yahao333/GoManus/pkg/config"
    "github.com/yahao333/GoManus/internal/logger"
    "go.uber.org/zap"
)

// PythonExecute Python执行工具
type PythonExecute struct {
	BaseTool
}

// NewPythonExecute 创建Python执行工具
func NewPythonExecute() *PythonExecute {
	return &PythonExecute{
		BaseTool: BaseTool{
			Name:        "PythonExecute",
			Description: "执行Python代码",
			Parameters: map[string]interface{}{
				"code": map[string]interface{}{
					"type":        "string",
					"description": "要执行的Python代码",
				},
			},
			Required: []string{"code"},
		},
	}
}

// Execute 执行Python代码
func (p *PythonExecute) Execute(ctx context.Context, arguments string) (interface{}, error) {
	args, err := parseArguments(arguments)
	if err != nil {
		return nil, err
	}

	if err := validateArguments(args, p.Required); err != nil {
		return nil, err
	}

	code, ok := args["code"].(string)
	if !ok {
		return nil, fmt.Errorf("参数code必须是字符串")
	}

	logger.Info("执行Python代码", zap.String("code", code))

	// 创建工作目录
	workDir := config.GetConfig().GetWorkspaceRoot()
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("创建工作目录失败: %w", err)
	}

	// 创建临时文件
	tempFile := filepath.Join(workDir, fmt.Sprintf("python_script_%d.py", time.Now().Unix()))
	if err := os.WriteFile(tempFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("写入临时文件失败: %w", err)
	}
	defer os.Remove(tempFile)

	// 执行Python代码
	cmd := exec.CommandContext(ctx, "python3", tempFile)
	cmd.Dir = workDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"output": string(output),
			"error":  err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"output": string(output),
		"success": true,
	}, nil
}

// StrReplaceEditor 文件编辑工具
type StrReplaceEditor struct {
	BaseTool
}

// NewStrReplaceEditor 创建文件编辑工具
func NewStrReplaceEditor() *StrReplaceEditor {
	return &StrReplaceEditor{
		BaseTool: BaseTool{
			Name:        "StrReplaceEditor",
			Description: "编辑文件内容",
			Parameters: map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "命令类型: create, view, str_replace",
					"enum":        []string{"create", "view", "str_replace"},
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "文件路径",
				},
				"file_text": map[string]interface{}{
					"type":        "string",
					"description": "文件内容（create命令时使用）",
				},
				"old_str": map[string]interface{}{
					"type":        "string",
					"description": "要替换的字符串（str_replace命令时使用）",
				},
				"new_str": map[string]interface{}{
					"type":        "string",
					"description": "替换后的字符串（str_replace命令时使用）",
				},
			},
			Required: []string{"command", "path"},
		},
	}
}

// Execute 执行文件编辑
func (s *StrReplaceEditor) Execute(ctx context.Context, arguments string) (interface{}, error) {
	args, err := parseArguments(arguments)
	if err != nil {
		return nil, err
	}

	if err := validateArguments(args, []string{"command", "path"}); err != nil {
		return nil, err
	}

	command, _ := args["command"].(string)
	path, _ := args["path"].(string)

	logger.Info("执行文件编辑", 
		zap.String("command", command),
		zap.String("path", path))

	switch command {
	case "create":
		return s.createFile(path, args)
	case "view":
		return s.viewFile(path)
	case "str_replace":
		return s.strReplace(path, args)
	default:
		return nil, fmt.Errorf("不支持的命令: %s", command)
	}
}

// createFile 创建文件
func (s *StrReplaceEditor) createFile(path string, args map[string]interface{}) (interface{}, error) {
	fileText, ok := args["file_text"].(string)
	if !ok {
		return nil, fmt.Errorf("创建文件需要提供file_text参数")
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(path, []byte(fileText), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	return map[string]interface{}{
		"message": "文件创建成功",
		"path":    path,
	}, nil
}

// viewFile 查看文件
func (s *StrReplaceEditor) viewFile(path string) (interface{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return map[string]interface{}{
		"content": string(content),
		"path":    path,
	}, nil
}

// strReplace 字符串替换
func (s *StrReplaceEditor) strReplace(path string, args map[string]interface{}) (interface{}, error) {
	oldStr, ok := args["old_str"].(string)
	if !ok {
		return nil, fmt.Errorf("str_replace命令需要提供old_str参数")
	}

	newStr, ok := args["new_str"].(string)
	if !ok {
		return nil, fmt.Errorf("str_replace命令需要提供new_str参数")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	newContent := strings.ReplaceAll(string(content), oldStr, newStr)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	return map[string]interface{}{
		"message": "字符串替换成功",
		"path":    path,
		"old_str": oldStr,
		"new_str": newStr,
	}, nil
}

// AskHuman 人类提问工具
type AskHuman struct {
	BaseTool
}

// NewAskHuman 创建人类提问工具
func NewAskHuman() *AskHuman {
	return &AskHuman{
		BaseTool: BaseTool{
			Name:        "AskHuman",
			Description: "向用户提问",
			Parameters: map[string]interface{}{
				"question": map[string]interface{}{
					"type":        "string",
					"description": "要提问的问题",
				},
			},
			Required: []string{"question"},
		},
	}
}

// Execute 执行提问
func (a *AskHuman) Execute(ctx context.Context, arguments string) (interface{}, error) {
	args, err := parseArguments(arguments)
	if err != nil {
		return nil, err
	}

	if err := validateArguments(args, a.Required); err != nil {
		return nil, err
	}

	question, _ := args["question"].(string)

	logger.Info("向用户提问", zap.String("question", question))

	// 在实际实现中，这里应该等待用户输入
	// 为了简化，返回一个模拟的响应
	return map[string]interface{}{
		"question": question,
		"answer":   "用户回答: 继续执行任务",
		"note":     "这是一个模拟响应，实际使用时需要实现用户输入机制",
	}, nil
}

// Terminate 终止工具
type Terminate struct {
	BaseTool
}

// NewTerminate 创建终止工具
func NewTerminate() *Terminate {
	return &Terminate{
		BaseTool: BaseTool{
			Name:        "Terminate",
			Description: "完成任务并终止执行",
			Parameters: map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "完成消息",
				},
			},
			Required: []string{"message"},
		},
	}
}

// Execute 执行终止
func (t *Terminate) Execute(ctx context.Context, arguments string) (interface{}, error) {
	args, err := parseArguments(arguments)
	if err != nil {
		return nil, err
	}

	if err := validateArguments(args, t.Required); err != nil {
		return nil, err
	}

	message, _ := args["message"].(string)

	logger.Info("任务完成", zap.String("message", message))

	return map[string]interface{}{
		"message": message,
		"status":  "completed",
	}, nil
}

// BrowserUseTool 浏览器工具
type BrowserUseTool struct {
	BaseTool
}

// NewBrowserUseTool 创建浏览器工具
func NewBrowserUseTool() *BrowserUseTool {
	return &BrowserUseTool{
		BaseTool: BaseTool{
			Name:        "BrowserUseTool",
			Description: "使用浏览器访问网页",
			Parameters: map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "要访问的URL",
				},
				"action": map[string]interface{}{
					"type":        "string",
					"description": "操作类型: visit, click, fill, screenshot",
					"enum":        []string{"visit", "click", "fill", "screenshot"},
				},
				"selector": map[string]interface{}{
					"type":        "string",
					"description": "CSS选择器（click和fill操作时使用）",
				},
				"text": map[string]interface{}{
					"type":        "string",
					"description": "要填充的文本（fill操作时使用）",
				},
			},
			Required: []string{"url", "action"},
		},
	}
}

// Execute 执行浏览器操作
func (b *BrowserUseTool) Execute(ctx context.Context, arguments string) (interface{}, error) {
	args, err := parseArguments(arguments)
	if err != nil {
		return nil, err
	}

	if err := validateArguments(args, []string{"url", "action"}); err != nil {
		return nil, err
	}

	url, _ := args["url"].(string)
	action, _ := args["action"].(string)

	logger.Info("执行浏览器操作", 
		zap.String("url", url),
		zap.String("action", action))

	// 这里应该实现实际的浏览器操作
	// 为了简化，返回模拟结果
	switch action {
	case "visit":
		return map[string]interface{}{
			"url":     url,
			"action":  action,
			"status":  "visited",
			"content": "模拟网页内容",
		}, nil
	case "click":
		selector, _ := args["selector"].(string)
		return map[string]interface{}{
			"url":      url,
			"action":   action,
			"selector": selector,
			"status":   "clicked",
		}, nil
	case "fill":
		selector, _ := args["selector"].(string)
		text, _ := args["text"].(string)
		return map[string]interface{}{
			"url":      url,
			"action":   action,
			"selector": selector,
			"text":     text,
			"status":   "filled",
		}, nil
	case "screenshot":
		return map[string]interface{}{
			"url":      url,
			"action":   action,
			"status":   "screenshot_taken",
			"image":    "模拟截图数据",
		}, nil
	default:
		return nil, fmt.Errorf("不支持的操作: %s", action)
	}
}
