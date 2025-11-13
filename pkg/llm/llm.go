package llm

import (
    "context"
    "fmt"
    "strings"

    "github.com/sashabaranov/go-openai"
    "github.com/yahao333/GoManus/pkg/config"
    "github.com/yahao333/GoManus/pkg/logger"
    "github.com/yahao333/GoManus/pkg/schema"
    "go.uber.org/zap"
)

// Provider LLM提供者接口
type Provider interface {
	GenerateResponse(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (*schema.Message, error)
	GenerateStreamResponse(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (<-chan string, error)
}

// LLM LLM客户端
type LLM struct {
	provider   Provider
	configName string
}

// NewLLM 创建新的LLM客户端
func NewLLM(configName string) (*LLM, error) {
	settings, ok := config.GetConfig().GetLLMSettings(configName)
	if !ok {
		settings = config.GetConfig().GetDefaultLLMSettings()
	}

	var provider Provider
	var err error

	switch strings.ToLower(settings.APIType) {
	case "openai":
		provider, err = NewOpenAIProvider(settings)
	case "azure":
		provider, err = NewAzureProvider(settings)
	case "ollama":
		provider, err = NewOllamaProvider(settings)
	default:
		return nil, fmt.Errorf("不支持的API类型: %s", settings.APIType)
	}

	if err != nil {
		return nil, err
	}

	return &LLM{
		provider:   provider,
		configName: configName,
	}, nil
}

// GenerateResponse 生成响应
func (l *LLM) GenerateResponse(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (*schema.Message, error) {
	return l.provider.GenerateResponse(ctx, messages, tools)
}

// GenerateStreamResponse 生成流式响应
func (l *LLM) GenerateStreamResponse(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (<-chan string, error) {
	return l.provider.GenerateStreamResponse(ctx, messages, tools)
}

// OpenAIProvider OpenAI提供者
type OpenAIProvider struct {
	client *openai.Client
	config config.LLMSettings
}

// NewOpenAIProvider 创建OpenAI提供者
func NewOpenAIProvider(settings config.LLMSettings) (*OpenAIProvider, error) {
	config := openai.DefaultConfig(settings.APIKey)
	if settings.BaseURL != "" {
		config.BaseURL = settings.BaseURL
	}

	client := openai.NewClientWithConfig(config)
	return &OpenAIProvider{
		client: client,
		config: settings,
	}, nil
}

// GenerateResponse 生成响应
func (o *OpenAIProvider) GenerateResponse(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (*schema.Message, error) {
	openaiMessages := o.convertMessages(messages)
	openaiTools := o.convertTools(tools)

	req := openai.ChatCompletionRequest{
		Model:       o.config.Model,
		Messages:    openaiMessages,
		MaxTokens:   o.config.MaxTokens,
		Temperature: float32(o.config.Temperature),
	}

	if len(openaiTools) > 0 {
		req.Tools = openaiTools
	}

	resp, err := o.client.CreateChatCompletion(ctx, req)
	if err != nil {
		logger.Error("OpenAI API调用失败", zap.Error(err))
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("没有收到响应")
	}

	choice := resp.Choices[0]
	content := choice.Message.Content

	// 转换工具调用
	var toolCalls []schema.ToolCall
	if choice.Message.ToolCalls != nil {
		toolCalls = make([]schema.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			toolCalls[i] = schema.ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: schema.Function{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	return &schema.Message{
		Role:      schema.RoleAssistant,
		Content:   &content,
		ToolCalls: toolCalls,
	}, nil
}

// GenerateStreamResponse 生成流式响应
func (o *OpenAIProvider) GenerateStreamResponse(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (<-chan string, error) {
	openaiMessages := o.convertMessages(messages)
	openaiTools := o.convertTools(tools)

	req := openai.ChatCompletionRequest{
		Model:       o.config.Model,
		Messages:    openaiMessages,
		MaxTokens:   o.config.MaxTokens,
		Temperature: float32(o.config.Temperature),
		Stream:      true,
	}

	if len(openaiTools) > 0 {
		req.Tools = openaiTools
	}

	stream, err := o.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	resultChan := make(chan string, 100)

	go func() {
		defer close(resultChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				if err.Error() != "EOF" {
					logger.Error("流式响应接收失败", zap.Error(err))
				}
				return
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					resultChan <- content
				}
			}
		}
	}()

	return resultChan, nil
}

// convertMessages 转换消息格式
func (o *OpenAIProvider) convertMessages(messages []schema.Message) []openai.ChatCompletionMessage {
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role: string(msg.Role),
		}
		
		if msg.Content != nil {
			openaiMsg.Content = *msg.Content
		}
		
		if msg.Name != nil {
			openaiMsg.Name = *msg.Name
		}
		
		if msg.ToolCallID != nil {
			openaiMsg.ToolCallID = *msg.ToolCallID
		}
		
		// 转换工具调用
		if msg.ToolCalls != nil {
			openaiMsg.ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				openaiMsg.ToolCalls[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}
		
		openaiMessages[i] = openaiMsg
	}
	return openaiMessages
}

// convertTools 转换工具定义
func (o *OpenAIProvider) convertTools(tools []schema.ToolDefinition) []openai.Tool {
	if len(tools) == 0 {
		return nil
	}

	openaiTools := make([]openai.Tool, len(tools))
	for i, tool := range tools {
		// 构建符合OpenAI API规范的参数schema
		params := map[string]interface{}{
			"type":       "object",
			"properties": tool.Parameters,
		}

		// 添加必需参数字段
		if len(tool.Required) > 0 {
			params["required"] = tool.Required
		}

		openaiTools[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  params,
			},
		}
	}
	return openaiTools
}

// AzureProvider Azure OpenAI提供者
type AzureProvider struct {
	*OpenAIProvider
}

// NewAzureProvider 创建Azure提供者
func NewAzureProvider(settings config.LLMSettings) (*AzureProvider, error) {
	config := openai.DefaultAzureConfig(settings.APIKey, settings.BaseURL)
	if settings.APIVersion != "" {
		config.APIVersion = settings.APIVersion
	}

	client := openai.NewClientWithConfig(config)
	return &AzureProvider{
		OpenAIProvider: &OpenAIProvider{
			client: client,
			config: settings,
		},
	}, nil
}

// OllamaProvider Ollama提供者
type OllamaProvider struct {
	baseURL string
	model   string
}

// NewOllamaProvider 创建Ollama提供者
func NewOllamaProvider(settings config.LLMSettings) (*OllamaProvider, error) {
	return &OllamaProvider{
		baseURL: settings.BaseURL,
		model:   settings.Model,
	}, nil
}

// GenerateResponse 生成响应（简化实现）
func (o *OllamaProvider) GenerateResponse(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (*schema.Message, error) {
	// 这里应该实现Ollama API调用
	// 为了简化，返回一个默认消息
	content := "Ollama响应（未实现）"
	return &schema.Message{
		Role:    schema.RoleAssistant,
		Content: &content,
	}, nil
}

// GenerateStreamResponse 生成流式响应
func (o *OllamaProvider) GenerateStreamResponse(ctx context.Context, messages []schema.Message, tools []schema.ToolDefinition) (<-chan string, error) {
	resultChan := make(chan string, 1)
	go func() {
		defer close(resultChan)
		resultChan <- "Ollama流式响应（未实现）"
	}()
	return resultChan, nil
}
