package chatusecase

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/agent"
	chatdomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
	scenariodomain "github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/scenario"
)

func TestSendMessage_UnsafeTransitionSubtractsScoreAndEndsScenario(t *testing.T) {
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

	if got := fixture.chats.chat.Score; got != -20 {
		t.Fatalf("score: got %d, want -20 (score_delta of the chosen transition)", got)
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

	if len(fixture.decisions.created) != 1 {
		t.Fatalf("decisions journal: got %d entries, want 1", len(fixture.decisions.created))
	}

	stored := fixture.decisions.created[0]
	if stored.TransitionID != "open_link" || stored.ScoreDelta != -20 {
		t.Fatalf("stored decision: got %+v", stored)
	}

	if fixture.decisions.scores[0] != -20 {
		t.Fatalf("score written with the journal entry: got %d, want -20", fixture.decisions.scores[0])
	}
}

func TestSendMessage_SafeTransitionAddsScore(t *testing.T) {
	fixture := newFixture(t, "start")

	fixture.agent.replies = []*agent.Reply{
		{
			Content:   "Понимаю",
			ToolCalls: []agent.ToolCall{applyTransitionCall("stay_on_platform")},
		},
		{Content: "Хорошо"},
	}

	if _, err := fixture.service.SendMessage(context.Background(), 1, 42, "По ссылкам не перехожу"); err != nil {
		t.Fatalf("send message: %v", err)
	}

	if got := fixture.chats.chat.Score; got != 20 {
		t.Fatalf("score: got %d, want 20", got)
	}

	if got := fixture.chats.chat.CurrentNodeID; got != "safe_ending" {
		t.Fatalf("current node: got %q, want %q", got, "safe_ending")
	}
}

// The engine rejects a transition that does not belong to the current node, and
// the rejection must reach the model as a tool result rather than kill the turn.
func TestSendMessage_UnknownTransitionLeavesScoreUntouched(t *testing.T) {
	fixture := newFixture(t, "start")

	fixture.agent.replies = []*agent.Reply{
		{
			Content:   "Ок",
			ToolCalls: []agent.ToolCall{applyTransitionCall("does_not_exist")},
		},
		{Content: "Продолжаем"},
	}

	if _, err := fixture.service.SendMessage(context.Background(), 1, 42, "Что-то нейтральное"); err != nil {
		t.Fatalf("send message: %v", err)
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

// The transition ids offered to the model must come from the node the chat
// currently sits on.
func TestSendMessage_ToolEnumComesFromCurrentNode(t *testing.T) {
	fixture := newFixture(t, "start")
	fixture.agent.replies = []*agent.Reply{{Content: "Здравствуйте"}}

	if _, err := fixture.service.SendMessage(context.Background(), 1, 42, "Привет"); err != nil {
		t.Fatalf("send message: %v", err)
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
	if err != chatdomain.ErrChatAccessDenied {
		t.Fatalf("got %v, want %v", err, chatdomain.ErrChatAccessDenied)
	}
}

func TestSendMessage_RejectsChatParkedOnEndingNode(t *testing.T) {
	fixture := newFixture(t, "safe_ending")

	_, err := fixture.service.SendMessage(context.Background(), 1, 42, "Привет")
	if err != chatdomain.ErrChatFinished {
		t.Fatalf("got %v, want %v", err, chatdomain.ErrChatFinished)
	}
}

// ---- fixture ----

type fixture struct {
	service   *Service
	chats     *fakeChatRepository
	decisions *fakeDecisionRepository
	agent     *fakeAgent
}

func newFixture(t *testing.T, currentNodeID string) *fixture {
	t.Helper()

	chats := &fakeChatRepository{chat: &chatdomain.Chat{
		ID:            1,
		UserID:        42,
		ScenarioID:    7,
		Status:        chatdomain.ChatStatusActive,
		Score:         0,
		CurrentNodeID: currentNodeID,
	}}
	decisions := &fakeDecisionRepository{}
	chatAgent := &fakeAgent{}

	return &fixture{
		service: New(
			chats,
			&fakeMessageRepository{},
			decisions,
			&fakeScenarioSource{scenario: testScenario()},
			chatAgent,
			Options{},
		),
		chats:     chats,
		decisions: decisions,
		agent:     chatAgent,
	}
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
		ID:          "seller_fake_delivery",
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

func (f *fakeScenarioSource) Scenario(context.Context, int64) (*scenariodomain.Scenario, error) {
	return f.scenario, nil
}

type fakeChatRepository struct {
	chat *chatdomain.Chat
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
	f.chat = chat
	return nil
}

func (f *fakeChatRepository) Update(context.Context, *chatdomain.Chat) error { return nil }
func (f *fakeChatRepository) Delete(context.Context, int64) error            { return nil }

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
}

func (f *fakeAgent) Complete(_ context.Context, request agent.Request) (*agent.Reply, error) {
	f.requests = append(f.requests, request)

	if f.calls >= len(f.replies) {
		return &agent.Reply{Content: "..."}, nil
	}

	reply := f.replies[f.calls]
	f.calls++

	return reply, nil
}
