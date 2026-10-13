# Задачи 002: модели, конфиг, эндпоинты

Порядок важен: конфиг и модели нужны доменам, домены нужны сервисам, сервисы нужны роутеру

- [x] 0. Создать ветку `feature/models` от `main` (`git switch -c feature/models`), дальше вся работа в ней. Коммиты, push и PR только по команде
- [x] 1. Заметка `reference/0005-orm-gorm-automigrate.md` (сравнение GORM, Bun, ent, sqlc, выбор и две оговорки из плана), ссылка на неё в `reference/0003-storage-postgres-fallback-sqlite-s3.md`
- [x] 2. Поднять Postgres в WSL: `wsl -e docker run -d --name footprints-pg -p 5432:5432 ...` (образ `postgres:17`, логин, пароль и имя базы из `.env`). Проверить `wsl -e docker exec footprints-pg psql -U footprints -c "select 1"`
- [x] 3. `.env.example` (шаблон с `PG_*` (включая `PG_CONNECT_ATTEMPTS` и `PG_CONNECT_DELAY`), `API_*`, `AUTH_JWT_SECRET`, `AUTH_ENC_KEY`), локальный `.env` с рабочими значениями (ключ шифрования сгенерировать, 32 байта в hex), `.env` добавить в `.gitignore`
- [x] 4. Зависимости: `go get gorm.io/gorm gorm.io/driver/postgres github.com/golang-jwt/jwt/v5 github.com/joho/godotenv`, `golang.org/x/crypto` перевести из indirect в прямую
- [x] 5. `config/config.go`: `TaskSettings`, `AuthSettings`, `PGSettings`, `APISettings`, `Load()`, `PGSettings.DSN()`, `APISettings.Addr()`, `TaskSettings.Params()` с копией среза дней. Падение на старте, если не заданы `PG_PASSWORD`, `AUTH_JWT_SECRET`, `AUTH_ENC_KEY`
- [x] 6. `models/db/task.go`: `Weekdays`, `Clock`, `TaskParams` (с `branch`), `Task`, `Validate` (с нормализацией времени и сортировкой дней). Проверить на реальной базе: встроенная структура разворачивается в колонки, `serializer:json` пишет и читает `Weekdays`, поле `User` создаёт внешний ключ с каскадом
- [x] 7. `models/db/task_test.go`: табличный тест на `Validate`, `go test ./models/db` проходит
- [x] 8. `models/db/user.go` и `models/db/connect.go`: `Connect(dsn, attempts, delay)` с циклом ретраев, логом каждой попытки, `gorm.Config{TranslateError: true}` и `AutoMigrate(&User{}, &Task{})`
- [x] 9. `models/api`: `RegisterRequest`, `LoginRequest`, `AuthResult`, `UserInfo`, `Error` с swag-тегами. Для `Weekdays` и `Clock` указать `swaggertype` и `example`
- [x] 10. `domains/auth.go`: `NewAuth`, `Register`, `Login`, выпуск и `Parse` JWT, bcrypt, шифрование токена Bitbucket (AES-GCM), ошибки домена
- [x] 11. `domains/tasks.go`: `NewTasks`, `Create`, `List`, `Get`, `Update` (через `Save`), `Delete`, везде фильтр по `user_id`
- [x] 12. `domains/users.go`: `NewUsers`, `List`, `Get` с `tasks_count`
- [x] 13. `services/auth/auth.go`: структура с зависимостью, `Register`, `Login` (cookie `token`), `Require` (проверка JWT, id пользователя в `c.Locals`), swag-аннотации
- [x] 14. `services/tasks/tasks.go`: переписать заглушки на структуру, обработчики с заготовкой из `TaskSettings.Params()` на create и с текущей таской на update, перевод ошибок домена в статусы (непредвиденные дают 500 без деталей), логирование через `slog`
- [x] 15. `services/users/users.go`: `List`, `Get`
- [x] 16. `router/router.go`: принять собранные сервисы, вешать `Require` на группы `/tasks` и `/users`, добавить маршруты users
- [x] 17. `main.go`: `config.Load`, `db.Connect`, конструкторы доменов и сервисов, `router.Setup`, `app.Listen(cfg.API.Addr())`
- [x] 18. `go generate ./...`, проверить, что в `docs/swagger.json` девять операций и схемы тел
- [x] 19. `go build ./...`, `go vet ./...`, `go test ./...`
- [x] 20. Запустить сервер и пройти шаги 1 по 5 сценария из раздела «Проверка» плана, результаты фиксировать в чате
- [x] 21. Шаг 6 сценария: проверить базу через `wsl -e docker exec`, повторный старт сервера
- [x] 22. Шаги 7 и 8 сценария: ретраи при старте и падение базы во время работы (`wsl -e docker stop footprints-pg` и `start`)
- [x] 23. Обновить `README.md` (структура, конфиг и `.env`, запуск Postgres, ORM, статус, число методов API) и поправить имена полей в `reference/0004-schedule-work-window-instead-of-cron.md`
