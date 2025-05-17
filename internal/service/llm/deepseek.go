package llm

import (
	"context"

	"github.com/cohesion-org/deepseek-go"
	"github.com/yumosx/got/pkg/stream"
	"vzero/internal/domain"
)

type Handler struct {
	client  *deepseek.Client
	session *Session
}

func NewHandler(client *deepseek.Client, session *Session) *Handler {
	return &Handler{client: client, session: session}
}

func (h *Handler) Invoke(ctx context.Context, request domain.LLMRequest) (domain.LLMResponse, error) {
	for _, msg := range request.Msgs {
		switch msg.Role {
		case "system":
			h.session.AppendMsg(deepseek.ChatCompletionMessage{Role: deepseek.ChatMessageRoleSystem, Content: msg.Content})
		case "user":
			h.session.AppendMsg(deepseek.ChatCompletionMessage{Role: deepseek.ChatMessageRoleUser, Content: msg.Content})
		case "assistant":
			h.session.AppendMsg(deepseek.ChatCompletionMessage{Role: deepseek.ChatMessageRoleAssistant, Content: msg.Content, ToolCalls: h.domainToToolCalls(msg.ToolCalls)})
		case "tool":
			h.session.AppendMsg(deepseek.ChatCompletionMessage{Role: deepseek.ChatMessageRoleTool, Content: msg.Content, ToolCallID: msg.Id})
		default:
			continue
		}
	}

	var tools []deepseek.Tool
	if len(request.Tools) != 0 {
		tools = h.domainToTools(request.Tools)
	}

	completion, err := h.client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
		Model:    deepseek.DeepSeekChat,
		Messages: h.session.messages,
		Tools:    tools,
	})

	if err != nil {
		return domain.LLMResponse{}, err
	}

	return h.msgToDomain(completion), nil
}

func (h *Handler) msgToDomain(msg *deepseek.ChatCompletionResponse) domain.LLMResponse {
	var resp domain.LLMResponse
	resp.Content = msg.Choices[0].Message.Content
	resp.ToolCalls = make([]domain.LLMToolCall, len(msg.Choices[0].Message.ToolCalls))
	for i, tool := range msg.Choices[0].Message.ToolCalls {
		resp.ToolCalls[i] = domain.LLMToolCall{
			ID:   tool.ID,
			Type: tool.Type,
			Function: domain.LLMToolCallFunction{
				Name:      tool.Function.Name,
				Arguments: tool.Function.Arguments,
			},
		}
	}

	return resp
}

func (h *Handler) domainToToolCalls(toolCalls []domain.LLMToolCall) []deepseek.ToolCall {
	return stream.Map[domain.LLMToolCall, deepseek.ToolCall](toolCalls, func(idx int, src domain.LLMToolCall) deepseek.ToolCall {
		return deepseek.ToolCall{
			ID:   src.ID,
			Type: src.Type,
			Function: deepseek.ToolCallFunction{
				Name:      src.Function.Name,
				Arguments: src.Function.Arguments,
			},
		}
	})
}

func (h *Handler) domainToTools(tools []domain.Tool) []deepseek.Tool {
	return stream.Map[domain.Tool, deepseek.Tool](tools, func(idx int, src domain.Tool) deepseek.Tool {
		return deepseek.Tool{
			Type: src.Type,
			Function: deepseek.Function{
				Name:        src.Function.Name,
				Description: src.Function.Description,
				Parameters: &deepseek.FunctionParameters{
					Type:       src.Function.Parameters.Type,
					Properties: src.Function.Parameters.Properties.ToMap(),
					Required:   src.Function.Parameters.Required,
				},
			},
		}
	})
}
