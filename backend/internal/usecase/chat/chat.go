package chatusecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/agent"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
)

const (
	defaultHistoryLimit  = 50
	defaultMaxToolRounds = 4
)

type Options struct {
	HistoryLimit  int
	MaxToolRounds int
}

type Service struct {
	chats         chatdomain.ChatRepository
	messages      chatdomain.MessageRepository
	incidents     chatdomain.IncidentRepository
	scenarios     chatdomain.ScenarioProvider
	agent         agent.Agent
	historyLimit  int
	maxToolRounds int
}

func New(
	chats chatdomain.ChatRepository,
	messages chatdomain.MessageRepository,
	incidents chatdomain.IncidentRepository,
	scenarios chatdomain.ScenarioProvider,
	chatAgent agent.Agent,
	options Options,
) *Service {
	if options.HistoryLimit <= 0 {
		options.HistoryLimit = defaultHistoryLimit
	}
	if options.MaxToolRounds <= 0 {
		options.MaxToolRounds = defaultMaxToolRounds
	}

	return &Service{
		chats:         chats,
		messages:      messages,
		incidents:     incidents,
		scenarios:     scenarios,
		agent:         chatAgent,
		historyLimit:  options.HistoryLimit,
		maxToolRounds: options.MaxToolRounds,
	}
}

func (s *Service) SendMessage(ctx context.Context, chatID, userID int64, content string) (*chatdomain.Message, error) {
	if strings.TrimSpace(content) == "" {
		return nil, chatdomain.ErrMessageEmpty
	}

	chat, err := s.chats.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if chat.UserID != userID {
		return nil, chatdomain.ErrChatAccessDenied
	}

	if !chat.IsActive() {
		return nil, chatdomain.ErrChatFinished
	}

	prompt, err := s.scenarios.SystemPrompt(ctx, chat.ScenarioID)
	if err != nil {
		return nil, err
	}

	userMessage := &chatdomain.Message{
		ChatID:     chat.ID,
		SenderType: chatdomain.SenderTypeUser,
		Content:    content,
	}
	if err := s.messages.Create(ctx, userMessage); err != nil {
		return nil, err
	}

	history, err := s.messages.ListByChatID(ctx, chat.ID, chatdomain.Cursor{Limit: s.historyLimit})
	if err != nil {
		return nil, err
	}

	reply, err := s.runAgent(ctx, chat, buildMessages(prompt, history))
	if err != nil {
		return nil, err
	}

	agentMessage := &chatdomain.Message{
		ChatID:     chat.ID,
		SenderType: chatdomain.SenderTypeAgent,
		Content:    reply,
	}
	if err := s.messages.Create(ctx, agentMessage); err != nil {
		return nil, err
	}

	return agentMessage, nil
}

func (s *Service) GetChat(ctx context.Context, chatID int64) (*chatdomain.Chat, error) {
	return s.chats.GetByID(ctx, chatID)
}

func (s *Service) ListMessages(ctx context.Context, chatID int64, cursor chatdomain.Cursor) ([]*chatdomain.Message, error) {
	if _, err := s.chats.GetByID(ctx, chatID); err != nil {
		return nil, err
	}

	return s.messages.ListByChatID(ctx, chatID, cursor)
}

func (s *Service) runAgent(ctx context.Context, chat *chatdomain.Chat, messages []agent.Message) (string, error) {
	content := ""

	for round := 0; round <= s.maxToolRounds; round++ {
		reply, err := s.agent.Complete(ctx, agent.Request{Messages: messages, Tools: toolDefinitions})
		if err != nil {
			return "", err
		}

		if reply.Content != "" {
			content = reply.Content
		}

		if len(reply.ToolCalls) == 0 || round == s.maxToolRounds {
			break
		}

		messages = append(messages, agent.Message{
			Role:      agent.RoleAssistant,
			Content:   reply.Content,
			ToolCalls: reply.ToolCalls,
		})

		for _, call := range reply.ToolCalls {
			result, err := s.executeTool(ctx, chat, call)
			if err != nil {
				return "", err
			}

			messages = append(messages, agent.Message{
				Role:       agent.RoleTool,
				Content:    result,
				ToolCallID: call.ID,
			})
		}
	}

	if content == "" {
		return "", fmt.Errorf("%w: agent produced no content", agent.ErrBadResponse)
	}

	return content, nil
}

func (s *Service) executeTool(ctx context.Context, chat *chatdomain.Chat, call agent.ToolCall) (string, error) {
	switch call.Name {
	case toolReportIncident:
		return s.reportIncident(ctx, chat, call.Arguments)
	case toolFinishChat:
		return s.finishChat(ctx, chat, call.Arguments)
	default:
		return toolFailure(agent.ErrToolCallInvalid), nil
	}
}

func (s *Service) reportIncident(ctx context.Context, chat *chatdomain.Chat, arguments []byte) (string, error) {
	var payload struct {
		Type    string `json:"type"`
		Comment string `json:"comment"`
	}
	if err := json.Unmarshal(arguments, &payload); err != nil {
		return toolFailure(agent.ErrToolCallInvalid), nil
	}

	incidentType := chatdomain.IncidentType(payload.Type)
	if !incidentType.Valid() {
		return toolFailure(chatdomain.ErrInvalidIncidentKind), nil
	}

	incident := &chatdomain.Incident{
		ChatID:  chat.ID,
		Type:    incidentType,
		Comment: payload.Comment,
	}

	if err := chat.ApplyIncident(incident); err != nil {
		return toolFailure(err), nil
	}

	if err := s.incidents.Create(ctx, incident, chat.Score); err != nil {
		return "", err
	}

	return fmt.Sprintf("incident %s recorded, current score %d", incident.Type, chat.Score), nil
}

func (s *Service) finishChat(ctx context.Context, chat *chatdomain.Chat, arguments []byte) (string, error) {
	var payload struct {
		Resume string `json:"resume"`
	}
	if err := json.Unmarshal(arguments, &payload); err != nil {
		return toolFailure(agent.ErrToolCallInvalid), nil
	}

	if err := chat.Finish(payload.Resume); err != nil {
		return toolFailure(err), nil
	}

	if err := s.chats.Update(ctx, chat); err != nil {
		return "", err
	}

	return fmt.Sprintf("chat finished with score %d", chat.Score), nil
}

func buildMessages(prompt string, history []*chatdomain.Message) []agent.Message {
	messages := make([]agent.Message, 0, len(history)+1)
	messages = append(messages, agent.Message{
		Role:    agent.RoleSystem,
		Content: prompt,
	})

	for i := len(history) - 1; i >= 0; i-- {
		messages = append(messages, agent.Message{
			Role:    agentRole(history[i].SenderType),
			Content: history[i].Content,
		})
	}

	return messages
}

func agentRole(sender chatdomain.SenderType) agent.Role {
	if sender == chatdomain.SenderTypeAgent {
		return agent.RoleAssistant
	}

	return agent.RoleUser
}

func toolFailure(err error) string {
	return fmt.Sprintf("error: %s", err)
}
