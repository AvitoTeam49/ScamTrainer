
-- +goose Up
CREATE TABLE chats (
    id          bigserial PRIMARY KEY,
    user_id     bigint NOT NULL,
    scenario_id bigint NOT NULL,
    title       text NOT NULL,
    status      text NOT NULL CHECK (status IN ('active', 'finished', 'abandoned')),
    resume      text NOT NULL DEFAULT '',
    score       bigint NOT NULL CHECK (score >= 0),
    created_at  timestamptz NOT NULL,
    finished_at timestamptz
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

-- +goose Down
DROP TABLE incidents;

DROP TABLE messages;

DROP TABLE chats;
