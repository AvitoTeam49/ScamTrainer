# ScamTrainer — цель и стек

Продуктовая цель и принятые решения по стеку. Границы домена чатов, контракты между
доменами и принятый технический долг — в `backend/docs/chat.md`. Запуск и сквозной
сценарий — в `README.md`.

## Цель

Тренажёр для противодействия скаму: хакатон Avito, домен чатов — их же постановка.
Система чатов с ИИ-агентами на базе DeepSeek.

Сценарий — граф обмана, внутри которого агент пытается развести пользователя, а через
tool-calls отмечает, где тот повёлся. Сценарии бывают составными, имитируя сложный обман:
сначала агент прогревает собеседника, потом уводит на фишинг и так далее. Реплики,
переходы и их веса задаёт автор сценария в YAML (`backend/scenarios`), а не код.

## Backend

- Язык — Go. Монолит: один бинарь `backend/cmd/app`, одна БД, одна папка миграций.
- БД — PostgreSQL. Обоснование выбора — `backend/docs/chat.md`, §3.
- Деплой — Docker Compose, `docker-compose.yml` в корне.
- Линтер — `backend/.golangci.yml`.

### Домены

| Домен | Пакеты |
|---|---|
| `chat` | `internal/domain/chat`, `internal/usecase/chat`, `internal/transport/rest/chat`, `internal/repository/postgres/chat` |
| `scenario` | `internal/domain/scenario`, `internal/repository/yaml/scenario`, `internal/transport/rest/scenarios` |
| `users` | `internal/domain/users`, `internal/usecase/users`, `internal/transport/rest/users`, `internal/repository/postgres/users` |
| `auth` | `internal/auth`, `internal/middleware` |

Обслуживающие пакеты: `internal/training` — сессии тренировки поверх движка сценариев,
`internal/agent` и `internal/agent/deepseek` — протокол общения с LLM, `internal/eventbus` —
in-process pub/sub для SSE, `internal/config`.

**Раскладка неоднородна.** `chat` и `users` разложены по слоям — domain / usecase /
transport / repository. `auth`, `training`, `middleware` — плоские пакеты. Это следствие
решения не переписывать чужие пакеты при слиянии веток; зафиксировано в
`backend/docs/chat.md`, §6.
