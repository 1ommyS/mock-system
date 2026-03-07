# DIPLOM 🎓

Репозиторий дипломного проекта: система для управления HTTP-моками, правами доступа и генерацией производных моков через DSL.

Если коротко:
- `user-svc` отвечает за регистрацию/логин и ACL 🔐
- `mock-svc` хранит моки и запускает генерации 🧩
- `dsl-runner-svc` исполняет DSL и возвращает результат ⚙️
- `frontend` даёт веб-интерфейс ко всему этому 🖥️

## 📦 Что внутри

- `user-svc` — сервис аутентификации и авторизации (RBAC + grants) 🔐
- `mock-svc` — CRUD моков, семейства моков, preview/apply генераций 🧪
- `dsl-runner-svc` — воркеры Kafka + выполнение DSL ⚙️
- `frontend` — Next.js интерфейс (`/mocks`, `/dsl-runner`, `/profile` и т.д.) 🌐
- `devops` — docker compose для Postgres и Redpanda 🐳
- `chapter3.md` — текст главы по практической части 📘

## 🚀 Локальный запуск

### 1. Что нужно заранее

- Docker + Docker Compose 🐳
- Go `1.24.x` 🐹
- Bun (для фронтенда) 🥟
- Task (`go-task`) — удобно, но не обязательно 🛠️
- `goose` — если используете задачи миграций через `task` 🪿

### 2. Поднять инфраструктуру

Из корня проекта:

```bash
docker compose -f devops/docker-compose.yml up -d
```

Поднимутся:
- Postgres для `user-svc` на `localhost:5432` 🗄️
- Postgres для `mock-svc` на `localhost:6432` 🗄️
- Postgres для `dsl-runner-svc` на `localhost:7432` 🗄️
- Redpanda (Kafka API) на `localhost:29092` 📨

### 3. Применить миграции

```bash
task -d user-svc migrate:up
task -d mock-svc migrate:up
task -d dsl-runner-svc migrate:up
```

### 4. Запустить сервисы

Вариант A (одной командой):

```bash
task dev
```

Вариант B (по отдельности, в разных терминалах):

```bash
task -d user-svc run
task -d mock-svc run
task -d dsl-runner-svc run
task frontend
```

## 🌍 Адреса по умолчанию

- Frontend: `http://localhost:3000` 🖥️
- user-svc: `http://localhost:8080` 🔐
- mock-svc: `http://localhost:8081` 🧩
- dsl-runner-svc: `http://localhost:8090` ⚙️

Swagger:
- `http://localhost:8080/swagger` 📄
- `http://localhost:8081/swagger` 📄
- `http://localhost:8090/swagger` 📄

## 🔧 Переменные окружения для frontend

В `frontend/.env.local`:

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_MOCK_API_BASE_URL=http://localhost:8081
USER_SVC_INTERNAL_SECRET=dev-secret
```

## ✅ Как быстро проверить, что всё работает

1. Открыть `http://localhost:3000`.
2. Зарегистрировать пользователя и войти.
3. Перейти в `/mocks`, создать базовый мок.
4. Перейти в `/dsl-runner`.
5. Вставить DSL-скрипт, выбрать базовый мок.
6. Нажать `Предпросмотр`, затем `Применить`.
7. Проверить, что в генерации статус дошёл до `DONE`, и появились новые моки.

## 🧠 DSL

Документация по текстовому DSL:
- [Спецификация DSL](dsl-runner-svc/docs/dsl-language.md)
- [Описание и список примеров DSL](dsl-runner-svc/examples/dsl/README.md)

Примеры:
- `dsl-runner-svc/examples/dsl/basic-repeat.dsl`
- `dsl-runner-svc/examples/dsl/for-loop.dsl`
- `dsl-runner-svc/examples/dsl/if-and-assert.dsl`
- `dsl-runner-svc/examples/dsl/closest-store.dsl`

## 🧪 Тесты

Backend:

```bash
task -d user-svc test
task -d mock-svc test
task -d dsl-runner-svc test
```

Frontend:

```bash
cd frontend
bun run test
```

## ℹ️ Полезно знать

- Конфиги сервисов лежат в `user-svc/config.yaml`, `mock-svc/config.yaml`, `dsl-runner-svc/config.yaml`.
- Везде уже проставлены локальные дефолты для `dev-secret` и локальных портов.
- JSON DSL по-прежнему поддерживается, но основной рабочий вариант теперь — текстовый DSL.
