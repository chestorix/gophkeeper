# GophKeeper - Менеджер паролей

Клиент-серверная система для безопасного хранения приватных данных с синхронизацией между устройствами.

## 🚀 Быстрый старт

### Запуск через Docker Compose (рекомендуется)

```bash
# Запуск сервера и PostgreSQL
docker-compose up -d

# Проверка статуса
docker-compose ps

# Логи сервера
docker-compose logs server

Сервер будет доступен по адресу: http://localhost:8090
Сборка клиентов для разных платформ


# Сделать скрипт исполняемым
chmod +x build.sh

# Собрать клиенты для всех платформ
./build.sh

Собранные клиенты будут в папке dist/:

    gophkeeper-client-linux - Linux 64-bit

    gophkeeper-client.exe - Windows 64-bit

    gophkeeper-client-macos-intel - MacOS Intel

    gophkeeper-client-macos-apple - MacOS Apple Silicon

📋 Функциональность
🔐 Аутентификация и безопасность

    Регистрация новых пользователей

    JWT-аутентификация с 24-часовыми токенами

    Хэширование паролей (SHA256)

    Защищенные эндпоинты

💾 Типы хранимых данных

    Логины/пароли - учетные данные для сайтов и сервисов

    Текстовые данные - произвольные текстовые заметки

    Бинарные файлы - документы, изображения, любые файлы

    Банковские карты - данные платежных карт

🔄 Синхронизация

    Автоматическая синхронизация между устройствами

    Разрешение конфликтов через версионирование

    История изменений с временными метками

🏗️ Архитектура
Серверная часть

    Язык: Go 1.21+

    Хранилище: PostgreSQL

    Аутентификация: JWT токены

    API: RESTful HTTP

    Контейнеризация: Docker

Клиентская часть

    Язык: Go 1.21+

    Интерфейс: CLI (Command Line Interface)

    Платформы: Windows, Linux, MacOS

    Локальное кэширование: Файловая система

⚙️ Конфигурация

Серверные переменные окружения

    DATABASE_URI - строка подключения к PostgreSQL

    JWT_SECRET - секретный ключ для JWT токенов

    RUN_ADDRESS - адрес и порт сервера (по умолчанию: :8090)

Клиентские переменные окружения

    SERVER_ADDRESS - адрес сервера (по умолчанию: localhost:8090)

    TOKEN - JWT токен для аутентификации

    DATA_DIR - директория для локальных данных (по умолчанию: ./data)

📡 API Endpoints
🔓 Публичные эндпоинты

    POST /api/user/register - регистрация нового пользователя

    POST /api/user/login - аутентификация пользователя

    GET /health - проверка здоровья сервера

🔐 Защищенные эндпоинты (требуют JWT токен)

    GET /api/data - список всех данных пользователя

    GET /api/data/{id} - получение данных по ID

    GET /api/data/name/{name} - получение данных по имени

    POST /api/data - сохранение новых данных

    PUT /api/data/{id} - обновление существующих данных

    DELETE /api/data/{id} - удаление данных по ID

    DELETE /api/data/name/{name} - удаление данных по имени

    POST /api/sync - синхронизация данных с сервером

💻 Использование клиента
Запуск клиента
bash

# Использование собранного бинарника
./gophkeeper-client-linux

# Или через go run
go run cmd/client/main.go

# Указать и адрес и токен
./gophkeeper-client-linux -a 192.168.1.100:8090 -t "your-jwt-token-here"

# Или через переменные
export SERVER_ADDRESS="192.168.1.100:8090"
export TOKEN="your-jwt-token-here"
./gophkeeper-client-linux

Основные команды
text

1. Register    - регистрация нового пользователя
2. Login       - вход в систему
3. Save data   - сохранение новых данных
4. List data   - просмотр списка данных
5. Get data    - получение конкретных данных
6. Update data - обновление существующих данных  
7. Delete data - удаление данных
8. Sync data   - синхронизация с сервером
9. Health check- проверка соединения с сервером
0. Exit        - выход из приложения

🔒 Безопасность
Меры защиты

    Пароли: Хэшируются с использованием SHA256

    Токены: JWT с ограниченным временем жизни (24 часа)

    Данные: Валидация принадлежности пользователю

    Аутентификация: Bearer token в заголовках

Docker Production
bash

# Сборка образа
docker build -f Dockerfile.server -t gophkeeper-server .

# Запуск
docker run -d \
  -e DATABASE_URI="postgresql://..." \
  -e JWT_SECRET="your-secret" \
  -p 8090:8090 \
  gophkeeper-server

🗄️ Структура проекта


gophkeeper/
├── cmd/
│   ├── server/          # Серверное приложение
│   └── client/          # Клиентское приложение
├── internal/
│   ├── api/            # HTTP handlers и роутинг
│   ├── service/        # Бизнес-логика
│   ├── repository/     # Работа с базой данных
│   ├── models/         # Структуры данных
│   ├── config/         # Конфигурация
│   ├── interfaces/     # Контракты приложения
│   ├── client/         # Клиентская логика
│   ├── ui/             # CLI интерфейс
│   └── errors/         # Ошибки приложения
├── dist/               # Собранные бинарники клиентов
├── docker-compose.yml  # Docker Compose для разработки
├── Dockerfile.server   # Dockerfile для сервера
├── build.sh           # Скрипт сборки клиентов
└── README.md          # Документация



