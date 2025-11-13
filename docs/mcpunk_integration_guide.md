# MCPunk 集成使用指南

## 概述

MCPunk是一个基于MCP协议的代码分析工具，它提供了多个实用的工具来帮助理解和分析代码库。通过GoManus的MCP客户端集成，你可以直接在对话中使用这些工具。

## MCPunk提供的工具

### 1. get_a_joke
- **描述**: 获取一个编程相关的笑话
- **参数**: 无
- **用途**: 轻松一下，获取编程相关的幽默内容

### 2. configure_project
- **描述**: 配置项目路径，让MCPunk分析指定目录
- **参数**: 
  - `project_path`: 项目目录路径
- **用途**: 设置要分析的代码库路径

### 3. list_all_files_in_project
- **描述**: 列出项目中所有文件
- **参数**: 无（需要先配置项目）
- **用途**: 获取项目文件结构概览

### 4. find_files_by_chunk_content
- **描述**: 根据内容搜索文件块
- **参数**:
  - `query`: 搜索关键词
  - `file_path`: 可选，指定文件路径
- **用途**: 智能搜索代码内容

## 配置方法

### 基本配置

在你的 `config.toml` 文件中添加以下配置：

```toml
[mcp]
enabled = true

[mcp.servers.mcpunk]
name = "mcpunk"
type = "stdio"
command = "uvx"
args = ["mcpunk"]
```

### 高级配置

如果需要指定完整路径或环境变量：

```toml
[mcp.servers.mcpunk]
name = "mcpunk"
type = "stdio"
command = "/Users/yourname/.local/bin/uvx"  # 完整路径
args = ["mcpunk"]

[mcp.servers.mcpunk.env]
MCPUNK_INCLUDE_CHARS_IN_RESPONSE = "false"
```

## 使用示例

### 1. 获取笑话

```bash
用户: 给我讲个编程笑话
助手: 我来为您获取一个编程相关的笑话。

[调用get_a_joke工具]

助手: 这是一个编程笑话：
"为什么程序员总是搞混万圣节和圣诞节？
因为 Oct 31 == Dec 25"
```

### 2. 分析代码库

```bash
用户: 请分析我的项目结构
助手: 我来帮您分析项目结构。首先配置项目路径。

[调用configure_project工具，参数: project_path="/path/to/your/project"]

[调用list_all_files_in_project工具]

助手: 您的项目包含以下主要文件：
- main.go: 主程序入口
- internal/: 内部包目录
- config/: 配置文件目录
- ...
```

### 3. 搜索特定代码

```bash
用户: 在我的项目中搜索数据库连接相关的代码
助手: 我来帮您搜索数据库连接相关的代码。

[调用find_files_by_chunk_content工具，参数: query="database connection"]

助手: 找到以下相关代码块：
1. 在db.go文件中，第45-60行：
   [显示具体的数据库连接代码]
2. 在config.go文件中，第120-135行：
   [显示数据库配置代码]
```

## 故障排除

### 常见问题

1. **uvx命令未找到**
   - 确保已安装uv包管理器
   - 使用完整路径：`command = "/Users/username/.local/bin/uvx"`

2. **MCPunk启动失败**
   - 检查Python环境是否正常
   - 尝试手动运行：`uvx mcpunk` 看是否有错误信息

3. **工具调用超时**
   - 增加超时时间配置
   - 检查网络连接（如果需要下载依赖）

### 验证安装

运行以下命令验证MCPunk是否正确安装：

```bash
uvx mcpunk --help
```

## 集成开发

在GoManus中使用MCPunk工具的示例代码：

```go
// 在agent对话中调用MCPunk工具
func (a *Agent) handleUserRequest(ctx context.Context, request string) error {
    // 判断用户意图
    if strings.Contains(request, "笑话") || strings.Contains(request, "joke") {
        // 调用get_a_joke工具
        result, err := a.mcpClient.CallTool(ctx, "mcpunk", "get_a_joke", nil)
        if err != nil {
            return err
        }
        a.SendMessage(result)
    }
    
    // 其他工具调用逻辑...
    return nil
}
```

## 最佳实践

1. **项目配置**: 在分析代码前先配置正确的项目路径
2. **搜索策略**: 使用具体的搜索词，避免过于宽泛的关键词
3. **结果验证**: 检查结果的相关性，必要时调整搜索参数
4. **错误处理**: 妥善处理工具调用失败的情况

## 相关链接

- [MCPunk GitHub仓库](https://github.com/jurasofish/mcpunk)
- [MCP协议文档](https://modelcontextprotocol.io/)
- [uv包管理器](https://docs.astral.sh/uv/)