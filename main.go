package main

import (
	_ "embed"
	"log"

	"github.com/gofiber/fiber/v3"

	"footprints/router"
)

//go:generate swag init --outputTypes json

//go:embed docs/swagger.json
var swaggerJSON []byte

// @title		Footprints API
// @version	1.0
// @BasePath	/api/v1
func main() {
	app := fiber.New()
	router.Setup(app, swaggerJSON)
	log.Fatal(app.Listen(":8080"))
}
