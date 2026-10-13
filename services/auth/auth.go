package auth

import (
	"github.com/gofiber/fiber/v3"

	"footprints/domains"
	"footprints/models/api"
	"footprints/services/apierr"
)

const cookieName = "token"

type userKey struct{}

type Service struct {
	auth *domains.Auth
}

func New(auth *domains.Auth) *Service {
	return &Service{auth: auth}
}

// UserID возвращает id пользователя, которого определил Require
func UserID(c fiber.Ctx) int64 {
	id, _ := c.Locals(userKey{}).(int64)
	return id
}

// Require пропускает дальше только запросы с действующим JWT в cookie
func (s *Service) Require(c fiber.Ctx) error {
	id, err := s.auth.Parse(c.Cookies(cookieName))
	if err != nil {
		return apierr.Reply(c, err)
	}
	c.Locals(userKey{}, id)
	return c.Next()
}

// Register godoc
//
//	@Summary		Регистрация инженера
//	@Description	Создаёт пользователя по логину, паролю и токену Bitbucket, кладёт JWT в cookie token. Пароль хранится в виде bcrypt-хеша, токен Bitbucket в шифрованном виде
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		api.RegisterRequest	true	"Данные пользователя"
//	@Success		201		{object}	api.AuthResult
//	@Failure		400		{object}	api.Error
//	@Failure		409		{object}	api.Error
//	@Failure		500		{object}	api.Error
//	@Router			/auth/register [post]
func (s *Service) Register(c fiber.Ctx) error {
	var req api.RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierr.Fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	session, err := s.auth.Register(c.Context(), req)
	if err != nil {
		return apierr.Reply(c, err)
	}
	return s.reply(c, fiber.StatusCreated, session)
}

// Login godoc
//
//	@Summary		Вход
//	@Description	Проверяет логин и пароль, кладёт JWT в cookie token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		api.LoginRequest	true	"Логин и пароль"
//	@Success		200		{object}	api.AuthResult
//	@Failure		400		{object}	api.Error
//	@Failure		401		{object}	api.Error
//	@Failure		500		{object}	api.Error
//	@Router			/auth/login [post]
func (s *Service) Login(c fiber.Ctx) error {
	var req api.LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apierr.Fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	session, err := s.auth.Login(c.Context(), req)
	if err != nil {
		return apierr.Reply(c, err)
	}
	return s.reply(c, fiber.StatusOK, session)
}

func (s *Service) reply(c fiber.Ctx, status int, session *domains.Session) error {
	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HTTPOnly: true,
		SameSite: "Lax",
	})
	return c.Status(status).JSON(api.AuthResult{
		ID:        session.UserID,
		Login:     session.Login,
		ExpiresAt: session.ExpiresAt,
	})
}
