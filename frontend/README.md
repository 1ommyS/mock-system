## Быстрый старт

```bash
bun install
bun dev
```

Откройте `http://localhost:3000`.

## Настройка API

Создайте файл `.env.local` в `frontend` и добавьте базовый URL сервиса:

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_MOCK_API_BASE_URL=http://localhost:8081
```

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

## Страницы

- `/login`
- `/register`
- `/profile` (защищенная страница)
- `/mocks` (интерфейс mock-svc)

## Заметки по авторизации

Токены сохраняются в `localStorage` или `sessionStorage` (в зависимости от флага "Запомнить меня").
