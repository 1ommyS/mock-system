## Быстрый старт

```bash
bun install
bun dev
```

Откройте `http://localhost:3000`.

## Настройка API

Создайте файл `.env` (или `.env.local`) в `frontend` и добавьте базовые URL:

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_MOCK_API_BASE_URL=http://localhost:8081
USER_SVC_INTERNAL_SECRET=dev-secret
```

`USER_SVC_INTERNAL_SECRET` используется серверным route для регистрации `dsl_script` ресурсов.

## Генерация OpenAPI клиента

Типы генерируются из `user-svc/internal/http/swagger/openapi.yaml`.

```bash
bun run gen:api
```

Моки:

```bash
bun run gen:mock-api
```

## Storybook

```bash
bun run storybook
```

## Тесты

```bash
bun run test
```

JUnit отчет сохраняется в `frontend/reports/junit.xml`.

## Страницы

- `/login`
- `/register`
- `/profile` (защищенная страница)
- `/mocks` (интерфейс mock-svc)
- `/dsl-runner` (генерация моков через DSL)

## Заметки по авторизации

Токены сохраняются в `localStorage` или `sessionStorage` (в зависимости от флага "Запомнить меня").
