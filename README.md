# Go E-commerce

[![CI](https://github.com/Timeobi/online-store/actions/workflows/ci.yml/badge.svg)](https://github.com/Timeobi/online-store/actions/workflows/ci.yml)

PET проект интернет-магазина на Go.

## Стек
- Go 1.26.3
- PostgreSQL
- Redis
- Docker / docker-compose
- JWT-аутентификация

## Запуск
\`\`\`bash
go run cmd/api/main.go
\`\`\`

## Структура проекта
См. \`docs/\` для подробностей по каждому этапу разработки.

## Запуск через Docker (рекомендуется)
1. Скопируй `.env.docker.example` в `.env.docker` и заполни своими значениями
2. Запусти:
   \`\`\`bash
   docker compose --env-file .env.docker up --build
   \`\`\`
3. Примени миграции (один раз, при первом запуске):
   \`\`\`bash
   migrate -database "postgres://<DB_USER>:<DB_PASSWORD>@localhost:5432/<DB_NAME>?sslmode=disable" -path migrations up
   \`\`\`
4. API доступен на http://localhost:8080

## Запуск локально (без Docker)

См. раздел "Установка" — требуется локально установленный PostgreSQL 16+ и Go 1.22+.