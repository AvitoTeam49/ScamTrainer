package chatusecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

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

func TestRunAgentTurn_ToolCallWithoutContentClosesScenarioWithEndingText(t *testing.T) {
	fixture := newFixture(t, "start")

	fixture.agent.replies = []*agent.Reply{
		{ToolCalls: []agent.ToolCall{applyTransitionCall("stay_on_platform")}},
	}

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err != nil {
		t.Fatalf("run agent turn: %v", err)
	}

	if got := fixture.chats.chat.Status; got != chatdomain.ChatStatusFinished {
		t.Fatalf("status: got %q, want %q", got, chatdomain.ChatStatusFinished)
	}

	want := []chatdomain.EventType{
		chatdomain.EventTypeDecision,
		chatdomain.EventTypeChat,
		chatdomain.EventTypeMessage,
	}
	if got := fixture.events.types(); !equalTypes(got, want) {
		t.Fatalf("event sequence: got %v, want %v", got, want)
	}

	if got := fixture.awards.calls; len(got) != 1 || got[0].scoreDelta != 20 {
		t.Fatalf("awards: got %v, want one award of 20", got)
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

func TestSendMessage_RejectsContentLongerThanLimit(t *testing.T) {
	fixture := newFixture(t, "start")
	content := strings.Repeat("я", chatdomain.MaxMessageLength+1)

	_, err := fixture.service.SendMessage(context.Background(), 1, 42, content)
	if !errors.Is(err, chatdomain.ErrMessageTooLong) {
		t.Fatalf("got %v, want %v", err, chatdomain.ErrMessageTooLong)
	}

	if got := len(fixture.messages.created); got != 0 {
		t.Fatalf("stored messages: got %d, want 0", got)
	}

	if got := fixture.events.types(); len(got) != 0 {
		t.Fatalf("events: got %v, want none", got)
	}
}

func TestSendMessage_MeasuresLimitInRunesNotBytes(t *testing.T) {
	fixture := newFixture(t, "start")
	content := strings.Repeat("я", chatdomain.MaxMessageLength)

	if len(content) <= chatdomain.MaxMessageLength {
		t.Fatalf("кириллица должна занимать больше байт, чем рун: %d байт", len(content))
	}

	if _, err := fixture.service.SendMessage(context.Background(), 1, 42, content); err != nil {
		t.Fatalf("сообщение ровно в лимит рун должно приниматься: %v", err)
	}
}

func TestRunAgentTurn_SerializesConcurrentTurnsOfSameChat(t *testing.T) {
	fixture := newFixture(t, "start")

	fixture.agent.replies = []*agent.Reply{
		{
			Content:   "Понимаю",
			ToolCalls: []agent.ToolCall{applyTransitionCall("stay_on_platform")},
		},
		{Content: "Хорошо"},
		{Content: "Лишняя реплика после финала"},
		{Content: "И ещё одна"},
	}

	var wg sync.WaitGroup

	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()

			_ = fixture.service.RunAgentTurn(context.Background(), 1)
		}()
	}
	wg.Wait()

	if got := len(fixture.decisions.created); got != 1 {
		t.Fatalf("decisions: got %d, want 1", got)
	}

	if got := len(fixture.awards.calls); got != 1 {
		t.Fatalf("awards: got %d, want 1", got)
	}

	if got := len(fixture.messages.bySender(chatdomain.SenderTypeAgent)); got != 1 {
		t.Fatalf("agent messages: got %d, want 1 — второй ход дописал реплику после финала", got)
	}

	if got := fixture.chats.chat.Score; got != 20 {
		t.Fatalf("score: got %d, want 20", got)
	}
}

func TestAbandonChat_ClosesChatWithoutAwardingScore(t *testing.T) {
	fixture := newFixture(t, "start")
	fixture.chats.chat.Score = 20

	chat, err := fixture.service.AbandonChat(context.Background(), 1, 42)
	if err != nil {
		t.Fatalf("abandon chat: %v", err)
	}

	if chat.Status != chatdomain.ChatStatusAbandoned {
		t.Fatalf("status: got %q, want %q", chat.Status, chatdomain.ChatStatusAbandoned)
	}

	if chat.FinishedAt == nil {
		t.Fatal("abandoned chat must carry a finish timestamp")
	}

	if got := len(fixture.awards.calls); got != 0 {
		t.Fatalf("awards: got %d, want none — брошенный сценарий не оценивается", got)
	}

	if got := fixture.events.types(); len(got) != 1 || got[0] != chatdomain.EventTypeChat {
		t.Fatalf("events: got %v, want [chat]", got)
	}
}

func TestAbandonChat_PublishesChatSnapshotDetachedFromLiveChat(t *testing.T) {
	fixture := newFixture(t, "start")
	fixture.chats.chat.Score = 20

	chat, err := fixture.service.AbandonChat(context.Background(), 1, 42)
	if err != nil {
		t.Fatalf("abandon chat: %v", err)
	}

	recorded := fixture.events.recorded()
	if len(recorded) != 1 || recorded[0].Chat == nil {
		t.Fatalf("events: got %v, want one chat event", fixture.events.types())
	}

	snapshot := recorded[0].Chat
	if snapshot == chat {
		t.Fatal("событие не должно ссылаться на живой чат")
	}

	if snapshot.FinishedAt == chat.FinishedAt {
		t.Fatal("FinishedAt должен копироваться, а не разделяться по указателю")
	}

	finishedAt := *chat.FinishedAt

	chat.Score = 999
	*chat.FinishedAt = time.Unix(0, 0).UTC()

	if snapshot.Score != 20 {
		t.Fatalf("snapshot score: got %d, want 20", snapshot.Score)
	}

	if !snapshot.FinishedAt.Equal(finishedAt) {
		t.Fatalf("snapshot finished at: got %v, want %v", snapshot.FinishedAt, finishedAt)
	}
}

func TestAbandonChat_RejectsChatOfAnotherUser(t *testing.T) {
	fixture := newFixture(t, "start")

	if _, err := fixture.service.AbandonChat(context.Background(), 1, 999); !errors.Is(err, chatdomain.ErrChatAccessDenied) {
		t.Fatalf("got %v, want %v", err, chatdomain.ErrChatAccessDenied)
	}
}

func TestAbandonChat_RejectsAlreadyClosedChat(t *testing.T) {
	fixture := newFixture(t, "start")
	fixture.chats.chat.Status = chatdomain.ChatStatusFinished

	if _, err := fixture.service.AbandonChat(context.Background(), 1, 42); !errors.Is(err, chatdomain.ErrChatFinished) {
		t.Fatalf("got %v, want %v", err, chatdomain.ErrChatFinished)
	}
}

func TestStartChat_DropsChatWhenOpeningMessageFails(t *testing.T) {
	fixture := newFixture(t, "start")
	fixture.messages.createErr = errors.New("storage is down")

	if _, err := fixture.service.StartChat(context.Background(), 42, 7, ""); err == nil {
		t.Fatal("start chat must fail when the opening message cannot be stored")
	}

	if got := fixture.chats.deleted; got != 1 {
		t.Fatalf("deleted chats: got %d, want 1 — чат без завязки должен откатываться", got)
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

	opening := fixture.messages.bySender(chatdomain.SenderTypeAgent)
	if len(opening) != 1 || opening[0].Content != "Готов купить, оформим по моей ссылке." {
		t.Fatalf("chat must open with the start node line, got %v", opening)
	}
}

func TestStartChat_RejectsTitleLongerThanLimit(t *testing.T) {
	fixture := newFixture(t, "start")
	title := strings.Repeat("я", chatdomain.MaxChatTitleLength+1)

	_, err := fixture.service.StartChat(context.Background(), 42, 7, title)
	if !errors.Is(err, chatdomain.ErrTitleTooLong) {
		t.Fatalf("got %v, want %v", err, chatdomain.ErrTitleTooLong)
	}

	if got := len(fixture.messages.created); got != 0 {
		t.Fatalf("отклонённый чат не должен заводить сообщений, got %d", got)
	}
}

func TestStartChat_MeasuresTitleLimitInRunesNotBytes(t *testing.T) {
	fixture := newFixture(t, "start")
	title := strings.Repeat("я", chatdomain.MaxChatTitleLength)

	if len(title) <= chatdomain.MaxChatTitleLength {
		t.Fatalf("кириллица должна занимать больше байт, чем рун: %d байт", len(title))
	}

	chat, err := fixture.service.StartChat(context.Background(), 42, 7, title)
	if err != nil {
		t.Fatalf("заголовок ровно в лимит рун должен приниматься: %v", err)
	}

	if chat.Title != title {
		t.Fatalf("title: got %q, want %q", chat.Title, title)
	}
}

func TestRunAgentTurn_ClampsHistoryLimitToCursorMax(t *testing.T) {
	fixture := newFixtureWith(t, "start", Options{HistoryLimit: chatdomain.MaxCursorLimit + 1})
	fixture.agent.replies = []*agent.Reply{{Content: "Ага"}}

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err != nil {
		t.Fatalf("run agent turn: %v", err)
	}

	if len(fixture.messages.cursors) == 0 {
		t.Fatal("ход агента должен читать историю чата")
	}

	if got := fixture.messages.cursors[0].Limit; got != chatdomain.MaxCursorLimit {
		t.Fatalf("history cursor limit: got %d, want %d", got, chatdomain.MaxCursorLimit)
	}
}

func TestRunAgentTurn_UsesDefaultHistoryLimitWhenUnset(t *testing.T) {
	fixture := newFixture(t, "start")
	fixture.agent.replies = []*agent.Reply{{Content: "Ага"}}

	if err := fixture.service.RunAgentTurn(context.Background(), 1); err != nil {
		t.Fatalf("run agent turn: %v", err)
	}

	if len(fixture.messages.cursors) == 0 {
		t.Fatal("ход агента должен читать историю чата")
	}

	if got := fixture.messages.cursors[0].Limit; got != defaultHistoryLimit {
		t.Fatalf("history cursor limit: got %d, want %d", got, defaultHistoryLimit)
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
	messages  *fakeMessageRepository
}

func newFixture(t *testing.T, currentNodeID string) *fixture {
	t.Helper()

	return newFixtureWith(t, currentNodeID, Options{})
}

func newFixtureWith(t *testing.T, currentNodeID string, options Options) *fixture {
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
	messages := &fakeMessageRepository{}

	return &fixture{
		service: New(
			chats,
			messages,
			decisions,
			scenarios,
			sessions,
			training.NewService(scenarios, sessions, scenariodomain.NewEngine(), training.UUIDGenerator{}),
			events,
			awards,
			chatAgent,
			options,
		),
		chats:     chats,
		decisions: decisions,
		events:    events,
		agent:     chatAgent,
		sessions:  sessions,
		awards:    awards,
		messages:  messages,
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
				ID:      "start",
				Type:    scenariodomain.NodeTypeDecision,
				Message: scenariodomain.Message{Author: "scammer", Text: "Готов купить, оформим по моей ссылке."},
				LLM:     scenariodomain.NodeLLM{ReplyInstruction: "Убеждай открыть ссылку"},
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
	deleted  int
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

func (f *fakeChatRepository) Close(context.Context, *chatdomain.Chat) (bool, error) {
	f.finishes++

	return f.finishes == 1, nil
}

func (f *fakeChatRepository) Delete(context.Context, int64) error {
	f.deleted++

	return nil
}

type fakeMessageRepository struct {
	created   []*chatdomain.Message
	cursors   []chatdomain.Cursor
	createErr error
}

func (f *fakeMessageRepository) ListByChatID(
	_ context.Context, _ int64, cursor chatdomain.Cursor,
) ([]*chatdomain.Message, error) {
	f.cursors = append(f.cursors, cursor)

	if err := cursor.Validate(); err != nil {
		return nil, err
	}

	return f.created, nil
}

func (f *fakeMessageRepository) Create(_ context.Context, message *chatdomain.Message) error {
	if f.createErr != nil {
		return f.createErr
	}

	f.created = append(f.created, message)

	return nil
}

func (f *fakeMessageRepository) bySender(sender chatdomain.SenderType) []*chatdomain.Message {
	found := make([]*chatdomain.Message, 0, len(f.created))
	for _, message := range f.created {
		if message.SenderType == sender {
			found = append(found, message)
		}
	}

	return found
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
