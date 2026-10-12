package router

import (
	"github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"

	"footprints/services/auth"
	"footprints/services/tasks"
)

func Setup(app *fiber.App, swaggerJSON []byte) {
	app.Use(swaggerui.New(swaggerui.Config{
		BasePath:    "/",
		Path:        "swagger",
		FileContent: swaggerJSON,
	}))

	v1 := app.Group("/api/v1")

	authGroup := v1.Group("/auth")
	authGroup.Post("/register", auth.Register)
	authGroup.Post("/login", auth.Login)

	tasksGroup := v1.Group("/tasks")
	tasksGroup.Post("", tasks.Create)
	tasksGroup.Get("", tasks.List)
	tasksGroup.Get("/:id", tasks.Get)
	tasksGroup.Put("/:id", tasks.Update)
	tasksGroup.Delete("/:id", tasks.Delete)
}
