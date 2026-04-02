# ozon-grapql-service
Тестовое задание в Ozon Банк.

Система для добавления и чтения постов и комментариев с использованием GraphQL, аналогичная комментариям к постам на популярных платформах, таких как Хабр или Reddit.
### Установка
```
git clone github.com/solumD/ozon-grapql-service
cd ozon-grapql-service
go mod tidy
```

### Важные моменты реализации


### Запуск
Скопировать содержимое файла .env.example в файл .env, поменять значения при необходимости. Для запуска должен быть установлен Docker Compose. В терминале выполнить команду:
```
make run
```
Будут подняты контейнеры с БД Postrgres и приложением, а также накатятся миграции с помощью утилиты goose.

### Тестирование
Unit-тесты написаны для базового функционала слоев usecase и delivery.

Для запуска тестов в терминале выполнить команду: 
```
make test
```

### Описание основных операций системы


### Структура проекта
```
.
├── bin
├── cmd
│   └── app
├── config
├── internal
│   ├── app
│   ├── broker
│   │   └── in_memory
│   ├── core_errors
│   ├── delivery
│   │   ├── graphql
│   │   │   ├── generated
│   │   │   ├── mock
│   │   │   ├── schema
│   │   │   └── tests
│   │   └── router
│   ├── model
│   ├── repository
│   │   ├── in_memory
│   │   └── postgres
│   ├── usecase
│   │   ├── mock
│   │   └── tests
│   └── utils
├── Makefile
├── migrations
├── pkg
│   ├── http_server
│   ├── logger
│   └── postgres
├── docker-compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```
