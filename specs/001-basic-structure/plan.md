# План 001: базовая структура

Основание: `spec.md`. Задача: каркас проекта на Go и Fiber, список методов API, заглушки в сервисах и Swagger. Бизнес-логики, БД, JWT и работы с Bitbucket на этом шаге нет, они приходят отдельными спеками. Расписание тасок строится по `reference/0004-schedule-work-window-instead-of-cron.md` (окно и число коммитов, без cron)

## Структура

```
Footprints/
  main.go
  go.mod
  docs/swagger.json        (генерируется swag)
  router/
    router.go
  services/
    auth/auth.go
    tasks/tasks.go
  models/
    db/.gitkeep
    api/.gitkeep
  domains/.gitkeep
```

Модуль называется `footprints`. Каталоги лежат в корне, без `cmd/` и `internal/`: бинарник один, прятать нечего

## Слои

- `main.go`: создаёт `fiber.App`, вызывает `router.Setup(app, swaggerJSON)`, запускает `app.Listen(":8080")`. Порт зашит константой, конфиг появится, когда понадобится. Здесь же общие swag-аннотации API, `//go:embed docs/swagger.json` и строка `//go:generate`
- `router/`: подключает Swagger UI, монтирует группу `/api/v1` и раскладывает маршруты по обработчикам из `services`
- `services/<имя>/`: один пакет на сервис, внутри функции-обработчики Fiber. При реализации каждая функция собирает модель из тела запроса, проверяет её, при ошибке пишет в лог и отдаёт статус-код, а саму работу делегирует в `domains`
- `models/db/` и `models/api/`: сущности базы и тела запросов с ответами. Пока `.gitkeep`
- `domains/`: бизнес-логика (создать пользователя, выпустить JWT, запрос в Bitbucket). Файлы будут группироваться по смыслу, пока `.gitkeep`

Сервисы сделаны пакетами с функциями, без структур и конструкторов: зависимостей у них пока нет. Структуры появятся вместе с первой зависимостью, то есть вместе с БД

## Fiber

Берём v3 (`github.com/gofiber/fiber/v3`), обработчик имеет вид `func(c fiber.Ctx) error`. Группировка маршрутов через `app.Group("/api/v1")` и вложенные `Group`, по примерам из документации Fiber

## API методы

Префикс `/api/v1`

| Метод | Путь | Обработчик | Что делает | Успех |
|---|---|---|---|---|
| POST | `/auth/register` | `auth.Register` | регистрация, JWT уходит в cookie | 201 |
| POST | `/auth/login` | `auth.Login` | вход, JWT уходит в cookie | 200 |
| POST | `/tasks` | `tasks.Create` | создать таску | 201 |
| GET | `/tasks` | `tasks.List` | список тасок со статусом | 200 |
| GET | `/tasks/:id` | `tasks.Get` | одна таска со статусом | 200 |
| PUT | `/tasks/:id` | `tasks.Update` | изменить таску | 200 |
| DELETE | `/tasks/:id` | `tasks.Delete` | удалить таску | 204 |

Группа `/tasks` будет закрыта проверкой JWT из cookie. Сама проверка живёт в `domains`, поэтому middleware появится в спеке про JWT

Поля тел, чтобы зафиксировать намерение (типы в `models/api` создаются при реализации, сейчас каталог пуст):

- register: `login`, `password`, `bitbucket_token`
- login: `login`, `password`
- task на вход (create, update): `repo_url`, `branch`, `jira_key`, `window_start`, `window_end`, `commits_per_day`
- task в ответе: поля выше плюс `id`, `runs_total`, `last_run_at`

Статус таски из шага 5 README (сколько запусков всего, когда последний) приходит полями `runs_total` и `last_run_at` в ответах `List` и `Get`, отдельного метода под него нет

## Заглушки

Каждый обработчик возвращает `501 Not Implemented` без тела. Над обработчиком стоят swag-аннотации. Логирование, валидация и вызовы `domains` добавляются при реализации

## Swagger

- Спецификация генерируется утилитой `swag` (уже стоит локально) по аннотациям над обработчиками и в `main.go`, формат Swagger 2.0, только json: `swag init --outputTypes json`
- `docs/swagger.json` лежит в репозитории, `main.go` вшивает его через `go:embed`. Файл на диске при запуске не нужен, это удобно при read-only корне пода
- Строка `//go:generate swag init --outputTypes json` в `main.go` даёт перегенерацию через `go generate ./...`
- UI поднимает `swaggerui` из `github.com/gofiber/contrib/v3/swaggerui`, ему отдаётся вшитый json через `Config.FileContent`
- Схем тел пока нет: каталог `models/api` пуст, поэтому в Swagger видны методы, пути, параметры пути и коды ответов. Схемы появятся вместе с моделями

## Решения после подтверждения

- Вход по `login` и `password`, токен Bitbucket принимается при регистрации
- `PUT /tasks/:id` остаётся
- Swagger входит в этот шаг
- README считается более верным источником, чем ранние записи в `reference/`. Расписание вынесено в `reference/0004`. Схему из `reference/0003` (одна таска на пользователя) придётся пересмотреть в спеке про хранилище

## Проверка

- `go build ./...` и `go vet ./...` проходят
- `go generate ./...` создаёт `docs/swagger.json` с семью операциями
- сервер стартует, на каждый из семи маршрутов `curl` отдаёт 501
- неизвестный путь отдаёт 404
- страница Swagger UI открывается и показывает все семь методов

БД на этом шаге не нужна, WSL не задействован
