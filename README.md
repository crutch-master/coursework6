# Платформа публикации и рецензирования статей

Веб-приложение для публикации научных статей в формате [Typst](https://typst.app). Пользователи могут регистрироваться, загружать `.typ` исходники статей и писать рецензии. Фоновый воркер автоматически компилирует исходники в PDF.

## Архитектура

- **web** - Go-сервис, HTTP-сервер на порту 9090. Серверный рендеринг HTML-шаблонов, JWT-авторизация, загрузка и скачивание документов через S3.
- **worker** - Rust-сервис. Поллит базу данных на предмет статей со статусом `pending`, компилирует Typst в PDF с помощью библиотеки typst, загружает результат в S3 и обновляет статус статьи.
- **migrator** - контейнер с [goose](https://github.com/pressly/goose). Применяет SQL-миграции и завершается.
- **postgres** - база данных PostgreSQL 18.
- **seaweedfs** - S3-совместимое объектное хранилище. Хранит исходники и скомпилированные PDF.

## Структура проекта

```
/
├── compose.yaml            # Docker Compose конфигурация
├── .env                    # Переменные окружения
├── Makefile                # make vendor - скачивание фронтенд-зависимостей
├── migrations/             # SQL-миграции (goose)
├── migrator/               # Контейнер для миграции
├── seaweedfs/
│   └── s3.json             # Конфигурация S3-гейтвея SeaweedFS
├── web/
│   ├── cmd/web/main.go     # Точка входа
│   ├── internal/
│   │   ├── auth/           # JWT-утилиты
│   │   ├── handler/        # HTTP-обработчики (index, login, register, submit, article, review, profile)
│   │   ├── middleware/     # Auth-мидлвэр
│   │   ├── model/          # Модели данных
│   │   ├── repository/     # Репозитории БД (user, article, review)
│   │   ├── s3/             # S3-клиент
│   │   └── wiring/         # Сборка приложения (DI)
│   ├── static/             # Статика (htmx, Pico CSS)
│   ├── templates/          # HTML-шаблоны
│   ├── Dockerfile
│   └── go.mod & go.sum
└── worker/
    ├── src/
    │   ├── main.rs         # Точка входа, цикл поллинга
    │   ├── compile.rs      # Typst -> PDF компиляция
    │   ├── db.rs           # Запросы к БД
    │   └── s3.rs           # S3-клиент (upload/download)
    ├── Dockerfile
    └── Cargo.toml & Cargo.lock
```

## Запуск

1. В директории `seaweedfs` скопируйте `s3.example.json` в `s3.json` и отредактируйте по вкусу.

2. Создайте файл `.env` на основе примера ниже с актуальными значениями из `seaweedfs/s3.json`:

```env
POSTGRES_PASSWORD=pass
POSTGRES_USER=user
POSTGRES_DB=db

S3_REGION=us-east-1
S3_ACCESS_KEY=admin
S3_SECRET_KEY=adminadmin
S3_BUCKET=articles

JWT_SECRET=changeme
```

3. Скачайте фронтенд-зависимости (htmx, Pico CSS):

```bash
make vendor
```

4. Запустите все сервисы:

```bash
docker compose up --build
```

После запуска веб-приложение будет доступно на [http://localhost:9090](http://localhost:9090).

Порядок запуска: PostgreSQL (healthcheck) -> migrator (применяет миграции и завершается) -> web + worker.
