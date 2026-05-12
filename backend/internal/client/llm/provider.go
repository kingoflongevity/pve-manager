package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChatMessage 统一消息格式
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ToolDefinition 工具定义
type ToolDefinition struct {
	Type     string `json:"type"`
	Function struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Parameters  map[string]interface{} `json:"parameters"`
	} `json:"function"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Usage     UsageInfo  `json:"usage,omitempty"`
}

// UsageInfo Token 使用信息
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk 流式响应块
type StreamChunk struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Done      bool       `json:"done"`
	Error     string     `json:"error,omitempty"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Provider    string
	BaseURL     string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
	Timeout     int
}

// LLMProvider 统一 LLM 接口
type LLMProvider interface {
	Chat(ctx context.Context, messages []ChatMessage, tools []ToolDefinition) (*ChatResponse, error)
	ChatStream(ctx context.Context, messages []ChatMessage, tools []ToolDefinition) (<-chan StreamChunk, error)
	Close() error
}

// NewProvider 创建 LLM 提供商
func NewProvider(cfg ProviderConfig) (LLMProvider, error) {
	return NewOpenAIProvider(cfg)
}

// OpenAIProvider OpenAI 兼容实现
type OpenAIProvider struct {
	baseURL string
	apiKey  string
	model   string
	timeout int
	client  *http.Client
}

// NewOpenAIProvider 创建 OpenAI 兼容客户端
func NewOpenAIProvider(cfg ProviderConfig) (*OpenAIProvider, error) {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if !strings.Contains(baseURL, "/v1") {
		baseURL = baseURL + "/v1"
	}

	return &OpenAIProvider{
		baseURL: baseURL,
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		timeout: cfg.Timeout,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}, nil
}

// Chat 发送聊天请求
func (p *OpenAIProvider) Chat(ctx context.Context, messages []ChatMessage, tools []ToolDefinition) (*ChatResponse, error) {
	reqBody := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"max_tokens":  4096,
		"temperature": 0.7,
		"stream":      false,
	}
	if len(tools) > 0 {
		reqBody["tools"] = tools
	}

	body, err := p.doRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage UsageInfo `json:"usage"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI 响应为空")
	}

	return &ChatResponse{
		Content:   resp.Choices[0].Message.Content,
		ToolCalls: resp.Choices[0].Message.ToolCalls,
		Usage:     resp.Usage,
	}, nil
}

// ChatStream 流式聊天
func (p *OpenAIProvider) ChatStream(ctx context.Context, messages []ChatMessage, tools []ToolDefinition) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 100)

	reqBody := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"max_tokens":  4096,
		"temperature": 0.7,
		"stream":      true,
	}
	if len(tools) > 0 {
		reqBody["tools"] = tools
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("发送请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		close(ch)
		return ch, fmt.Errorf("API 返回错误 (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					ch <- StreamChunk{Error: err.Error(), Done: true}
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- StreamChunk{Done: true}
				return
			}

			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content   string     `json:"content"`
						ToolCalls []ToolCall `json:"tool_calls"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) == 0 {
				continue
			}

			chunk := StreamChunk{
				Content:   streamResp.Choices[0].Delta.Content,
				ToolCalls: streamResp.Choices[0].Delta.ToolCalls,
			}

			if streamResp.Choices[0].FinishReason != nil {
				chunk.Done = true
			}

			ch <- chunk

			if chunk.Done {
				return
			}
		}
	}()

	return ch, nil
}

// Close 关闭连接
func (p *OpenAIProvider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}

func (p *OpenAIProvider) doRequest(ctx context.Context, reqBody interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}
