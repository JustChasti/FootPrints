package auth

import "github.com/gofiber/fiber/v3"

// Register godoc
//
//	@Summary		Регистрация инженера
//	@Description	Создаёт пользователя по логину, паролю и токену Bitbucket, кладёт JWT в cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		201
//	@Failure		501
//	@Router			/auth/register [post]
func Register(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

// Login godoc
//
//	@Summary		Вход
//	@Description	Проверяет логин и пароль, кладёт JWT в cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		501
//	@Router			/auth/login [post]
func Login(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}
