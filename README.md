# Balance Microservice

## Инструкция по запуску

```
make up_build
```
- поднятие и запуск docker контейнера

### Выполнить скрипт для создания таблиц
`scripts/balances.sql`

### Импортировать postman коллекцию для теста API <br> 
`/postman/Test API Collection.postman_collection.json`

#### [Комментарий]

Изначально планировал применить паттерн outbox compensating transaction, SAGA, 
с формированием ключа идемпотентности, pub/sub,
но не хватило практики для качественной реализации.
