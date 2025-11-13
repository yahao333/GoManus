package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yahao333/GoManus/internal/logger"
	"go.uber.org/zap"
)

// PythonBridge Python桥接层
type PythonBridge struct {
	pythonPath string
	envPath    string
	workDir    string
}

// NewPythonBridge 创建新的Python桥接层
func NewPythonBridge(workDir string) (*PythonBridge, error) {
	// 查找Python解释器
	pythonPath, err := findPython()
	if err != nil {
		return nil, fmt.Errorf("未找到Python解释器: %w", err)
	}

	bridge := &PythonBridge{
		pythonPath: pythonPath,
		workDir:    workDir,
	}

	// 创建虚拟环境（如果不存在）
	if err := bridge.setupVirtualEnv(); err != nil {
		logger.Warn("创建虚拟环境失败，使用系统Python", zap.Error(err))
	}

	return bridge, nil
}

// findPython 查找Python解释器
func findPython() (string, error) {
	// 尝试不同的Python命令
	pythonCommands := []string{"python3", "python", "py"}
	
	for _, cmd := range pythonCommands {
		if path, err := exec.LookPath(cmd); err == nil {
			return path, nil
		}
	}
	
	return "", fmt.Errorf("未找到Python解释器")
}

// setupVirtualEnv 设置虚拟环境
func (b *PythonBridge) setupVirtualEnv() error {
	envPath := filepath.Join(b.workDir, ".venv")
	
	// 检查虚拟环境是否已存在
	if _, err := os.Stat(envPath); err == nil {
		b.envPath = envPath
		return nil
	}

	// 创建虚拟环境
	cmd := exec.Command(b.pythonPath, "-m", "venv", envPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("创建虚拟环境失败: %w", err)
	}

	b.envPath = envPath
	return nil
}

// getPythonPath 获取Python路径
func (b *PythonBridge) getPythonPath() string {
	if b.envPath != "" {
		// 根据操作系统返回正确的Python路径
		if isWindows() {
			return filepath.Join(b.envPath, "Scripts", "python.exe")
		}
		return filepath.Join(b.envPath, "bin", "python")
	}
	return b.pythonPath
}

// InstallRequirements 安装Python依赖
func (b *PythonBridge) InstallRequirements(requirements []string) error {
	pythonPath := b.getPythonPath()
	
	// 安装pip（如果需要）
	if err := b.ensurePip(); err != nil {
		return fmt.Errorf("确保pip可用失败: %w", err)
	}

	for _, req := range requirements {
		cmd := exec.Command(pythonPath, "-m", "pip", "install", req)
		cmd.Dir = b.workDir
		
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("安装依赖 %s 失败: %w, 输出: %s", req, err, string(output))
		}
		
		logger.Info("已安装Python依赖", zap.String("package", req))
	}

	return nil
}

// ensurePip 确保pip可用
func (b *PythonBridge) ensurePip() error {
	pythonPath := b.getPythonPath()
	cmd := exec.Command(pythonPath, "-m", "pip", "--version")
	
	if err := cmd.Run(); err != nil {
		// 尝试安装pip
		getPipCmd := exec.Command(pythonPath, "-c", "import urllib.request; urllib.request.urlretrieve('https://bootstrap.pypa.io/get-pip.py', 'get-pip.py')")
		if err := getPipCmd.Run(); err != nil {
			return fmt.Errorf("下载pip安装脚本失败: %w", err)
		}
		
		installPipCmd := exec.Command(pythonPath, "get-pip.py")
		if err := installPipCmd.Run(); err != nil {
			return fmt.Errorf("安装pip失败: %w", err)
		}
		
		os.Remove("get-pip.py")
	}
	
	return nil
}

// ExecuteScript 执行Python脚本
func (b *PythonBridge) ExecuteScript(ctx context.Context, script string, args map[string]interface{}) (interface{}, error) {
	pythonPath := b.getPythonPath()
	
	// 准备参数
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("序列化参数失败: %w", err)
	}

	// 创建完整的Python脚本
	fullScript := fmt.Sprintf(`
import json
import sys

# 解析参数
args = json.loads('%s')

# 用户脚本
%s

# 执行并返回结果
try:
    result = main(args)
    print(json.dumps({"success": True, "result": result}))
except Exception as e:
    print(json.dumps({"success": False, "error": str(e)}))
`, string(argsJSON), script)

	// 执行脚本
	cmd := exec.CommandContext(ctx, pythonPath, "-c", fullScript)
	cmd.Dir = b.workDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("执行Python脚本失败: %w, stderr: %s", err, stderr.String())
	}

	// 解析结果
	var result struct {
		Success bool        `json:"success"`
		Result  interface{} `json:"result"`
		Error   string      `json:"error"`
	}
	
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("解析Python输出失败: %w, 输出: %s", err, stdout.String())
	}

	if !result.Success {
		return nil, fmt.Errorf("Python脚本执行错误: %s", result.Error)
	}

	return result.Result, nil
}

// ExecuteFile 执行Python文件
func (b *PythonBridge) ExecuteFile(ctx context.Context, scriptPath string, args map[string]interface{}) (interface{}, error) {
	pythonPath := b.getPythonPath()
	
	// 准备参数
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("序列化参数失败: %w", err)
	}

	// 创建包装脚本
	wrapperScript := fmt.Sprintf(`
import json
import sys
sys.path.append('%s')

# 解析参数
args = json.loads('%s')

# 导入并执行用户脚本
exec(open('%s').read())

# 执行并返回结果
try:
    result = main(args)
    print(json.dumps({"success": True, "result": result}))
except Exception as e:
    print(json.dumps({"success": False, "error": str(e)}))
`, b.workDir, string(argsJSON), scriptPath)

	// 执行包装脚本
	cmd := exec.CommandContext(ctx, pythonPath, "-c", wrapperScript)
	cmd.Dir = b.workDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("执行Python文件失败: %w, stderr: %s", err, stderr.String())
	}

	// 解析结果
	var result struct {
		Success bool        `json:"success"`
		Result  interface{} `json:"result"`
		Error   string      `json:"error"`
	}
	
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("解析Python输出失败: %w, 输出: %s", err, stdout.String())
	}

	if !result.Success {
		return nil, fmt.Errorf("Python脚本执行错误: %s", result.Error)
	}

	return result.Result, nil
}

// CallLLM 调用LLM API（通过Python）
func (b *PythonBridge) CallLLM(ctx context.Context, provider, model string, messages []map[string]interface{}, options map[string]interface{}) (interface{}, error) {
	args := map[string]interface{}{
		"provider": provider,
		"model":    model,
		"messages": messages,
		"options":  options,
	}

	script := `
import openai
import json

def main(args):
    provider = args["provider"]
    model = args["model"]
    messages = args["messages"]
    options = args.get("options", {})
    
    if provider == "openai":
        client = openai.OpenAI(api_key=options.get("api_key"))
        response = client.chat.completions.create(
            model=model,
            messages=messages,
            **options
        )
        return {
            "content": response.choices[0].message.content,
            "usage": response.usage.dict()
        }
    else:
        raise ValueError(f"不支持的提供者: {provider}")
`

	return b.ExecuteScript(ctx, script, args)
}

// Close 关闭桥接层
func (b *PythonBridge) Close() error {
	// 清理虚拟环境（可选）
	return nil
}

// isWindows 检查是否为Windows系统
func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

// GetPythonVersion 获取Python版本
func (b *PythonBridge) GetPythonVersion() (string, error) {
	pythonPath := b.getPythonPath()
	cmd := exec.Command(pythonPath, "--version")
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取Python版本失败: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// InstallOpenAI 安装OpenAI库
func (b *PythonBridge) InstallOpenAI() error {
	return b.InstallRequirements([]string{"openai>=1.0.0"})
}

// InstallLangChain 安装LangChain库
func (b *PythonBridge) InstallLangChain() error {
	return b.InstallRequirements([]string{"langchain>=0.1.0", "langchain-openai>=0.0.2"})
}