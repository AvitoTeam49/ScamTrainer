package deepseek

import (
	"encoding/json"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/agent"
)

const toolTypeFunction = "function"

type functionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function functionCall `json:"function"`
}

type functionDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type tool struct {
	Type     string             `json:"type"`
	Function functionDefinition `json:"function"`
}

type message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type request struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Tools    []tool    `json:"tools,omitempty"`
	Stream   bool      `json:"stream"`
}

type choice struct {
	Message      message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type response struct {
	Choices []choice `json:"choices"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func toAPIMessages(messages []agent.Message) []message {
	converted := make([]message, 0, len(messages))
	for _, source := range messages {
		converted = append(converted, message{
			Role:       string(source.Role),
			Content:    source.Content,
			ToolCalls:  toAPIToolCalls(source.ToolCalls),
			ToolCallID: source.ToolCallID,
		})
	}

	return converted
}

func toAPIToolCalls(calls []agent.ToolCall) []toolCall {
	if len(calls) == 0 {
		return nil
	}

	converted := make([]toolCall, 0, len(calls))
	for _, call := range calls {
		converted = append(converted, toolCall{
			ID:   call.ID,
			Type: toolTypeFunction,
			Function: functionCall{
				Name:      call.Name,
				Arguments: string(call.Arguments),
			},
		})
	}

	return converted
}

func toAPITools(tools []agent.ToolDefinition) []tool {
	if len(tools) == 0 {
		return nil
	}

	converted := make([]tool, 0, len(tools))
	for _, source := range tools {
		converted = append(converted, tool{
			Type: toolTypeFunction,
			Function: functionDefinition{
				Name:        source.Name,
				Description: source.Description,
				Parameters:  source.Parameters,
			},
		})
	}

	return converted
}

func replyFromMessage(source message) *agent.Reply {
	reply := &agent.Reply{Content: source.Content}
	if len(source.ToolCalls) == 0 {
		return reply
	}

	reply.ToolCalls = make([]agent.ToolCall, 0, len(source.ToolCalls))
	for _, call := range source.ToolCalls {
		reply.ToolCalls = append(reply.ToolCalls, agent.ToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: []byte(call.Function.Arguments),
		})
	}

	return reply
}
