# Сервис для укорочения ссылок

---

## Структура проекта

```bash
.
├───.github
│   └───workflow
├───cmd
│   ├───migrator
│   └───url-shortener
├───config
├───domain
│   └───models
├───internal
│   ├───clients
│   │   └───sso
│   │       └───grpc
│   ├───config
│   ├───http-server
│   │   ├───handlers
│   │   │   ├───redirect
│   │   │   │   └───mocks
│   │   │   └───url
│   │   │       ├───delete
│   │   │       │   └───mocks
│   │   │       └───save
│   │   │           └───mocks
│   │   └───middleware
│   │       ├───auth
│   │       └───logger
│   ├───lib
│   │   ├───api
│   │   │   └───response
│   │   ├───jwt
│   │   ├───logger
│   │   │   ├───handlers
│   │   │   │   ├───slogdiscard
│   │   │   │   └───slogpretty
│   │   │   └───sl
│   │   └───random
│   └───storage
│       └───sqlite
├───migrations
└───storage
```


Для работы сервиса необходима авторизация в сервисе [SSO](https://github.com/dedmouze/sso)

___

## Сервис предоставляет 3 эндпоинта

### SaveURL: host/url

#### Request:
```json
{
    "url":   "url",  // required, url
    "alias": "alias" // omitemtpy
}
```

#### Response:
```json
{
    "status": "status",
    "error":  "error", // omitempty
    "alias":  "alias"
}
```

#### Возможные HTTP запросы:
```batch
curl --location 'localhost:8085/url' --header 'Content-Type: application/json' --header 'Authorization: Basic XXXXXXXXXXXX' --data '{"url":"https://yandex.ru", "alias":"ya"}'
curl --location 'localhost:8085/url' --header 'Content-Type: application/json' --header 'Authorization: Basic XXXXXXXXXXXX' --data '{"url":"https://mail.ru"}'
```

---

### GetURL: host/'alias'
#### Возможный HTTP запрос:
```batch
curl --location 'localhost:8085/ya'
```
#### Response:
```json
{
    "status": "status",
    "error":  "error" // omitempty
}
```

---

### DeleteURL: host/url/'alias'
#### Возможный HTTP запрос:
```batch
curl --location --request DELETE 'localhost:8085/url/ya' --header 'Authorization: Basic XXXXXXXXXXXX'
```
#### Response:
```json
{
    "status": "status",
    "error":  "error" // omitempty
}
```