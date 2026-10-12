# Задачи 001: базовая структура

Порядок важен: пакеты нужны роутеру, роутер нужен `main.go`

- [x] 0. Новое решение `reference/0004-schedule-work-window-instead-of-cron.md` и строка о нём в `README.md`
- [x] 1. В корне `go mod init footprints`, затем `go get github.com/gofiber/fiber/v3` и `go get github.com/gofiber/contrib/v3/swaggerui`
- [x] 2. Создать каталоги с `.gitkeep`: `models/db/`, `models/api/`, `domains/`
- [x] 3. `services/auth/auth.go`: заглушки `Register` и `Login` со swag-аннотациями
- [x] 4. `services/tasks/tasks.go`: заглушки `Create`, `List`, `Get`, `Update`, `Delete` со swag-аннотациями
- [x] 5. `router/router.go`: функция `Setup(app *fiber.App, swaggerJSON []byte)`, Swagger UI, группа `/api/v1`, подгруппы `/auth` и `/tasks` с семью маршрутами из плана
- [x] 6. `main.go`: общие swag-аннотации (`@title`, `@version`, `@BasePath /api/v1`), `go:embed`, `go:generate`, `fiber.New()`, `router.Setup`, `app.Listen(":8080")`
- [x] 7. `go generate ./...`, проверить, что `docs/swagger.json` содержит семь операций
- [x] 8. `go build ./...` и `go vet ./...`
- [x] 9. Запустить сервер, пройтись `curl` по семи маршрутам (ожидается 501) и по неизвестному пути (ожидается 404), открыть Swagger UI и убедиться, что видны все семь методов
- [x] 10. Обновить `README.md`: строку статуса («кода пока нет») и короткий раздел со структурой каталогов и командой `go generate ./...`
