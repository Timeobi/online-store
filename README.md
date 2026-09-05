# Go E-commerce

![CI](https://github.com/Timeobi/online-store/actions/workflows/ci.yml/badge.svg)

PET проект интернет-магазина на Go.

## Возможности

- Каталог товаров с категориями, пагинацией, фильтрацией и поиском
- Регистрация и аутентификация через JWT, роли (customer/admin)
- Корзина и оформление заказов с транзакционной атомарностью и защитой от гонки данных
- Структурированное логирование и сквозная трассировка запросов (request ID)
- Rate limiting на аутентификацию
- Docker и docker-compose для локального развёртывания
- Unit- и интеграционные тесты, CI/CD через GitHub Actions
- Интерактивная документация API (Swagger UI)

## Стек технологий

- **Язык:** Go 1.26
- **Роутер:** chi
- **База данных:** PostgreSQL, драйвер pgx
- **Кэш (подготовлен):** Redis
- **Аутентификация:** JWT (golang-jwt) + bcrypt
- **Миграции:** golang-migrate
- **Тестирование:** testify, testcontainers-go
- **Контейнеризация:** Docker, docker-compose
- **CI/CD:** GitHub Actions
- **Документация:** OpenAPI/Swagger (swaggo)

## Архитектура

Проект построен по слоистой архитектуре:

\`\`\`
HTTP-запрос → handler → service → repository → PostgreSQL
\`\`\`

- **handler** — разбор HTTP-запросов, формирование ответов
- **service** — бизнес-логика, валидация
- **repository** — единственный слой, работающий с SQL

Подробное описание каждого этапа разработки — в папке [\`docs/\`](docs/).

## Быстрый старт (Docker)

1. Скопируй \`.env.docker.example\` в \`.env.docker\`, заполни значения
2. Запусти:
   \`\`\`bash
   docker compose --env-file .env.docker up --build
   \`\`\`
3. Примени миграции:
   \`\`\`bash
   migrate -database "postgres://<DB_USER>:<DB_PASSWORD>@localhost:5432/<DB_NAME>?sslmode=disable" -path migrations up
   \`\`\`
4. API доступен на http://localhost:8080
5. Интерактивная документация: http://localhost:8080/swagger/index.html

## Запуск локально (без Docker)

Требуется: Go 1.26+, PostgreSQL 16+.

1. Скопируй \`.env.example\` в \`.env\`, заполни значения
2. Примени миграции (см. выше)
3. Запусти:
   \`\`\`bash
   go run cmd/api/main.go
   \`\`\`

## Тестирование

\`\`\`bash
# Unit-тесты
go test ./...

# Unit-тесты с детектором гонки данных
go test -race ./...

# Интеграционные тесты (требует Docker)
go test -tags=integration ./...

# Покрытие кода
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
\`\`\`

## Postman

Готовая коллекция запросов лежит в [\`docs/postman/\`](docs/postman/) — импортируй в Postman для быстрого старта.

## Продакшен-соображения (не реализовано в рамках учебного проекта)

- HTTPS должен обеспечиваться через reverse proxy (Nginx/Caddy) или балансировщик облачного провайдера
- \`DB_SSLMODE\` должен быть \`require\` или строже, а не \`disable\`
- Секреты должны храниться в secret-менеджере (Vault, AWS Secrets Manager), а не в \`.env\`-файле
- Rate limiter сейчас хранит состояние в памяти процесса — для нескольких инстансов приложения нужно вынести в Redis

## Структура проекта

\`\`\`
cmd/api/           — точка входа приложения
internal/
  config/          — загрузка конфигурации
  model/           — сущности данных
  repository/      — работа с БД
  service/         — бизнес-логика
  handler/         — HTTP-хендлеры
  middleware/      — auth, логирование, rate limiting, recovery
  logger/          — настройка структурированного логирования
migrations/        — SQL-миграции
docs/              — документация по этапам разработки, Postman-коллекция, Swagger
.github/workflows/ — CI/CD пайплайн
\`\`\`

## Автор

[Timeobi](https://github.com/Timeobi)