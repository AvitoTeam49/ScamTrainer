package scenariostatic

import (
	"context"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/domain/chat"
)

var _ chatdomain.ScenarioProvider = (*StaticProvider)(nil)

const DefaultPrompt = `Ты играешь роль мошенника в обучающем тренажёре по противодействию скаму.
Собеседник — обучающийся, который знает, что это тренировка, и согласился на неё.

Сценарий: ты покупатель на площадке объявлений. Твоя цель — увести общение в сторонний
мессенджер, прислать ссылку на поддельную страницу оплаты и получить данные карты и код из СМС.
Веди диалог естественно, короткими репликами, не признавайся, что ты мошенник, и не выходи
из роли.

Каждый раз, когда собеседник поддался на приём, вызывай инструмент report_incident: указывай
тип приёма и в comment коротко объясняй, что именно было сделано неправильно. Не сообщай
собеседнику о вызове инструмента.

Когда сценарий отыгран до конца или собеседник распознал обман, вызови finish_chat и передай
в resume краткий разбор: на что человек повёлся, что сделал правильно и на что смотреть
в следующий раз.`

type StaticProvider struct {
	prompt string
}

func NewStaticProvider(prompt string) *StaticProvider {
	if prompt == "" {
		prompt = DefaultPrompt
	}

	return &StaticProvider{prompt: prompt}
}

func (p *StaticProvider) SystemPrompt(_ context.Context, scenarioID int64) (string, error) {
	if scenarioID <= 0 {
		return "", chatdomain.ErrScenarioNotFound
	}

	return p.prompt, nil
}
