# ozon-graphql-service
Тестовое задание в Ozon Банк.

Система для добавления и чтения постов и комментариев с использованием GraphQL, аналогичная комментариям к постам на популярных платформах, таких как Хабр или Reddit.
### Установка
```
git clone github.com/solumD/ozon-graphql-service
cd ozon-graphql-service
go mod tidy
```

### Важные моменты реализации
- Для получения комментариев выбрана cursor-based пагинация. Курсор строится на основе пары `(createdAt, id)`, что позволяет стабильно и предсказуемо листать комментарии без зависимости от offset/limit.
- Комментарии поддерживают неограниченную вложенность. В in-memory реализации они группируются по бакетам вида `postID:parentID`
- Поддерживаются два режима хранения данных: in-memory** и **PostgreSQL. Режим выбирается через переменную окружения `STORAGE_TYPE`.
- Для GraphQL subscriptions реализован отдельный in-memory broker комментариев, через который новые комментарии доставляются подписанным клиентам асинхронно.
- Для слоёв `usecase` и `delivery` написаны unit-тесты, моки сгенерированы с использованием `minimock`.
- Миграции накатываются с помощью утилиты `goose`.


### Запуск
Скопировать содержимое файла .env.example в файл .env, поменять значения при необходимости. Для запуска должен быть установлен Docker Compose. В терминале выполнить команду:
```
make run
```
Будут подняты контейнеры с PostrgreSQL и приложением, а также накатятся миграции с помощью утилиты goose.

### Тестирование
Unit-тесты написаны для базового функционала слоев usecase и delivery.

Для запуска тестов в терминале выполнить команду: 
```
make test
```

### Описание основных операций системы
Ниже приведён последовательный сценарий, который можно выполнять шаг за шагом для ручной проверки сервиса.

#### 1. Создание поста
```graphql
mutation {
  createPost(
    userUUID: "user-1"
    title: "First post"
    content: "Hello, world!"
    commentsEnabled: true
  ) {
    id
    userUUID
    title
    content
    commentsEnabled
    createdAt
  }
}
```

Пример ответа:
```json
{
  "data": {
    "createPost": {
      "id": 1,
      "userUUID": "user-1",
      "title": "First post",
      "content": "Hello, world!",
      "commentsEnabled": true,
      "createdAt": "2026-04-02T17:00:00Z"
    }
  }
}
```

#### 2. Получение списка постов
```graphql
query {
  posts {
    id
    userUUID
    title
    content
    commentsEnabled
    createdAt
  }
}
```

Пример ответа:
```json
{
  "data": {
    "posts": [
      {
        "id": 1,
        "userUUID": "user-1",
        "title": "First post",
        "content": "Hello, world!",
        "commentsEnabled": true,
        "createdAt": "2026-04-02T17:00:00Z"
      }
    ]
  }
}
```

#### 3. Получение конкретного поста по `id`
```graphql
query {
  post(id: 1) {
    id
    userUUID
    title
    content
    commentsEnabled
    createdAt
  }
}
```

Пример ответа:
```json
{
  "data": {
    "post": {
      "id": 1,
      "userUUID": "user-1",
      "title": "First post",
      "content": "Hello, world!",
      "commentsEnabled": true,
      "createdAt": "2026-04-02T17:00:00Z"
    }
  }
}
```

#### 4. Создание корневого комментария к посту
```graphql
mutation {
  createComment(
    userUUID: "user-2"
    postID: 1
    content: "Root comment"
  ) {
    id
    userUUID
    postID
    parentID
    hasReplies
    content
    createdAt
  }
}
```

Пример ответа:
```json
{
  "data": {
    "createComment": {
      "id": 1,
      "userUUID": "user-2",
      "postID": 1,
      "parentID": null,
      "hasReplies": false,
      "content": "Root comment",
      "createdAt": "2026-04-02T17:01:00Z"
    }
  }
}
```

#### 5. Создание вложенного комментария
```graphql
mutation {
  createComment(
    userUUID: "user-3"
    postID: 1
    parentID: 1
    content: "Reply to root comment"
  ) {
    id
    userUUID
    postID
    parentID
    hasReplies
    content
    createdAt
  }
}
```

Пример ответа:
```json
{
  "data": {
    "createComment": {
      "id": 2,
      "userUUID": "user-3",
      "postID": 1,
      "parentID": 1,
      "hasReplies": false,
      "content": "Reply to root comment",
      "createdAt": "2026-04-02T17:02:00Z"
    }
  }
}
```

#### 6. Получение корневых комментариев к посту
```graphql
query {
  comments(postID: 1, parentID: null, first: 20, after: null) {
    edges {
      cursor
      node {
        id
        userUUID
        postID
        parentID
        hasReplies
        content
        createdAt
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

Пример ответа:
```json
{
  "data": {
    "comments": {
      "edges": [
        {
          "cursor": "MTcxMjA3NzI2MDAwMDAwMDAwMDox",
          "node": {
            "id": 1,
            "userUUID": "user-2",
            "postID": 1,
            "parentID": null,
            "hasReplies": true,
            "content": "Root comment",
            "createdAt": "2026-04-02T17:01:00Z"
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "endCursor": "MTcxMjA3NzI2MDAwMDAwMDAwMDox"
      }
    }
  }
}
```

#### 7. Получение дочерних комментариев для конкретного комментария
```graphql
query {
  comments(postID: 1, parentID: 1, first: 20, after: null) {
    edges {
      cursor
      node {
        id
        userUUID
        postID
        parentID
        hasReplies
        content
        createdAt
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

Пример ответа:
```json
{
  "data": {
    "comments": {
      "edges": [
        {
          "cursor": "MTcxMjA3NzMyMDAwMDAwMDAwMDoy",
          "node": {
            "id": 2,
            "userUUID": "user-3",
            "postID": 1,
            "parentID": 1,
            "hasReplies": false,
            "content": "Reply to root comment",
            "createdAt": "2026-04-02T17:02:00Z"
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "endCursor": "MTcxMjA3NzMyMDAwMDAwMDAwMDoy"
      }
    }
  }
}
```

#### 8. Отключение комментариев для поста
```graphql
mutation {
  changePostCommentsAvailability(postID: 1, enabled: false) {
    id
    userUUID
    title
    content
    commentsEnabled
    createdAt
  }
}
```

Пример ответа:
```json
{
  "data": {
    "changePostCommentsAvailability": {
      "id": 1,
      "userUUID": "user-1",
      "title": "First post",
      "content": "Hello, world!",
      "commentsEnabled": false,
      "createdAt": "2026-04-02T17:00:00Z"
    }
  }
}
```

#### 9. Подписка на новые комментарии к посту
Этот шаг удобно выполнять в отдельной вкладке GraphQL Playground перед созданием нового комментария к посту.

```graphql
subscription {
  commentAdded(postID: 1) {
    id
    userUUID
    postID
    parentID
    hasReplies
    content
    createdAt
  }
}
```

Пример события:
```json
{
  "data": {
    "commentAdded": {
      "id": 3,
      "userUUID": "user-4",
      "postID": 1,
      "parentID": null,
      "hasReplies": false,
      "content": "New live comment",
      "createdAt": "2026-04-02T17:03:00Z"
    }
  }
}
```


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
