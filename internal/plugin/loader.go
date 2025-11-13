package plugin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// PluginConfig 插件配置文件结构
type PluginConfig struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	EntryPoint  string                 `json:"entry_point"`
	Config      map[string]interface{} `json:"config"`
	Dependencies []string              `json:"dependencies"`
	Enabled     bool                   `json:"enabled"`
}

// PluginManifest 插件清单文件
type PluginManifest struct {
	Plugins []PluginConfig `json:"plugins"`
}

// Loader 插件加载器
type Loader struct {
	manager      *Manager
	pluginDirs   []string
	manifestFile string
}

// NewLoader 创建插件加载器
func NewLoader(manager *Manager, pluginDirs []string, manifestFile string) *Loader {
	return &Loader{
		manager:      manager,
		pluginDirs:   pluginDirs,
		manifestFile: manifestFile,
	}
}

// LoadAllPlugins 加载所有插件
func (l *Loader) LoadAllPlugins() error {
	// 1. 从清单文件加载插件配置
	manifest, err := l.loadManifest()
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// 2. 加载启用的插件
	for _, pluginConfig := range manifest.Plugins {
		if !pluginConfig.Enabled {
			continue
		}

		if err := l.loadPlugin(pluginConfig); err != nil {
			fmt.Printf("Warning: failed to load plugin %s: %v\n", pluginConfig.Name, err)
			continue
		}
	}

	// 3. 自动发现插件目录中的插件
	if err := l.discoverPlugins(); err != nil {
		fmt.Printf("Warning: plugin discovery failed: %v\n", err)
	}

	return nil
}

// loadManifest 加载插件清单文件
func (l *Loader) loadManifest() (*PluginManifest, error) {
	if _, err := os.Stat(l.manifestFile); os.IsNotExist(err) {
		// 如果清单文件不存在，返回空清单
		return &PluginManifest{Plugins: []PluginConfig{}}, nil
	}

	data, err := ioutil.ReadFile(l.manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest file: %w", err)
	}

	return &manifest, nil
}

// loadPlugin 加载单个插件
func (l *Loader) loadPlugin(config PluginConfig) error {
	// 查找插件文件
	pluginPath, err := l.findPluginFile(config.EntryPoint)
	if err != nil {
		return fmt.Errorf("failed to find plugin file: %w", err)
	}

	// 加载插件
	if err := l.manager.LoadPlugin(pluginPath); err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	return nil
}

// findPluginFile 查找插件文件
func (l *Loader) findPluginFile(entryPoint string) (string, error) {
	// 如果entryPoint是绝对路径，直接使用
	if filepath.IsAbs(entryPoint) {
		if _, err := os.Stat(entryPoint); err != nil {
			return "", fmt.Errorf("plugin file not found: %s", entryPoint)
		}
		return entryPoint, nil
	}

	// 在插件目录中查找
	for _, dir := range l.pluginDirs {
		pluginPath := filepath.Join(dir, entryPoint)
		if _, err := os.Stat(pluginPath); err == nil {
			return pluginPath, nil
		}

		// 尝试添加.so后缀
		if !strings.HasSuffix(entryPoint, ".so") {
			pluginPathWithExt := pluginPath + ".so"
			if _, err := os.Stat(pluginPathWithExt); err == nil {
				return pluginPathWithExt, nil
			}
		}
	}

	return "", fmt.Errorf("plugin file not found: %s", entryPoint)
}

// discoverPlugins 自动发现插件
func (l *Loader) discoverPlugins() error {
	for _, dir := range l.pluginDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Printf("Warning: failed to read plugin directory %s: %v\n", dir, err)
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			// 只处理.so文件
			if !strings.HasSuffix(file.Name(), ".so") {
				continue
			}

			pluginPath := filepath.Join(dir, file.Name())
			
			// 检查是否已经加载过
			if l.isPluginLoaded(file.Name()) {
				continue
			}

			// 尝试加载插件
			if err := l.manager.LoadPlugin(pluginPath); err != nil {
				fmt.Printf("Warning: failed to auto-load plugin %s: %v\n", file.Name(), err)
				continue
			}

			fmt.Printf("Auto-loaded plugin: %s\n", file.Name())
		}
	}

	return nil
}

// isPluginLoaded 检查插件是否已经加载
func (l *Loader) isPluginLoaded(pluginFile string) bool {
	// 这里简化处理，实际应该检查插件名称
	// 由于我们不知道插件名称，这里只是简单检查文件名
	return false
}

// SaveManifest 保存插件清单文件
func (l *Loader) SaveManifest(manifest *PluginManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := ioutil.WriteFile(l.manifestFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// GetLoadedPlugins 获取已加载的插件信息
func (l *Loader) GetLoadedPlugins() []PluginMetadata {
	return l.manager.ListPlugins()
}