# Сервис Сокращения Ссылок (Backend)

Современный, масштабируемый сервис сокращения URL-адресов на Go с использованием принципов чистой архитектуры.

## Возможности

- 🔗 **Сокращение URL**: Создание коротких, запоминающихся ссылок
- 🎯 **Пользовательские короткие коды**: Возможность использовать собственные псевдонимы
- 📊 **Аналитика**: Отслеживание кликов и статистика
- 🔐 **Аутентификация**: JWT-аутентификация пользователей
- ⏱️ **Срок действия ссылок**: Установка даты истечения для временных ссылок
- 🚦 **Ограничение скорости**: Защита от злоупотреблений через Redis
- 📱 **RESTful API**: Чистый, интуитивный дизайн API
- 📚 **Документация API**: Интерактивная Swagger-документация (/swagger/index.html)
- 🐳 **Контейнеризация**: Docker и Docker Compose

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
└── migrations/         # Миграции базы данных
```

## Тестирование

### Типы тестов
- **Unit-тесты**: Testify и моки
- **Статический анализ (SAST)** (сейчас без поддержки локального запуска, только в CI):
  - **Gosec**: Специализированный для Golang, проверяет hardcoded credentials, криптографию и т.д.
  - **Semgrep**: Универсальный инструмент с правилами для Golang
  - **CodeQL**: Семантический анализ от GitHub
- **Динамический анализ (DAST)**:
  - **OWASP ZAP**: Сканирование запущенного приложения на уязвимости (XSS, SQL-инъекции и т.д.)

### Запуск тестов
```bash
# Unit-тесты
go test ./...
```

## Установка и запуск

### Требования
- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker и Docker Compose (опционально)

### Локальный запуск
```bash
# Клонировать репозиторий
git clone https://github.com/raison-collab/LinkShortenerBackend.git
cd LinkShortenerBackend

# Настроить конфигурацию
cp env.example .env
# Отредактировать .env

# Установить зависимости
go mod download

# Запустить миграции
# psql -U postgres -d link_shortener < migrations/00*.sql

# Запустить приложение
go run cmd/api/main.go
```

### Docker Compose
```bash
# Запуск всех сервисов
docker-compose up -d

# Остановка
docker-compose down
```

### Конфигурация (основные параметры)
| Переменная | Описание | По умолчанию |
|----------|-------------|---------|
| `APP_NAME` | Название приложения | `link-shortener` |
| `APP_PORT` | Порт приложения | `8080` |
| `APP_ENV` | Окружение (development/production) | `development` |
| `APP_DEBUG` | Режим отладки | `true` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_NAME` | Имя базы данных | `link_shortener` |
| `DB_USER` | Пользователь БД | `postgres` |
| `DB_PASSWORD` | Пароль БД | `postgres` |
| `DB_SSL_MODE` | Режим SSL для PostgreSQL | `disable` |
| `REDIS_HOST` | Хост Redis | `localhost` |
| `REDIS_PORT` | Порт Redis | `6379` |
| `REDIS_PASSWORD` | Пароль Redis | `` |
| `REDIS_DB` | Номер БД Redis | `0` |
| `JWT_SECRET` | Секретный ключ JWT | `your-secret-key-here` |
| `JWT_EXPIRE_HOURS` | Время жизни JWT токена (часы) | `24` |
| `BASE_URL` | Базовый URL для коротких ссылок | `http://localhost:8080` |
| `API_HOST` | Хост для Swagger-документации | `localhost:8080` |
| `SHORT_URL_LENGTH` | Длина генерируемых коротких кодов | `6` |
| `CORS_ALLOW_ORIGINS` | Разрешенные источники для CORS | `http://localhost:3000,https://app.example.com` |
| `CORS_ALLOW_METHODS` | Разрешенные методы для CORS | `GET,POST,PUT,DELETE,OPTIONS,PATCH` |
| `CORS_ALLOW_HEADERS` | Разрешенные заголовки для CORS | `Origin,Content-Type,Accept,Authorization` |
| `RATE_LIMIT_REQUESTS` | Количество запросов для rate limit | `100` |
| `RATE_LIMIT_WINDOW_MINUTES` | Окно времени для rate limit (минуты) | `1` |
| `LOG_LEVEL` | Уровень логирования | `debug` |
| `LOG_OUTPUT` | Вывод логов (console, file, both) | `console` |
| `DOCKER_NETWORK_PROXY` | Сеть Docker для проксирования | `proxy` |

## API Эндпоинты

### Публичные эндпоинты
- `GET /health` - Проверка состояния сервиса
- `GET /swagger/*` - Документация API (Swagger UI)
- `GET /:code` - Переход по короткой ссылке
- `POST /api/v1/auth/register` - Регистрация пользователя
- `POST /api/v1/auth/login` - Вход в систему

### Защищенные эндпоинты (требуют JWT)
- **Пользователи**:
  - `GET /api/v1/users/me` - Профиль текущего пользователя
  - `PUT /api/v1/users/me` - Обновить профиль
  - `PUT /api/v1/users/me/password` - Изменить пароль
  - `GET /api/v1/users/me/stats` - Статистика пользователя

- **Ссылки**:
  - `POST /api/v1/links` - Создать короткую ссылку
  - `GET /api/v1/links` - Список ссылок пользователя
  - `GET /api/v1/links/:id` - Детали ссылки
  - `PUT /api/v1/links/:id` - Обновить ссылку
  - `DELETE /api/v1/links/:id` - Удалить ссылку
  - `GET /api/v1/links/:id/stats` - Статистика ссылки

> Полная документация API доступна по адресу `/swagger/index.html` после запуска сервиса.
