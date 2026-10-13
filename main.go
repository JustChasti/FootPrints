package main

import (
	_ "embed"
	"log"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"footprints/config"
	"footprints/domains"
	"footprints/models/db"
	"footprints/router"
	"footprints/services/auth"
	"footprints/services/tasks"
	"footprints/services/users"
)

//go:generate swag init --outputTypes json

//go:embed docs/swagger.json
var swaggerJSON []byte

// @title		Footprints API
// @version	1.0
// @BasePath	/api/v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	gdb, err := db.Connect(cfg.PG.DSN(), cfg.PG.ConnectAttempts, cfg.PG.ConnectDelay)
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("database is ready")

	authDomain, err := domains.NewAuth(gdb, cfg.Auth)
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()
	router.Setup(app, swaggerJSON,
		auth.New(authDomain),
		tasks.New(domains.NewTasks(gdb, cfg.Task)),
		users.New(domains.NewUsers(gdb)),
	)
	log.Fatal(app.Listen(cfg.API.Addr()))
}
