package sandbox

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    "github.com/yahao333/GoManus/pkg/config"
    "github.com/yahao333/GoManus/pkg/logger"
    "go.uber.org/zap"
)

// Sandbox 沙盒接口
type Sandbox interface {
	Create(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Remove(ctx context.Context) error
	Execute(ctx context.Context, command string, timeout time.Duration) (string, error)
	GetStatus() string
}

// LocalSandbox 本地沙盒实现
type LocalSandbox struct {
	workDir     string
	tempDir     string
	status      string
	config      *config.SandboxSettings
}

// NewLocalSandbox 创建新的本地沙盒
func NewLocalSandbox(config *config.SandboxSettings) (*LocalSandbox, error) {
	return &LocalSandbox{
		workDir: config.WorkDir,
		status:  "created",
		config:  config,
	}, nil
}

// Create 创建沙盒
func (l *LocalSandbox) Create(ctx context.Context) error {
	logger.Info("创建本地沙盒")

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "gomanus_sandbox_*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	l.tempDir = tempDir
	l.status = "created"

	logger.Info("本地沙盒创建成功", zap.String("temp_dir", tempDir))
	return nil
}

// Start 启动沙盒
func (l *LocalSandbox) Start(ctx context.Context) error {
	if l.tempDir == "" {
		return fmt.Errorf("沙盒未创建")
	}

	logger.Info("启动本地沙盒", zap.String("temp_dir", l.tempDir))
	l.status = "running"
	return nil
}

// Stop 停止沙盒
func (l *LocalSandbox) Stop(ctx context.Context) error {
	if l.tempDir == "" {
		return fmt.Errorf("沙盒未创建")
	}

	logger.Info("停止本地沙盒")
	l.status = "stopped"
	return nil
}

// Remove 移除沙盒
func (l *LocalSandbox) Remove(ctx context.Context) error {
	if l.tempDir == "" {
		return fmt.Errorf("沙盒未创建")
	}

	logger.Info("移除本地沙盒", zap.String("temp_dir", l.tempDir))

	// 清理临时目录
	if err := os.RemoveAll(l.tempDir); err != nil {
		return fmt.Errorf("清理临时目录失败: %w", err)
	}

	l.tempDir = ""
	l.status = "removed"
	return nil
}

// Execute 在沙盒中执行命令
func (l *LocalSandbox) Execute(ctx context.Context, command string, timeout time.Duration) (string, error) {
	if l.tempDir == "" {
		return "", fmt.Errorf("沙盒未创建")
	}

	if l.status != "running" {
		return "", fmt.Errorf("沙盒未运行")
	}

	logger.Info("执行命令", 
		zap.String("command", command),
		zap.String("work_dir", l.tempDir))

	// 创建命令
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = l.tempDir
	
	// 设置超时
	if timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cmd = exec.CommandContext(timeoutCtx, "sh", "-c", command)
	}

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %w", err)
	}

	return string(output), nil
}

// GetStatus 获取沙盒状态
func (l *LocalSandbox) GetStatus() string {
	return l.status
}

// GetWorkDir 获取工作目录
func (l *LocalSandbox) GetWorkDir() string {
	return l.tempDir
}

// SandboxManager 沙盒管理器
type SandboxManager struct {
	sandboxes map[string]Sandbox
	config    *config.SandboxSettings
}

// NewSandboxManager 创建新的沙盒管理器
func NewSandboxManager(config *config.SandboxSettings) *SandboxManager {
	return &SandboxManager{
		sandboxes: make(map[string]Sandbox),
		config:    config,
	}
}

// CreateSandbox 创建沙盒
func (sm *SandboxManager) CreateSandbox(id string) (Sandbox, error) {
	if _, exists := sm.sandboxes[id]; exists {
		return nil, fmt.Errorf("沙盒已存在: %s", id)
	}

	sandbox, err := NewLocalSandbox(sm.config)
	if err != nil {
		return nil, err
	}

	sm.sandboxes[id] = sandbox
	return sandbox, nil
}

// GetSandbox 获取沙盒
func (sm *SandboxManager) GetSandbox(id string) (Sandbox, error) {
	sandbox, exists := sm.sandboxes[id]
	if !exists {
		return nil, fmt.Errorf("沙盒不存在: %s", id)
	}
	return sandbox, nil
}

// RemoveSandbox 移除沙盒
func (sm *SandboxManager) RemoveSandbox(id string) error {
	sandbox, exists := sm.sandboxes[id]
	if !exists {
		return fmt.Errorf("沙盒不存在: %s", id)
	}

	delete(sm.sandboxes, id)
	return sandbox.Remove(context.Background())
}

// Cleanup 清理所有沙盒
func (sm *SandboxManager) Cleanup() error {
	for id, sandbox := range sm.sandboxes {
		if err := sandbox.Remove(context.Background()); err != nil {
			logger.Error("移除沙盒失败", 
				zap.String("id", id),
				zap.Error(err))
		}
	}

	sm.sandboxes = make(map[string]Sandbox)
	return nil
}
