-- name: GetChatByID :one
SELECT id, user_id, scenario_id, session_id, title, status, resume, score, current_node_id, created_at, finished_at
FROM chats
WHERE id = sqlc.arg(id);

-- name: ListChatsByUserID :many
SELECT id, user_id, scenario_id, session_id, title, status, resume, score, current_node_id, created_at, finished_at
FROM chats
WHERE user_id = sqlc.arg(user_id)
  AND (sqlc.arg(after_id)::bigint = 0 OR id < sqlc.arg(after_id)::bigint)
ORDER BY id DESC
LIMIT sqlc.arg(lim);

-- name: CreateChat :one
INSERT INTO chats (user_id, scenario_id, session_id, title, status, resume, score, current_node_id, created_at)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(scenario_id),
    sqlc.arg(session_id),
    sqlc.arg(title),
    sqlc.arg(status),
    sqlc.arg(resume),
    sqlc.arg(score),
    sqlc.arg(current_node_id),
    sqlc.arg(created_at)
)
RETURNING id;

-- name: CloseChat :execrows
UPDATE chats
SET status = sqlc.arg(status),
    resume = sqlc.arg(resume),
    score = sqlc.arg(score),
    finished_at = sqlc.narg(finished_at)
WHERE id = sqlc.arg(id)
  AND status = 'active';

-- name: DeleteChat :execrows
DELETE FROM chats
WHERE id = sqlc.arg(id);

-- name: ListMessagesByChatID :many
SELECT id, chat_id, sender_type, content, created_at
FROM messages
WHERE chat_id = sqlc.arg(chat_id)
  AND (sqlc.arg(after_id)::bigint = 0 OR id < sqlc.arg(after_id)::bigint)
ORDER BY id DESC
LIMIT sqlc.arg(lim);

-- name: CreateMessage :one
INSERT INTO messages (chat_id, sender_type, content)
VALUES (sqlc.arg(chat_id), sqlc.arg(sender_type), sqlc.arg(content))
RETURNING id, created_at;

-- name: ListDecisionsByChatID :many
SELECT id, chat_id, node_id, transition_id, target_node_id, score_delta, feedback, created_at
FROM chat_decisions
WHERE chat_id = sqlc.arg(chat_id)
  AND (sqlc.arg(after_id)::bigint = 0 OR id < sqlc.arg(after_id)::bigint)
ORDER BY id DESC
LIMIT sqlc.arg(lim);

-- name: CreateDecision :one
WITH updated_chat AS (
    UPDATE chats
    SET score = sqlc.arg(score),
        current_node_id = sqlc.arg(target_node_id)
    WHERE chats.id = sqlc.arg(chat_id)
      AND chats.status = 'active'
    RETURNING chats.id
), new_decision AS (
    INSERT INTO chat_decisions (chat_id, node_id, transition_id, target_node_id, score_delta, feedback)
    SELECT
        updated_chat.id,
        sqlc.arg(node_id)::text,
        sqlc.arg(transition_id)::text,
        sqlc.arg(target_node_id)::text,
        sqlc.arg(score_delta)::bigint,
        sqlc.arg(feedback)::text
    FROM updated_chat
    RETURNING id, created_at
)
SELECT id, created_at
FROM new_decision;
