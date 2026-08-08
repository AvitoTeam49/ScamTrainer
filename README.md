# ScamTrainer

Тренажёр против мошенников: пользователь ведёт диалог со сценарным LLM-агентом,
его решения оцениваются движком сценариев, итог попадает в профиль и лидерборд.

## Запуск

```bash
DEEPSEEK_API_KEY=<ключ> docker compose up --build
```

Миграции накатываются на старте. API доступен на `http://localhost:8080/api`,
проверка живости — `GET http://localhost:8080/healthz`.

Переменные: `DEEPSEEK_API_KEY` обязательна, `JWT_SECRET` и `DEEPSEEK_MODEL` имеют
dev-значения по умолчанию. Перед публичным деплоем `JWT_SECRET` нужно задать своим.

## Сквозной сценарий

```bash
API=http://localhost:8080/api/v1

curl -X POST $API/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"trainee@example.com","password":"supersecret"}'

TOKEN=$(curl -s -X POST $API/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"trainee@example.com","password":"supersecret"}' \
  | sed -E 's/.*"access_token":"([^"]+)".*/\1/')

curl -X POST $API/users -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"username":"trainee"}'

curl $API/scenarios -H "Authorization: Bearer $TOKEN"

curl -X POST $API/chats -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"scenario_id":1}'

# в соседнем терминале: поток событий чата
curl -N $API/chats/1/events -H "Authorization: Bearer $TOKEN"

curl -X POST $API/chats/1/messages -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"content":"По сторонним ссылкам я не перехожу, давайте через приложение"}'

curl $API/users/me -H "Authorization: Bearer $TOKEN"
curl $API/users/leaderboard -H "Authorization: Bearer $TOKEN"
```

Отправка сообщения отвечает `202`: ход агента идёт в фоне, результат приходит
событиями `message` / `decision` / `chat` в SSE-поток.

## Разработка

```bash
cd backend
go build ./... && go vet ./... && go test ./...
golangci-lint run ./...
sqlc generate   # после правок queries.sql
```

Границы доменов, контракты и принятые решения — в `backend/docs/chat.md`,
продуктовая цель и стек — в `AGENT.MD`.
