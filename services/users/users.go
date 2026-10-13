package users

import (
	"github.com/gofiber/fiber/v3"

	"footprints/domains"
	"footprints/services/apierr"
)

type Service struct {
	users *domains.Users
}

func New(users *domains.Users) *Service {
	return &Service{users: users}
}

// List godoc
//
//	@Summary		Список пользователей
//	@Description	Все пользователи с числом их тасок
//	@Tags			users
//	@Produce		json
//	@Success		200	{array}		api.UserInfo
//	@Failure		401	{object}	api.Error
//	@Failure		500	{object}	api.Error
//	@Router			/users [get]
func (s *Service) List(c fiber.Ctx) error {
	users, err := s.users.List(c.Context())
	if err != nil {
		return apierr.Reply(c, err)
	}
	return c.JSON(users)
}

// Get godoc
//
//	@Summary		Один пользователь
//	@Description	Имя пользователя и число его тасок
//	@Tags			users
//	@Produce		json
//	@Param			id	path		int	true	"ID пользователя"
//	@Success		200	{object}	api.UserInfo
//	@Failure		400	{object}	api.Error
//	@Failure		401	{object}	api.Error
//	@Failure		404	{object}	api.Error
//	@Failure		500	{object}	api.Error
//	@Router			/users/{id} [get]
func (s *Service) Get(c fiber.Ctx) error {
	id, err := apierr.ID(c)
	if err != nil {
		return apierr.Reply(c, err)
	}
	user, err := s.users.Get(c.Context(), id)
	if err != nil {
		return apierr.Reply(c, err)
	}
	return c.JSON(user)
}
