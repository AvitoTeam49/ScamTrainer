-- +goose Up
CREATE TABLE chats (
    id              bigserial PRIMARY KEY,
    user_id         bigint NOT NULL,
    scenario_id     bigint NOT NULL,
    title           text NOT NULL,
    status          text NOT NULL CHECK (status IN ('active', 'finished', 'abandoned')),
    resume          text NOT NULL DEFAULT '',
    score           bigint NOT NULL DEFAULT 0,
    current_node_id text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL,
    finished_at     timestamptz
);

CREATE INDEX idx_chats_user_id_id ON chats (user_id, id DESC);

CREATE TABLE messages (
    id          bigserial PRIMARY KEY,
    chat_id     bigint NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    sender_type text NOT NULL CHECK (sender_type IN ('agent', 'user')),
    content     text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_messages_chat_id_id ON messages (chat_id, id DESC);

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

-- +goose Down
DROP TABLE chat_decisions;

DROP TABLE messages;

DROP TABLE chats;
