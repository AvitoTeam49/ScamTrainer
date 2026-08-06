-- +goose Up

-- Счёт чата теперь ведёт движок сценариев: старт с нуля и накопление знаковых
-- score_delta из переходов графа. Прежняя шкала (старт 100 минус вес инцидента)
-- уходит вместе с ограничением на неотрицательность.
ALTER TABLE chats DROP CONSTRAINT chats_score_check;
ALTER TABLE chats ALTER COLUMN score SET DEFAULT 0;

-- Позиция в графе сценария. Пустая строка означает, что чат создан до перехода
-- на графы и текущий узел неизвестен.
ALTER TABLE chats ADD COLUMN current_node_id text NOT NULL DEFAULT '';

-- Журнал принятых решений повторяет scenario.Decision: он и есть разбор
-- диалога, который получает пользователь.
CREATE TABLE chat_decisions (
    id             bigserial PRIMARY KEY,
    chat_id        bigint NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    node_id        text NOT NULL,
    transition_id  text NOT NULL,
    target_node_id text NOT NULL,
    score_delta    bigint NOT NULL,
    feedback       text NOT NULL DEFAULT '',
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_chat_decisions_chat_id_id ON chat_decisions (chat_id, id DESC);

DROP TABLE incidents;

-- +goose Down

CREATE TABLE incidents (
    id         bigserial PRIMARY KEY,
    chat_id    bigint NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    type       text NOT NULL CHECK (type IN (
        'left_platform',
        'disclosed_personal_data',
        'followed_phishing_link',
        'installed_software',
        'disclosed_card_data',
        'disclosed_code',
        'made_transfer'
    )),
    comment    text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_incidents_chat_id_id ON incidents (chat_id, id DESC);

DROP TABLE chat_decisions;

ALTER TABLE chats DROP COLUMN current_node_id;

ALTER TABLE chats ALTER COLUMN score DROP DEFAULT;
UPDATE chats SET score = 0 WHERE score < 0;
ALTER TABLE chats ADD CONSTRAINT chats_score_check CHECK (score >= 0);
