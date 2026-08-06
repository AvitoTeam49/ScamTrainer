package agent

import (
	"context"
	"errors"
)

var (
	ErrUnavailable     = errors.New("agent is unavailable")
	ErrRateLimited     = errors.New("agent rate limit exceeded")
	ErrBadResponse     = errors.New("agent returned unexpected response")
	ErrToolCallInvalid = errors.New("invalid tool call arguments")
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type ToolCall struct {
	ID        string
	Name      string
	Arguments []byte
}

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  []byte
}

type Message struct {
	Role       Role
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
}

type Request struct {
	Messages []Message
	Tools    []ToolDefinition
}

type Reply struct {
	Content   string
	ToolCalls []ToolCall
}

type Agent interface {
	Complete(ctx context.Context, request Request) (*Reply, error)
}
