package chatusecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/agent"
	chatdomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
	"github.com/AvitoTeam49/ScamTrainer/backend/internal/training"
)

func TestSendMessage_StoresAndPublishesUserMessageWithoutCallingAgent(t *testing.T) {
	fixture := newFixture(t, "start")

	message, err := fixture.service.SendMessage(context.Background(), 1, 42, "Привет")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}

	if message.SenderType != chatdomain.SenderTypeUser {
		t.Fatalf("sender: got %q, want %q", message.SenderType, chatdomain.SenderTypeUser)
	}

	if fixture.agent.calls != 0 {
		t.Fatalf("agent must not be called synchronously, got %d calls", fixture.agent.calls)
	}

	if got := fixture.events.types(); len(got) != 1 || got[0] != chatdomain.EventTypeMessage {
		t.Fatalf("events: got %v, want [message]", got)
	}
}

func TestRunAgentTurn_UnsafeTransitionSubtractsScoreAndEndsScenario(t *testing.T) {
	fixture := newFixture(t, "start")

	fixture.agent.replies = []*agent.Reply{
		{
			Content:   "Вот ссылка на оплату",
			ToolCalls: []agent.ToolCall{applyTransitionCall("open_link")},
		},
		{Content: "Спасибо, жду оплату"},
	}

	if _, err := fixture.service.SendMessage(context.Background(), 1, 42, "Хорошо, сейчас открою"); err != nil {
		t.Fatalf("send message: %v", err)
	}

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err != nil {
		t.Fatalf("run agent turn: %v", err)
	}

	if got := fixture.chats.chat.Score; got != -20 {
		t.Fatalf("score: got %d, want -20", got)
	}

	if got := fixture.chats.chat.CurrentNodeID; got != "unsafe_ending" {
		t.Fatalf("current node: got %q, want %q", got, "unsafe_ending")
	}

	if got := fixture.chats.chat.Status; got != chatdomain.ChatStatusFinished {
		t.Fatalf("status: got %q, want %q", got, chatdomain.ChatStatusFinished)
	}

	if !strings.Contains(fixture.chats.chat.Resume, "фишинга") {
		t.Fatalf("resume should carry the transition feedback, got %q", fixture.chats.chat.Resume)
	}

	want := []chatdomain.EventType{
		chatdomain.EventTypeMessage,
		chatdomain.EventTypeDecision,
		chatdomain.EventTypeChat,
		chatdomain.EventTypeMessage,
	}
	if got := fixture.events.types(); !equalTypes(got, want) {
		t.Fatalf("event sequence: got %v, want %v", got, want)
	}
}

func TestRunAgentTurn_SafeTransitionAddsScore(t *testing.T) {
	fixture := newFixture(t, "start")

	fixture.agent.replies = []*agent.Reply{
		{
			Content:   "Понимаю",
			ToolCalls: []agent.ToolCall{applyTransitionCall("stay_on_platform")},
		},
		{Content: "Хорошо"},
	}

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err != nil {
		t.Fatalf("run agent turn: %v", err)
	}

	if got := fixture.chats.chat.Score; got != 20 {
		t.Fatalf("score: got %d, want 20", got)
	}

	if got := fixture.chats.chat.CurrentNodeID; got != "safe_ending" {
		t.Fatalf("current node: got %q, want %q", got, "safe_ending")
	}
}

func TestRunAgentTurn_AwardsChatScoreToUserOnce(t *testing.T) {
	fixture := newFixture(t, "start")

	fixture.agent.replies = []*agent.Reply{
		{
			Content:   "Понимаю",
			ToolCalls: []agent.ToolCall{applyTransitionCall("stay_on_platform")},
		},
		{Content: "Хорошо"},
	}

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err != nil {
		t.Fatalf("run agent turn: %v", err)
	}

	want := []scoreAward{{userID: 42, scoreDelta: 20}}
	if got := fixture.awards.calls; !equalAwards(got, want) {
		t.Fatalf("awards: got %v, want %v", got, want)
	}
}

func TestRunAgentTurn_DoesNotAwardTwiceWhenChatAlreadyFinished(t *testing.T) {
	fixture := newFixture(t, "start")
	fixture.chats.finishes = 1

	fixture.agent.replies = []*agent.Reply{
		{
			Content:   "Понимаю",
			ToolCalls: []agent.ToolCall{applyTransitionCall("stay_on_platform")},
		},
		{Content: "Хорошо"},
	}

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err == nil {
		t.Fatal("run agent turn should fail when the chat was finished by another turn")
	}

	if got := len(fixture.awards.calls); got != 0 {
		t.Fatalf("awards: got %d calls, want none", got)
	}
}

func TestRunAgentTurn_RestoresLostSession(t *testing.T) {
	fixture := newFixture(t, "start")

	if _, err := fixture.sessions.GetById(context.Background(), "session-1"); err == nil {
		t.Fatal("session must be absent before the turn")
	}

	fixture.agent.replies = []*agent.Reply{
		{
			Content:   "Понимаю",
			ToolCalls: []agent.ToolCall{applyTransitionCall("stay_on_platform")},
		},
		{Content: "Хорошо"},
	}

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err != nil {
		t.Fatalf("run agent turn: %v", err)
	}

	if got := fixture.chats.chat.CurrentNodeID; got != "safe_ending" {
		t.Fatalf("current node: got %q, want %q", got, "safe_ending")
	}
}

func TestRunAgentTurn_UnknownTransitionLeavesScoreUntouched(t *testing.T) {
	fixture := newFixture(t, "start")

	fixture.agent.replies = []*agent.Reply{
		{
			Content:   "Ок",
			ToolCalls: []agent.ToolCall{applyTransitionCall("does_not_exist")},
		},
		{Content: "Продолжаем"},
	}

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err != nil {
		t.Fatalf("run agent turn: %v", err)
	}

	if got := fixture.chats.chat.Score; got != 0 {
		t.Fatalf("score: got %d, want 0", got)
	}

	if got := fixture.chats.chat.CurrentNodeID; got != "start" {
		t.Fatalf("current node: got %q, want %q", got, "start")
	}

	if len(fixture.decisions.created) != 0 {
		t.Fatalf("nothing should be journalled, got %d entries", len(fixture.decisions.created))
	}
}

func TestRunAgentTurn_PublishesErrorEventWhenAgentFails(t *testing.T) {
	fixture := newFixture(t, "start")
	fixture.agent.err = errors.New("deepseek is down")

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err == nil {
		t.Fatal("expected the agent failure to be returned")
	}

	events := fixture.events.recorded()
	if len(events) != 1 || events[0].Type != chatdomain.EventTypeError {
		t.Fatalf("events: got %v, want one error event", fixture.events.types())
	}

	if events[0].Reason != genericFailureReason {
		t.Fatalf("reason must not leak internals: got %q", events[0].Reason)
	}

	if events[0].ChatID != 1 {
		t.Fatalf("error event chat id: got %d, want 1", events[0].ChatID)
	}
}

func TestRunAgentTurn_ToolEnumComesFromCurrentNode(t *testing.T) {
	fixture := newFixture(t, "start")
	fixture.agent.replies = []*agent.Reply{{Content: "Здравствуйте"}}

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err != nil {
		t.Fatalf("run agent turn: %v", err)
	}

	var schema struct {
		Properties struct {
			TransitionID struct {
				Enum []string `json:"enum"`
			} `json:"transition_id"`
		} `json:"properties"`
	}

	tools := fixture.agent.requests[0].Tools
	if len(tools) == 0 || tools[0].Name != toolApplyTransition {
		t.Fatalf("expected %s to be offered first, got %+v", toolApplyTransition, tools)
	}

	if err := json.Unmarshal(tools[0].Parameters, &schema); err != nil {
		t.Fatalf("unmarshal tool schema: %v", err)
	}

	got := schema.Properties.TransitionID.Enum
	if len(got) != 2 || got[0] != "open_link" || got[1] != "stay_on_platform" {
		t.Fatalf("enum: got %v, want [open_link stay_on_platform]", got)
	}
}

func TestSendMessage_RejectsChatOfAnotherUser(t *testing.T) {
	fixture := newFixture(t, "start")

	_, err := fixture.service.SendMessage(context.Background(), 1, 999, "Привет")
	if !errors.Is(err, chatdomain.ErrChatAccessDenied) {
		t.Fatalf("got %v, want %v", err, chatdomain.ErrChatAccessDenied)
	}

	if len(fixture.events.recorded()) != 0 {
		t.Fatal("a rejected message must not be published")
	}
}

func TestSendMessage_RejectsChatParkedOnEndingNode(t *testing.T) {
	fixture := newFixture(t, "safe_ending")

	_, err := fixture.service.SendMessage(context.Background(), 1, 42, "Привет")
	if !errors.Is(err, chatdomain.ErrChatFinished) {
		t.Fatalf("got %v, want %v", err, chatdomain.ErrChatFinished)
	}
}

func TestStartChat_PositionsChatOnStartNode(t *testing.T) {
	fixture := newFixture(t, "start")

	chat, err := fixture.service.StartChat(context.Background(), 42, 7, "")
	if err != nil {
		t.Fatalf("start chat: %v", err)
	}

	if chat.CurrentNodeID != "start" {
		t.Fatalf("current node: got %q, want %q", chat.CurrentNodeID, "start")
	}

	if chat.Score != 0 {
		t.Fatalf("score: got %d, want 0", chat.Score)
	}

	if chat.Title != "Поддельная ссылка на доставку" {
		t.Fatalf("title should fall back to the scenario title, got %q", chat.Title)
	}
}

type fixture struct {
	service   *Service
	chats     *fakeChatRepository
	decisions *fakeDecisionRepository
	events    *fakeEventPublisher
	agent     *fakeAgent
	sessions  *training.InMemorySessionRepository
	awards    *fakeScoreAwarder
}

func newFixture(t *testing.T, currentNodeID string) *fixture {
	t.Helper()

	chats := &fakeChatRepository{chat: &chatdomain.Chat{
		ID:            1,
		UserID:        42,
		ScenarioID:    7,
		SessionID:     "session-1",
		Status:        chatdomain.ChatStatusActive,
		Score:         0,
		CurrentNodeID: currentNodeID,
	}}
	decisions := &fakeDecisionRepository{}
	events := &fakeEventPublisher{}
	chatAgent := &fakeAgent{}
	scenarios := &fakeScenarioSource{scenario: testScenario()}
	sessions := training.NewInMemorySessionRepository()
	awards := &fakeScoreAwarder{}

	return &fixture{
		service: New(
			chats,
			&fakeMessageRepository{},
			decisions,
			scenarios,
			sessions,
			training.NewService(scenarios, sessions, scenariodomain.NewEngine(), training.UUIDGenerator{}),
			events,
			awards,
			chatAgent,
			Options{},
		),
		chats:     chats,
		decisions: decisions,
		events:    events,
		agent:     chatAgent,
		sessions:  sessions,
		awards:    awards,
	}
}

type fakeScoreAwarder struct {
	calls []scoreAward
}

type scoreAward struct {
	userID     int64
	scoreDelta int
}

func (f *fakeScoreAwarder) UpdateUserScore(_ context.Context, userID int64, scoreDelta int) error {
	f.calls = append(f.calls, scoreAward{userID: userID, scoreDelta: scoreDelta})

	return nil
}

func equalAwards(got, want []scoreAward) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

func equalTypes(got, want []chatdomain.EventType) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

func applyTransitionCall(transitionID string) agent.ToolCall {
	arguments, err := json.Marshal(map[string]string{"transition_id": transitionID})
	if err != nil {
		panic(err)
	}

	return agent.ToolCall{ID: "call-1", Name: toolApplyTransition, Arguments: arguments}
}

func testScenario() *scenariodomain.Scenario {
	return &scenariodomain.Scenario{
		ID:          7,
		Title:       "Поддельная ссылка на доставку",
		Role:        scenariodomain.RoleSeller,
		StartNodeID: "start",
		LLM:         scenariodomain.ScenarioLLM{CharacterPrompt: "Ты играешь роль покупателя"},
		Nodes: map[string]*scenariodomain.Node{
			"start": {
				ID:   "start",
				Type: scenariodomain.NodeTypeDecision,
				LLM:  scenariodomain.NodeLLM{ReplyInstruction: "Убеждай открыть ссылку"},
				Transitions: []scenariodomain.Transition{
					{
						ID:          "open_link",
						Description: "Пользователь соглашается открыть внешнюю ссылку",
						Examples:    []string{"Хорошо, сейчас открою"},
						ToNodeID:    "unsafe_ending",
						ScoreDelta:  -20,
						Feedback:    "Переход по ссылке повышает риск фишинга.",
					},
					{
						ID:          "stay_on_platform",
						Description: "Пользователь остаётся внутри платформы",
						ToNodeID:    "safe_ending",
						ScoreDelta:  20,
						Feedback:    "Сделка внутри платформы безопаснее.",
					},
				},
			},
			"unsafe_ending": {
				ID:      "unsafe_ending",
				Type:    scenariodomain.NodeTypeEnding,
				Outcome: "unsafe",
				Message: scenariodomain.Message{Text: "Ссылка вела на поддельную страницу оплаты."},
			},
			"safe_ending": {
				ID:      "safe_ending",
				Type:    scenariodomain.NodeTypeEnding,
				Outcome: "safe",
				Message: scenariodomain.Message{Text: "Вы отказались переходить по ссылке."},
			},
		},
	}
}

type fakeScenarioSource struct {
	scenario *scenariodomain.Scenario
}

func (f *fakeScenarioSource) GetById(context.Context, int) (*scenariodomain.Scenario, error) {
	return f.scenario, nil
}

type fakeEventPublisher struct {
	events []chatdomain.Event
}

func (f *fakeEventPublisher) Publish(_ context.Context, event chatdomain.Event) {
	f.events = append(f.events, event)
}

func (f *fakeEventPublisher) recorded() []chatdomain.Event {
	return f.events
}

func (f *fakeEventPublisher) types() []chatdomain.EventType {
	types := make([]chatdomain.EventType, 0, len(f.events))
	for _, event := range f.events {
		types = append(types, event.Type)
	}

	return types
}

type fakeChatRepository struct {
	chat     *chatdomain.Chat
	finishes int
}

func (f *fakeChatRepository) GetByID(context.Context, int64) (*chatdomain.Chat, error) {
	return f.chat, nil
}

func (f *fakeChatRepository) ListByUserID(
	context.Context, int64, chatdomain.Cursor,
) ([]*chatdomain.Chat, error) {
	return nil, nil
}

func (f *fakeChatRepository) Create(_ context.Context, chat *chatdomain.Chat) error {
	chat.ID = 1
	return nil
}

func (f *fakeChatRepository) Finish(context.Context, *chatdomain.Chat) (bool, error) {
	f.finishes++

	return f.finishes == 1, nil
}

func (f *fakeChatRepository) Delete(context.Context, int64) error { return nil }

type fakeMessageRepository struct {
	created []*chatdomain.Message
}

func (f *fakeMessageRepository) ListByChatID(
	context.Context, int64, chatdomain.Cursor,
) ([]*chatdomain.Message, error) {
	return f.created, nil
}

func (f *fakeMessageRepository) Create(_ context.Context, message *chatdomain.Message) error {
	f.created = append(f.created, message)
	return nil
}

type fakeDecisionRepository struct {
	created []*chatdomain.Decision
	scores  []int64
}

func (f *fakeDecisionRepository) ListByChatID(
	context.Context, int64, chatdomain.Cursor,
) ([]*chatdomain.Decision, error) {
	return f.created, nil
}

func (f *fakeDecisionRepository) Create(
	_ context.Context,
	decision *chatdomain.Decision,
	score int64,
) error {
	f.created = append(f.created, decision)
	f.scores = append(f.scores, score)

	return nil
}

type fakeAgent struct {
	replies  []*agent.Reply
	requests []agent.Request
	calls    int
	err      error
}

func (f *fakeAgent) Complete(_ context.Context, request agent.Request) (*agent.Reply, error) {
	f.requests = append(f.requests, request)
	f.calls++

	if f.err != nil {
		return nil, f.err
	}

	if f.calls > len(f.replies) {
		return &agent.Reply{Content: "..."}, nil
	}

	return f.replies[f.calls-1], nil
}
