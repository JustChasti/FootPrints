# Задачи 003: фоновая горутина и ядро core/

Работа идёт в ветке `feature/models`, новую ветку не создаём. Коммиты, push и PR только по команде

Порядок важен: сначала весь код (включая переписывание `domains/users.go`), потом сборка, потом прогон сценария на живой базе. Короткие проверки на базе внутри задач 3 и 4 нужны, чтобы убедиться в поведении GORM до того, как на нём строится остальное

- [x] 1. Заметка `reference/0007-daemon-planned-commits-and-grid.md` (решение, варианты из плана, счётчики и последний запуск в API, последствия, открытые вопросы), уточнение от 2026-10-15 в начале `reference/0004-schedule-work-window-instead-of-cron.md` со ссылкой на `0007`
- [x] 2. `models/db/task.go`: метод `Clock.At(day time.Time) (time.Time, error)`, связь `PlannedCommits` (`json:"-"`, тег каскада), вычисляемые поля `DoneCommits`, `FailedCommits`, `LastRunAt` (тег `gorm:"-"`) и метод `Task.FillStats()`. В `task_test.go` добавить тесты на `Clock.At` и `FillStats`
- [x] 3. `models/db/planned_commit.go`: `PlannedCommit` с полями `PlannedAt`, `ExecutedAt` и остальными из плана (на `Task` тега каскада нет), тип `CommitStatus` и константы статусов `planned`, `done`, `failed`. В `connect.go` добавить `&PlannedCommit{}` в `AutoMigrate`. Проверить на живой базе: у `planned_commits.task_id` ровно один внешний ключ с каскадом, в `tasks` нет колонок `done_commits`, `failed_commits`, `last_run_at`
- [x] 4. `domains/tasks.go`: `Get` и `List` подгружают записи через `Preload` с `Select` трёх колонок (`task_id`, `status`, `executed_at`) и вызывают `FillStats`, `Save` получает `Omit(clause.Associations)`. Проверить на живой базе, записи в `planned_commits` на время проверки вставляются вручную через `psql`: счётчики и `last_run_at` верны, `PUT` не создаёт лишних записей в `planned_commits`, удаление таски убирает её записи
- [x] 5. `domains/users.go`: переписать `List` и `Get` без `query()` и сырого SQL (`Find` по `db.User`, счётчики через `Pluck` и `Count`, как в разделе «Users без сырого SQL» плана). Проверка вместе со сценарием, задача 18
- [x] 6. `config/config.go`: `CoreSettings{TickInterval}`, поле `Core` в `Config`, чтение `CORE_TICK_INTERVAL`, значение по умолчанию `10m`, ошибка при значении не больше нуля. Строка в `.env.example`
- [x] 7. `core/schedule.go`: чистые функции `slots`, `pick`, `overdueReason`
- [x] 8. `core/schedule_test.go`: табличные тесты `slots`, `pick`, `overdueReason` из раздела «Проверка» плана, `go test ./core` проходит
- [x] 9. `core/push.go`: заглушка `push(task db.Task) error`, пишет в `slog` id, репозиторий, ветку и имя коммита
- [x] 10. `core/core.go`: `Core`, `New(gdb, tickInterval)`, `Run(ctx)` (первый опрос сразу, затем ожидание ближайшей точки сетки или отмены `ctx`), `tick(now)` с двумя шагами (планирование дня через `Pluck` и `Find` активных тасок без `NOT EXISTS`, запуск через `Joins("Task")`, на живой базе проверить регистр алиаса присоединённой таблицы), проверка `ctx` между записями, `push` и запись статуса без `ctx`, лог ошибок без падения процесса
- [x] 11. `main.go`: `signal.NotifyContext`, горутина ядра с каналом завершения, `app.Listen` с `fiber.ListenConfig{GracefulContext: ctx}`, после `Listen` отмена `ctx`, ожидание горутины, выход с ошибкой если `Listen` вернул её. Проверить в документации `fiber v3.6.0`, ждёт ли `Listen` завершения запросов при `GracefulContext`
- [x] 12. `go generate ./...`, `go build ./...`, `go vet ./...`, `go test ./...`, убедиться, что в схеме `db.Task` в `docs/swagger.json` есть `done_commits`, `failed_commits` и `last_run_at`
- [x] 13. Запустить сервер с `CORE_TICK_INTERVAL=1m`, пройти шаги 1 и 2 сценария из раздела «Проверка» плана (таблица и индексы, три записи на разные минуты и их выполнение, `done_commits` 3, `failed_commits` 0, `last_run_at` совпадает с последней записью), результаты фиксировать в чате
- [x] 14. Шаги 3 и 4 сценария: новая таска посреди работы, не её день и закрытое окно
- [x] 15. Шаги 5, 6 и 7 сценария: выключение таски, рестарт в тот же день, просроченные записи
- [x] 16. Шаг 8 сценария: ошибка пуша. Заглушку подменяем на возврат ошибки через `go build -overlay` (файл в репозитории не правится), проверяем `failed`, отсутствие повтора и значения `done_commits`, `failed_commits`, `last_run_at` в `GET /tasks`, `GET /tasks/{id}` и ответе `PUT`
- [x] 17. Шаг 9 сценария: остановка по сигналу
- [x] 18. Шаг 10 сценария: `GET /users` и `GET /users/{id}` после переписывания, значения сверить с `psql`
- [x] 19. Обновить `README.md` (структура с `core/` и `planned_commits`, `CORE_TICK_INTERVAL` в настройках, остановка по сигналу, поля `done_commits`, `failed_commits`, `last_run_at` в описании API (вместо `runs_total` из статуса), строка про `0007` в решениях, статус, время окна по времени сервера)
- [x] 20. Типизировать статусы (`db.CommitStatus`), заметка `reference/0008-typed-string-constants.md` и строка в `README.md`. Проверено на живой базе: пути `done` и `failed`, счётчики
