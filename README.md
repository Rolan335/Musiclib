
## API для выборки музыкальных песен

Стек: gin, pgx, миграции goose, кодогенерация через oapi-codegen

## Настройка

Необходим Docker
```bash
  git clone https://github.com/Rolan335/Musiclib
  cd Musiclib
  docker compose up
```
Сервис будет локально доступен по адресу localhost:8080

---

1. Реализованы методы
- Получение данных библиотеки с фильтрацией по всем полям и пагинацией
- Получение текста песни с пагинацией по куплетам
- Удаление песни
- Изменение данных песни
- Добавление новой песни в формате JSON с обращением к внешнему API. 

2. При добавлении сделать запрос в АПИ, описанного сваггером. Апи,
описанный сваггером. (Если песня не найдена во внешнем API, сервис возвращает ошибку 404)

3. Данных хранятся в Postgres. Миграция настраивается с помощью .env файла

4. Код покрыт Debug и Info логами

5. Добавлен .env.example и .env.docker который используется для конфигурации Docker

6. Документация к API находится по адресу localhost:8080/swagger
