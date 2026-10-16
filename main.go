package main

import (
	"context"
	_ "embed"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"

	"footprints/config"
	"footprints/core"
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

	// сигнал остановки доходит и до сервера, и до ядра
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ядро живёт в фоне, веб-сервер вокруг него
	coreDone := make(chan struct{})
	go func() {
		defer close(coreDone)
		core.New(gdb, cfg.Core.TickInterval, cfg.Bitbucket.APIVersion, authDomain.DecryptToken).Run(ctx)
	}()

	app := fiber.New()
	router.Setup(app, swaggerJSON,
		auth.New(authDomain),
		tasks.New(domains.NewTasks(gdb, cfg.Task)),
		users.New(domains.NewUsers(gdb)),
	)
	err = app.Listen(cfg.API.Addr(), fiber.ListenConfig{GracefulContext: ctx})

	// Listen вернулся: по сигналу или с ошибкой (например, порт занят).
	// Останавливаем ядро и ждём, пока оно закончит текущий пуш
	stop()
	<-coreDone
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("stopped")
}
