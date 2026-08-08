package chatusecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/agent"
	chatdomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/training"
)

const (
	defaultHistoryLimit  = 50
	defaultMaxToolRounds = 4
	genericFailureReason = "internal error"
)

type EventPublisher interface {
	Publish(ctx context.Context, event chatdomain.Event)
}

// ScoreAwarder начисляет пользователю дельту очков. Сколько именно начислить
// и когда — решает чат, снаружи про сценарии и чаты ничего не известно.
type ScoreAwarder interface {
	UpdateUserScore(ctx context.Context, userID int64, scoreDelta int) error
}

type Options struct {
	HistoryLimit  int
	MaxToolRounds int
}

type Service struct {
	chats         chatdomain.ChatRepository
	messages      chatdomain.MessageRepository
	decisions     chatdomain.DecisionRepository
	scenarios     scenariodomain.ScenarioRepository
	sessions      training.SessionRepository
	events        EventPublisher
	awards        ScoreAwarder
	training      *training.TrainingService
	agent         agent.Agent
	historyLimit  int
	maxToolRounds int
}

func New(
	chats chatdomain.ChatRepository,
	messages chatdomain.MessageRepository,
	decisions chatdomain.DecisionRepository,
	scenarios scenariodomain.ScenarioRepository,
	sessions training.SessionRepository,
	trainingService *training.TrainingService,
	events EventPublisher,
	awards ScoreAwarder,
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
		decisions:     decisions,
		scenarios:     scenarios,
		sessions:      sessions,
		events:        events,
		awards:        awards,
		training:      trainingService,
		agent:         chatAgent,
		historyLimit:  options.HistoryLimit,
		maxToolRounds: options.MaxToolRounds,
	}
}

func (s *Service) StartChat(ctx context.Context, userID, scenarioID int64, title string) (*chatdomain.Chat, error) {
	scenario, err := s.scenarios.GetById(ctx, int(scenarioID))
	if err != nil {
		return nil, err
	}

	started, err := s.training.Start(ctx, userID, scenario.ID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(title) == "" {
		title = scenario.Title
	}

	chat := chatdomain.NewChat(userID, scenarioID, started.Session.ID, title, started.Session.CurrentNodeID)
	if err := s.chats.Create(ctx, chat); err != nil {
		return nil, err
	}

	if err := s.openingMessage(ctx, chat, started.Node); err != nil {
		return nil, err
	}

	return chat, nil
}

// Первая реплика берётся из самого сценария, поэтому пользователь сразу видит завязку.
func (s *Service) openingMessage(ctx context.Context, chat *chatdomain.Chat, node *scenariodomain.Node) error {
	text := strings.TrimSpace(node.Message.Text)
	if text == "" {
		return nil
	}

	message := &chatdomain.Message{
		ChatID:     chat.ID,
		SenderType: chatdomain.SenderTypeAgent,
		Content:    text,
	}
	if err := s.messages.Create(ctx, message); err != nil {
		return err
	}

	s.events.Publish(ctx, chatdomain.MessageEvent(message))

	return nil
}

func (s *Service) SendMessage(ctx context.Context, chatID, userID int64, content string) (*chatdomain.Message, error) {
	if strings.TrimSpace(content) == "" {
		return nil, chatdomain.ErrMessageEmpty
	}

	chat, err := s.ownedChat(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}

	if !chat.IsActive() {
		return nil, chatdomain.ErrChatFinished
	}

	scenario, err := s.scenarios.GetById(ctx, int(chat.ScenarioID))
	if err != nil {
		return nil, err
	}

	if _, err := currentNode(scenario, chat); err != nil {
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

	s.events.Publish(ctx, chatdomain.MessageEvent(userMessage))

	return userMessage, nil
}

func (s *Service) RunAgentTurn(ctx context.Context, chatID int64) error {
	chat, err := s.chats.GetByID(ctx, chatID)
	if err != nil {
		return s.fail(ctx, chatID, err)
	}

	if !chat.IsActive() {
		return nil
	}

	scenario, err := s.scenarios.GetById(ctx, int(chat.ScenarioID))
	if err != nil {
		return s.fail(ctx, chatID, err)
	}

	node, err := currentNode(scenario, chat)
	if err != nil {
		return s.fail(ctx, chatID, err)
	}

	history, err := s.messages.ListByChatID(ctx, chat.ID, chatdomain.Cursor{Limit: s.historyLimit})
	if err != nil {
		return s.fail(ctx, chatID, err)
	}

	reply, err := s.runAgent(ctx, scenario, chat, node, buildMessages(systemPrompt(scenario, node), history))
	if err != nil {
		return s.fail(ctx, chatID, err)
	}

	agentMessage := &chatdomain.Message{
		ChatID:     chat.ID,
		SenderType: chatdomain.SenderTypeAgent,
		Content:    reply,
	}
	if err := s.messages.Create(ctx, agentMessage); err != nil {
		return s.fail(ctx, chatID, err)
	}

	s.events.Publish(ctx, chatdomain.MessageEvent(agentMessage))

	return nil
}

func (s *Service) fail(ctx context.Context, chatID int64, err error) error {
	s.events.Publish(ctx, chatdomain.ErrorEvent(chatID, failureReason(err)))

	return err
}

func failureReason(err error) string {
	known := []error{
		chatdomain.ErrChatNotFound,
		chatdomain.ErrChatFinished,
		chatdomain.ErrChatAccessDenied,
		chatdomain.ErrScenarioNotFound,
		scenariodomain.ErrScenarioNotFound,
		scenariodomain.ErrCurrentNodeNotFound,
	}

	for _, sentinel := range known {
		if errors.Is(err, sentinel) {
			return sentinel.Error()
		}
	}

	return genericFailureReason
}

func (s *Service) GetChat(ctx context.Context, chatID, userID int64) (*chatdomain.Chat, error) {
	return s.ownedChat(ctx, chatID, userID)
}

func (s *Service) ListMessages(
	ctx context.Context,
	chatID, userID int64,
	cursor chatdomain.Cursor,
) ([]*chatdomain.Message, error) {
	if _, err := s.ownedChat(ctx, chatID, userID); err != nil {
		return nil, err
	}

	return s.messages.ListByChatID(ctx, chatID, cursor)
}

func (s *Service) ListDecisions(
	ctx context.Context,
	chatID, userID int64,
	cursor chatdomain.Cursor,
) ([]*chatdomain.Decision, error) {
	if _, err := s.ownedChat(ctx, chatID, userID); err != nil {
		return nil, err
	}

	return s.decisions.ListByChatID(ctx, chatID, cursor)
}

func (s *Service) ownedChat(ctx context.Context, chatID, userID int64) (*chatdomain.Chat, error) {
	chat, err := s.chats.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if chat.UserID != userID {
		return nil, chatdomain.ErrChatAccessDenied
	}

	return chat, nil
}

func (s *Service) runAgent(
	ctx context.Context,
	scenario *scenariodomain.Scenario,
	chat *chatdomain.Chat,
	node *scenariodomain.Node,
	messages []agent.Message,
) (string, error) {
	content := ""
	tools := toolDefinitionsFor(node)

	for round := 0; round <= s.maxToolRounds; round++ {
		reply, err := s.agent.Complete(ctx, agent.Request{Messages: messages, Tools: tools})
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
			result, err := s.executeTool(ctx, scenario, chat, call)
			if err != nil {
				return "", err
			}

			messages = append(messages, agent.Message{
				Role:       agent.RoleTool,
				Content:    result,
				ToolCallID: call.ID,
			})
		}

		if !chat.IsActive() {
			break
		}
	}

	if content == "" {
		return "", fmt.Errorf("%w: agent produced no content", agent.ErrBadResponse)
	}

	return content, nil
}

func (s *Service) executeTool(
	ctx context.Context,
	scenario *scenariodomain.Scenario,
	chat *chatdomain.Chat,
	call agent.ToolCall,
) (string, error) {
	switch call.Name {
	case toolApplyTransition:
		return s.applyTransition(ctx, scenario, chat, call.Arguments)
	case toolFinishChat:
		return s.finishChat(ctx, chat, call.Arguments)
	default:
		return toolFailure(agent.ErrToolCallInvalid), nil
	}
}

func (s *Service) applyTransition(
	ctx context.Context,
	scenario *scenariodomain.Scenario,
	chat *chatdomain.Chat,
	arguments []byte,
) (string, error) {
	var payload struct {
		TransitionID string `json:"transition_id"`
	}
	if err := json.Unmarshal(arguments, &payload); err != nil {
		return toolFailure(agent.ErrToolCallInvalid), nil
	}

	if err := s.restoreSession(ctx, chat); err != nil {
		return "", err
	}

	turn, err := s.training.ApplyChoice(ctx, chat.SessionID, payload.TransitionID)
	if err != nil {
		if isChoiceRejected(err) {
			return toolFailure(err), nil
		}

		return "", err
	}

	decision := turn.Decision

	if err := chat.ApplyDecision(decision.ScoreDelta, decision.TargetNodeID); err != nil {
		return toolFailure(err), nil
	}

	stored := &chatdomain.Decision{
		ChatID:       chat.ID,
		NodeID:       decision.NodeID,
		TransitionID: decision.TransitionID,
		TargetNodeID: decision.TargetNodeID,
		ScoreDelta:   decision.ScoreDelta,
		Feedback:     decision.Feedback,
	}
	if err := s.decisions.Create(ctx, stored, chat.Score); err != nil {
		return "", err
	}

	s.events.Publish(ctx, chatdomain.DecisionEvent(stored))

	if !turn.Completed {
		s.events.Publish(ctx, chatdomain.ChatEvent(chat))

		return fmt.Sprintf(
			"transition %s applied, current score %d",
			decision.TransitionID,
			chat.Score,
		), nil
	}

	if err := s.finish(ctx, chat, endingResume(scenario, decision.TargetNodeID, decision.Feedback)); err != nil {
		return "", err
	}

	return fmt.Sprintf("scenario completed with score %d", chat.Score), nil
}

func (s *Service) finishChat(ctx context.Context, chat *chatdomain.Chat, arguments []byte) (string, error) {
	var payload struct {
		Resume string `json:"resume"`
	}
	if err := json.Unmarshal(arguments, &payload); err != nil {
		return toolFailure(agent.ErrToolCallInvalid), nil
	}

	if err := s.finish(ctx, chat, payload.Resume); err != nil {
		return "", err
	}

	return fmt.Sprintf("chat finished with score %d", chat.Score), nil
}

// finish завершает чат ровно один раз: параллельный ход агента получит false и уйдёт без событий.
func (s *Service) finish(ctx context.Context, chat *chatdomain.Chat, resume string) error {
	if err := chat.Finish(resume); err != nil {
		return err
	}

	finished, err := s.chats.Finish(ctx, chat)
	if err != nil {
		return err
	}

	if !finished {
		return chatdomain.ErrChatFinished
	}

	s.award(ctx, chat)

	s.events.Publish(ctx, chatdomain.ChatEvent(chat))

	return nil
}

// Чат уже завершён и сохранён, поэтому неудачное начисление не должно ронять ход агента.
func (s *Service) award(ctx context.Context, chat *chatdomain.Chat) {
	if err := s.awards.UpdateUserScore(ctx, chat.UserID, int(chat.Score)); err != nil {
		slog.ErrorContext(ctx, "failed to award chat score",
			"chat_id", chat.ID, "user_id", chat.UserID, "score", chat.Score, "error", err)
	}
}

// Сессии тренировки живут в памяти, поэтому после рестарта их восстанавливаем из строки чата.
func (s *Service) restoreSession(ctx context.Context, chat *chatdomain.Chat) error {
	_, err := s.sessions.GetById(ctx, chat.SessionID)
	if err == nil {
		return nil
	}

	if !errors.Is(err, training.ErrSessionNotFound) {
		return err
	}

	return s.sessions.Create(ctx, sessionFrom(chat))
}

func isChoiceRejected(err error) bool {
	rejections := []error{
		scenariodomain.ErrSessionCompleted,
		scenariodomain.ErrScenarioMismatch,
		scenariodomain.ErrCurrentNodeNotFound,
		scenariodomain.ErrTransitionNotAvailable,
		scenariodomain.ErrTargetNodeNotFound,
	}

	for _, rejection := range rejections {
		if errors.Is(err, rejection) {
			return true
		}
	}

	return false
}

func sessionFrom(chat *chatdomain.Chat) *scenariodomain.TrainingSession {
	status := scenariodomain.SessionStatusInProgress
	if !chat.IsActive() {
		status = scenariodomain.SessionStatusCompleted
	}

	return &scenariodomain.TrainingSession{
		ID:            chat.SessionID,
		UserID:        chat.UserID,
		ScenarioID:    int(chat.ScenarioID),
		CurrentNodeID: chat.CurrentNodeID,
		Status:        status,
		Score:         int(chat.Score),
		StartedAt:     chat.CreatedAt,
		UpdatedAt:     chat.CreatedAt,
		CompletedAt:   chat.FinishedAt,
	}
}

func currentNode(scenario *scenariodomain.Scenario, chat *chatdomain.Chat) (*scenariodomain.Node, error) {
	node, ok := scenario.Nodes[chat.CurrentNodeID]
	if !ok {
		return nil, fmt.Errorf("%w: node %q", scenariodomain.ErrCurrentNodeNotFound, chat.CurrentNodeID)
	}

	if node.Type == scenariodomain.NodeTypeEnding {
		return nil, chatdomain.ErrChatFinished
	}

	return node, nil
}

func systemPrompt(scenario *scenariodomain.Scenario, node *scenariodomain.Node) string {
	parts := make([]string, 0, 3)

	if prompt := strings.TrimSpace(scenario.LLM.CharacterPrompt); prompt != "" {
		parts = append(parts, prompt)
	}

	if instruction := strings.TrimSpace(node.LLM.ReplyInstruction); instruction != "" {
		parts = append(parts, instruction)
	}

	if message := strings.TrimSpace(node.Message.Text); message != "" {
		parts = append(parts, "Реплика этой точки сценария:\n"+message)
	}

	return strings.Join(parts, "\n\n")
}

func endingResume(scenario *scenariodomain.Scenario, endingNodeID, feedback string) string {
	parts := make([]string, 0, 2)

	if trimmed := strings.TrimSpace(feedback); trimmed != "" {
		parts = append(parts, trimmed)
	}

	if ending, ok := scenario.Nodes[endingNodeID]; ok {
		if message := strings.TrimSpace(ending.Message.Text); message != "" {
			parts = append(parts, message)
		}
	}

	return strings.Join(parts, "\n\n")
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
