# Сервис Сокращения Ссылок (Backend)

Современный, масштабируемый сервис сокращения URL-адресов, построенный на Go с использованием принципов чистой архитектуры.

## Возможности

- 🔗 **Сокращение URL**: Создание коротких, запоминающихся ссылок из длинных URL
- 🎯 **Пользовательские короткие коды**: Возможность использовать собственные псевдонимы для ссылок
- 📊 **Аналитика**: Детальное отслеживание кликов и статистика
- 🔐 **Аутентификация**: Аутентификация пользователей на основе JWT
- ⏱️ **Срок действия ссылок**: Установка даты истечения для временных ссылок
- 🚦 **Ограничение скорости**: Защита от злоупотреблений с помощью Redis
- 📱 **RESTful API**: Чистый, интуитивно понятный дизайн API
- 📚 **Документация API**: Интерактивная документация Swagger
- 🐳 **Контейнеризация**: Поддержка Docker и Docker Compose
- ✅ **Тестирование**: Комплексные unit-тесты с использованием моков

## Технологический стек

- **Язык**: Go 1.21+
- **Веб-фреймворк**: Gin
- **База данных**: PostgreSQL
- **Кэш**: Redis
- **Аутентификация**: JWT
- **Документация API**: Swagger (Swaggo)
- **Тестирование**: Testify + Mocks
- **CI/CD**: GitHub Actions

## Архитектура

Проект следует принципам чистой архитектуры:

```
├── cmd/api/              # Точки входа приложения
├── internal/
│   ├── domain/          # Бизнес-сущности и интерфейсы репозиториев
│   │   ├── entity/      # Доменные модели
│   │   └── repository/  # Интерфейсы репозиториев
│   ├── usecase/         # Бизнес-логика
│   ├── delivery/        # Внешние интерфейсы (HTTP обработчики)
│   │   └── http/
│   │       ├── handler/    # HTTP обработчики
│   │       ├── middleware/ # HTTP middleware
│   │       ├── dto/       # Объекты передачи данных
│   │       └── router/    # Определение маршрутов
│   └── infrastructure/  # Внешние реализации
│       ├── config/      # Конфигурация
│       ├── database/    # Подключения к БД
│       └── repository/  # Реализации репозиториев
├── pkg/                 # Общие пакеты
│   ├── logger/         # Утилиты логирования
│   ├── utils/          # Общие утилиты
│   └── validator/      # Помощники валидации
├── migrations/         # Миграции базы данных
├── tests/             # Тестовые файлы
└── docs/              # Документация
```

## API Эндпоинты

### Публичные эндпоинты

- `GET /health` - Проверка состояния сервиса
- `GET /swagger/*` - Документация API
- `GET /:code` - Переход по короткой ссылке
- `POST /api/v1/auth/register` - Регистрация пользователя
- `POST /api/v1/auth/login` - Вход в систему

### Защищенные эндпоинты (требуют аутентификации)

#### Управление пользователями
- `GET /api/v1/users/me` - Получить профиль текущего пользователя
- `PUT /api/v1/users/me` - Обновить профиль пользователя
- `PUT /api/v1/users/me/password` - Изменить пароль
- `GET /api/v1/users/me/stats` - Получить статистику пользователя

#### Управление ссылками
- `POST /api/v1/links` - Создать короткую ссылку
- `GET /api/v1/links` - Получить список ссылок пользователя
- `GET /api/v1/links/:id` - Получить детали ссылки
- `PUT /api/v1/links/:id` - Обновить ссылку
- `DELETE /api/v1/links/:id` - Удалить ссылку
- `GET /api/v1/links/:id/stats` - Получить статистику ссылки

## Справочник REST API

### Стандартная схема ошибки

```json
{
  "error": "Сообщение об ошибке"
}
```

| HTTP-код | Когда выдаётся |
|----------|----------------|
| 400      | Ошибка валидации входных данных |
| 401      | Отсутствует или некорректен JWT-токен |
| 403      | Нет доступа к ресурсу (чужая ссылка) |
| 404      | Ресурс не найден |
| 409      | Ссылка просрочена |
| 500      | Внутренняя ошибка сервера |

---

### 1. Управление ссылками (требует Bearer-JWT)

#### 1.1 Создать ссылку

```
POST /api/v1/links
Content-Type: application/json
```

Тело запроса:

| Поле        | Тип        | Обязательно | Описание |
|-------------|-----------|-------------|-----------|
| `url`       | string (URL) | ✔ | Оригинальный адрес |
| `custom_code` | string | – | Пользовательский код (если нужен) |
| `expires_at` | RFC3339-datetime | – | Дата/время окончания действия (UTC или со смещением) |

Ответ 201 `LinkResponse`:

```json
{
  "id": 2,
  "short_code": "K6tD5A",
  "short_url": "http://localhost:8080/K6tD5A",
  "original_url": "https://example.com",
  "clicks": 0,
  "expires_at": "2025-12-31T23:59:59Z",
  "created_at": "2025-06-24T00:10:18Z",
  "updated_at": "2025-06-24T00:10:18Z"
}
```

Ошибки: 400 (невалидный URL/код занят/дата в прошлом).

#### 1.2 Список ссылок пользователя

```
GET /api/v1/links?page=1&limit=20
```

Ответ 200: массив `LinkResponse`.

Пагинация: `page` (>=1), `limit` (1-100). Ошибка параметров → 400.

#### 1.3 Получить ссылку по ID

```
GET /api/v1/links/{id}
```

Ответ 200 `LinkResponse`.

Ошибки: 403 (чужая), 404 (не найдена).

#### 1.4 Обновить ссылку

```
PUT /api/v1/links/{id}
Content-Type: application/json
```

Тело:

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|-----------|
| `expires_at` | RFC3339-datetime | – | Новая дата окончания |

Ответ 200 `{ "message": "Link updated successfully" }`

Ошибки: 400 (дата в прошлом), 403, 404.

#### 1.5 Удалить ссылку

```
DELETE /api/v1/links/{id}
```

Ответ 204 без тела.

Ошибки: 403, 404.

#### 1.6 Статистика ссылки

```
GET /api/v1/links/{id}/stats?from=2025-01-01T00:00:00Z&to=2025-12-31T23:59:59Z
```

Ответ 200 `LinkStatsResponse`.

Ошибки: 403, 404.

---

### 2. Редирект по короткому коду (публичный)

* `GET /{code}`
* `GET /api/v1/{code}` (удобный алиас)

Ответы:

| Код | Что происходит |
|-----|----------------|
| 302 | Перенаправление на `original_url` |
| 404 | Ссылка не найдена |
| 409 | Срок действия истёк |

Записывает клик и увеличивает счётчик.

---

### 3. Пользователь (JWT)

| Метод | URL | Описание |
|-------|-----|----------|
| GET | `/api/v1/users/me` | Профиль |
| PUT | `/api/v1/users/me` | Изменить email |
| PUT | `/api/v1/users/me/password` | Сменить пароль |
| GET | `/api/v1/users/me/stats` | Общая статистика ссылок |

Форматы запросов/ответов см. DTO в `internal/delivery/http/dto`.

---

### 4. Аутентификация (публично)

#### 4.1 Регистрация

```
POST /api/v1/auth/register
Content-Type: application/json

{ "email": "user@example.com", "password": "StrongP@ss1" }
```

Ответ 201: пользователь + JWT токен.

Ошибки: 400 (невалидные поля), 409 (email занят).

#### 4.2 Логин

```
POST /api/v1/auth/login
Content-Type: application/json

{ "email": "user@example.com", "password": "StrongP@ss1" }
```

Ответ 200: пользователь + JWT токен.

Ошибки: 400 (невалидные поля), 401 (неверные учётные данные).

---

> Полный OpenAPI-спек доступен по `/swagger/index.html` после запуска сервиса.

## Начало работы

### Требования

- Go 1.21 или выше
- PostgreSQL 15+
- Redis 7+
- Docker и Docker Compose (опционально)

### Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/raison-collab/LinkShorternetBackend.git
cd LinkShorternetBackend
```

2. Скопируйте файл конфигурации:
```bash
cp env.example .env
```

3. Обновите файл `.env` с вашими настройками.

### Запуск с Docker Compose

Самый простой способ запустить приложение:

```bash
docker-compose up -d
```

Это запустит:
- База данных PostgreSQL
- Кэш Redis
- Приложение

API будет доступен по адресу `http://localhost:8080`

### Локальный запуск

1. Установите зависимости:
```bash
go mod download
```

2. Выполните миграции базы данных:
```bash
# Убедитесь, что PostgreSQL запущен
psql -U postgres -d link_shortener < migrations/001_create_users_table.sql
psql -U postgres -d link_shortener < migrations/002_create_links_table.sql
psql -U postgres -d link_shortener < migrations/003_create_link_clicks_table.sql
```

3. Сгенерируйте документацию Swagger:
```bash
swag init -g cmd/api/main.go
```

4. Запустите приложение:
```bash
go run cmd/api/main.go
```

## Разработка

### Запуск тестов

```bash
# Запустить все тесты
go test ./...

# Запустить тесты с покрытием
go test -v -race -coverprofile=coverage.out ./...

# Просмотреть покрытие в браузере
go tool cover -html=coverage.out
```

### Принципы структуры проекта

- **Слой Domain**: Содержит бизнес-сущности и интерфейсы репозиториев. Без внешних зависимостей.
- **Слой Use Case**: Содержит бизнес-логику. Зависит только от слоя domain.
- **Слой Delivery**: Содержит HTTP обработчики, middleware и DTO. Зависит от слоя use case.
- **Слой Infrastructure**: Содержит внешние реализации (база данных, кэш и т.д.). Реализует интерфейсы domain.

### Добавление новых функций

1. Определите доменные сущности в `internal/domain/entity`
2. Создайте интерфейсы репозиториев в `internal/domain/repository`
3. Реализуйте бизнес-логику в `internal/usecase`
4. Создайте реализации репозиториев в `internal/infrastructure/repository`
5. Добавьте HTTP обработчики в `internal/delivery/http/handler`
6. Обновите маршруты в `internal/delivery/http/router`
7. Напишите тесты для каждого слоя

## Конфигурация

Конфигурация управляется через переменные окружения:

| Переменная | Описание | По умолчанию |
|----------|-------------|---------|
| `APP_PORT` | Порт приложения | `8080` |
| `APP_ENV` | Окружение (development/production) | `development` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_NAME` | Имя базы данных | `link_shortener` |
| `DB_USER` | Пользователь БД | `postgres` |
| `DB_PASSWORD` | Пароль БД | `postgres` |
| `REDIS_HOST` | Хост Redis | `localhost` |
| `REDIS_PORT` | Порт Redis | `6379` |
| `JWT_SECRET` | Секретный ключ JWT | `your-secret-key-here` |
| `JWT_EXPIRE_HOURS` | Время жизни JWT токена (часы) | `24` |
| `BASE_URL` | Базовый URL для коротких ссылок | `http://localhost:8080` |
| `SHORT_URL_LENGTH` | Длина генерируемых коротких кодов | `6` |
| `RATE_LIMIT_REQUESTS` | Количество запросов для rate limit | `100` |
| `RATE_LIMIT_WINDOW_MINUTES` | Окно времени для rate limit (минуты) | `1` |
| `LOG_LEVEL` | Уровень логирования | `debug` |

## Документация API

После запуска приложения документация Swagger доступна по адресу:
```
http://localhost:8080/swagger/index.html
```

## Развертывание

### Использование Docker

Сборка Docker образа:
```bash
docker build -t link-shortener .
```

Запуск контейнера:
```bash
docker run -p 8080:8080 --env-file .env link-shortener
```

### Ручное развертывание

1. Соберите бинарный файл:
```bash
CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go
```

2. Скопируйте бинарный файл и миграции на сервер
3. Настройте PostgreSQL и Redis
4. Настройте переменные окружения
5. Запустите приложение

## Использование Makefile

Проект включает Makefile для упрощения общих задач:

```bash
# Сборка приложения
make build

# Запуск приложения
make run

# Запуск тестов
make test

# Запуск с Docker Compose
make docker-run

# Остановка Docker контейнеров
make docker-stop

# Применение миграций
make migrate-up

# Полная настройка среды разработки
make dev-setup
```
