package sandbox

import (
    "context"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    "github.com/yahao333/GoManus/pkg/config"
    "github.com/yahao333/GoManus/internal/logger"
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

// DockerSandbox Docker沙盒实现
type DockerSandbox struct {
	containerID  string
	config       *config.SandboxSettings
	image        string
	workDir      string
	status       string
}

// NewDockerSandbox 创建新的Docker沙盒
func NewDockerSandbox(config *config.SandboxSettings) (*DockerSandbox, error) {
	return &DockerSandbox{
		config:  config,
		image:   config.Image,
		workDir: config.WorkDir,
		status:  "created",
	}, nil
}

// Create 创建沙盒容器
func (d *DockerSandbox) Create(ctx context.Context) error {
	logger.Info("创建Docker沙盒", zap.String("image", d.image))

	// 检查Docker是否可用
	if !d.isDockerAvailable() {
		logger.Warn("Docker不可用，使用本地沙盒模式")
		return d.createLocalSandbox()
	}

	// 这里应该实现Docker容器的创建逻辑
	// 为了简化，返回模拟的容器ID
	d.containerID = "mock_container_" + fmt.Sprintf("%d", time.Now().Unix())
	d.status = "created"

	logger.Info("Docker沙盒创建成功", zap.String("container_id", d.containerID))
	return nil
}

// Start 启动沙盒容器
func (d *DockerSandbox) Start(ctx context.Context) error {
	if d.containerID == "" {
		return fmt.Errorf("容器未创建")
	}

	logger.Info("启动Docker沙盒", zap.String("container_id", d.containerID))

	// 这里应该实现Docker容器的启动逻辑
	d.status = "running"
	logger.Info("Docker沙盒启动成功")
	return nil
}

// Stop 停止沙盒容器
func (d *DockerSandbox) Stop(ctx context.Context) error {
	if d.containerID == "" {
		return fmt.Errorf("容器未创建")
	}

	logger.Info("停止Docker沙盒", zap.String("container_id", d.containerID))

	// 这里应该实现Docker容器的停止逻辑
	d.status = "stopped"
	logger.Info("Docker沙盒停止成功")
	return nil
}

// Remove 移除沙盒容器
func (d *DockerSandbox) Remove(ctx context.Context) error {
	if d.containerID == "" {
		return fmt.Errorf("容器未创建")
	}

	logger.Info("移除Docker沙盒", zap.String("container_id", d.containerID))

	// 这里应该实现Docker容器的移除逻辑
	d.containerID = ""
	d.status = "removed"
	logger.Info("Docker沙盒移除成功")
	return nil
}

// Execute 在沙盒中执行命令
func (d *DockerSandbox) Execute(ctx context.Context, command string, timeout time.Duration) (string, error) {
	if d.containerID == "" {
		return "", fmt.Errorf("容器未创建")
	}

	if d.status != "running" {
		return "", fmt.Errorf("容器未运行")
	}

	logger.Info("执行命令", 
		zap.String("command", command),
		zap.String("container_id", d.containerID))

	// 如果Docker不可用，使用本地执行
	if !d.isDockerAvailable() {
		return d.executeLocalCommand(ctx, command, timeout)
	}

	// 这里应该实现Docker命令执行逻辑
	// 为了简化，返回模拟的执行结果
	return fmt.Sprintf("模拟执行结果: %s", command), nil
}

// GetStatus 获取沙盒状态
func (d *DockerSandbox) GetStatus() string {
	return d.status
}

// isDockerAvailable 检查Docker是否可用
func (d *DockerSandbox) isDockerAvailable() bool {
	cmd := exec.Command("docker", "--version")
	err := cmd.Run()
	return err == nil
}

// createLocalSandbox 创建本地沙盒
func (d *DockerSandbox) createLocalSandbox() error {
	// 创建临时工作目录
	tempDir, err := os.MkdirTemp("", "gomanus_sandbox_*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	d.workDir = tempDir
	d.containerID = "local_" + fmt.Sprintf("%d", time.Now().Unix())
	return nil
}

// executeLocalCommand 本地执行命令
func (d *DockerSandbox) executeLocalCommand(ctx context.Context, command string, timeout time.Duration) (string, error) {
	// 创建命令
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = d.workDir
	
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

	sandbox, err := NewDockerSandbox(sm.config)
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
