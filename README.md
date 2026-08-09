# ScamTrainer

Тренажёр против мошенников: пользователь ведёт диалог со сценарным LLM-агентом,
его решения оценивает движок сценариев, итог попадает в профиль и лидерборд.

## Запуск

```bash
echo "DEEPSEEK_API_KEY=<ключ>" > .env
docker compose up --build
```

Миграции накатываются на старте. API — `http://localhost:8080/api`, проверка живости —
`GET http://localhost:8080/healthz` (пингует Postgres, отдаёт `503`, если база недоступна).

Обязательная переменная одна — `DEEPSEEK_API_KEY`. У `JWT_SECRET`, `DEEPSEEK_BASE_URL`
и `DEEPSEEK_MODEL` есть dev-значения по умолчанию; `JWT_SECRET` перед публичным деплоем
нужно задать своим.

## Проверка

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
```

Профиль (`POST /users`) нужен до первого чата — без него создание чата вернёт `409`.

Дальше: `POST /chats/1/messages` отвечает `202` сразу, ответ агента приходит в SSE-поток
`GET /chats/1/events` событиями `message` / `decision` / `chat` / `error`.

## Разработка

```bash
cd backend
go build ./... && go vet ./... && go test ./...
golangci-lint run ./...
sqlc generate   # после правок queries.sql
```

## Куда смотреть

| Что | Где |
|---|---|
| Цель, стек, карта доменов | `AGENT.MD` |
| Домен чатов: контракты, решения, техдолг | `backend/docs/chat.md` |
| Полный HTTP-контракт и коды ошибок | `backend/docs/chat.md`, §4 |
| Как устроены тесты и что не покрыто | `backend/docs/chat.md`, §5 |
| Сценарии обмана | `backend/scenarios/*.yaml` |
| Схема БД | `backend/migrations/migrations` |
