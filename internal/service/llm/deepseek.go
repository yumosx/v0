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

func NewHandler(client *deepseek.Client, session Session) *Handler {
	return &Handler{client: client}
}

func (h *Handler) Invoke(ctx context.Context, request domain.LLMRequest) (domain.LLMResponse, error) {
	for _, msg := range request.Msgs {
		switch msg.Role {
		case "system":
			h.session.AppendMsg(deepseek.ChatCompletionMessage{Role: deepseek.ChatMessageRoleSystem, Content: msg.Content})
		case "user":
			h.session.AppendMsg(deepseek.ChatCompletionMessage{Role: deepseek.ChatMessageRoleSystem, Content: msg.Content})
		case "assistant":
			h.session.AppendMsg(deepseek.ChatCompletionMessage{Role: deepseek.ChatMessageRoleAssistant, Content: msg.Content})
		case "tool":
			h.session.AppendMsg(deepseek.ChatCompletionMessage{Role: deepseek.ChatMessageRoleTool, Content: msg.Content, ToolCallID: msg.Id})
		default:
			continue
		}
	}
	completion, err := h.client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
		Model:    deepseek.DeepSeekChat,
		Messages: h.session.messages,
	})

	if err != nil {
		return domain.LLMResponse{}, err
	}

	message, err := deepseek.MapMessageToChatCompletionMessage(completion.Choices[0].Message)
	if err != nil {
		return domain.LLMResponse{}, err
	}
	h.session.AppendMsg(message)

	return h.msgToDomain(completion), nil
}

func (h *Handler) msgToDomain(msg *deepseek.ChatCompletionResponse) domain.LLMResponse {
	var resp domain.LLMResponse
	resp.Content = msg.Choices[0].Message.Content
	resp.ToolCalls = stream.Map[deepseek.ToolCall, domain.LLMToolCall](msg.Choices[0].Message.ToolCalls, func(idx int, src deepseek.ToolCall) domain.LLMToolCall {
		return domain.LLMToolCall{
			ID:    src.ID,
			Index: src.Index,
			Type:  src.Type,
			Function: domain.LLMToolCallFunction{
				Name:      src.Function.Name,
				Arguments: src.Function.Arguments,
			},
		}
	})

	return resp
}

func (h *Handler) domainToToolCalls(toolCalls []domain.LLMToolCall) {
	stream.Map[domain.LLMToolCall, deepseek.ToolCall](toolCalls, func(idx int, src domain.LLMToolCall) deepseek.ToolCall {
		return deepseek.ToolCall{
			ID:       src.ID,
			Function: deepseek.ToolCallFunction{},
		}
	})
}
