package router

import (
	"github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"

	"footprints/services/auth"
	"footprints/services/tasks"
	"footprints/services/users"
)

func Setup(app *fiber.App, swaggerJSON []byte, authSvc *auth.Service, tasksSvc *tasks.Service, usersSvc *users.Service) {
	app.Use(swaggerui.New(swaggerui.Config{
		BasePath:    "/",
		Path:        "swagger",
		FileContent: swaggerJSON,
	}))

	v1 := app.Group("/api/v1")

	authGroup := v1.Group("/auth")
	authGroup.Post("/register", authSvc.Register)
	authGroup.Post("/login", authSvc.Login)

	tasksGroup := v1.Group("/tasks", authSvc.Require)
	tasksGroup.Post("", tasksSvc.Create)
	tasksGroup.Get("", tasksSvc.List)
	tasksGroup.Get("/:id", tasksSvc.Get)
	tasksGroup.Put("/:id", tasksSvc.Update)
	tasksGroup.Delete("/:id", tasksSvc.Delete)

	usersGroup := v1.Group("/users", authSvc.Require)
	usersGroup.Get("", usersSvc.List)
	usersGroup.Get("/:id", usersSvc.Get)
}
