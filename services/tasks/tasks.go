package tasks

import (
	"github.com/gofiber/fiber/v3"

	"footprints/domains"
	"footprints/services/apierr"
	"footprints/services/auth"
)

type Service struct {
	tasks *domains.Tasks
}

func New(tasks *domains.Tasks) *Service {
	return &Service{tasks: tasks}
}

// Create godoc
//
//	@Summary		Создать таску
//	@Description	Обязательны repo_url, branch и commit_name. Пропущенные days, start_time, end_time, commits_per_day и active берутся из настроек по умолчанию (дни 1-5, окно 09:00-18:00, один коммит в день, таска активна)
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			body	body		db.TaskParams	true	"Параметры таски"
//	@Success		201		{object}	db.Task
//	@Failure		400		{object}	api.Error
//	@Failure		401		{object}	api.Error
//	@Failure		500		{object}	api.Error
//	@Router			/tasks [post]
func (s *Service) Create(c fiber.Ctx) error {
	params := s.tasks.NewParams()
	if err := c.Bind().JSON(&params); err != nil {
		return apierr.Fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	task, err := s.tasks.Create(c.Context(), auth.UserID(c), params)
	if err != nil {
		return apierr.Reply(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(task)
}

// List godoc
//
//	@Summary		Список тасок
//	@Description	Таски текущего пользователя
//	@Tags			tasks
//	@Produce		json
//	@Success		200	{array}		db.Task
//	@Failure		401	{object}	api.Error
//	@Failure		500	{object}	api.Error
//	@Router			/tasks [get]
func (s *Service) List(c fiber.Ctx) error {
	tasks, err := s.tasks.List(c.Context(), auth.UserID(c))
	if err != nil {
		return apierr.Reply(c, err)
	}
	return c.JSON(tasks)
}

// Get godoc
//
//	@Summary		Одна таска
//	@Tags			tasks
//	@Produce		json
//	@Param			id	path		int	true	"ID таски"
//	@Success		200	{object}	db.Task
//	@Failure		400	{object}	api.Error
//	@Failure		401	{object}	api.Error
//	@Failure		404	{object}	api.Error
//	@Failure		500	{object}	api.Error
//	@Router			/tasks/{id} [get]
func (s *Service) Get(c fiber.Ctx) error {
	id, err := apierr.ID(c)
	if err != nil {
		return apierr.Reply(c, err)
	}
	task, err := s.tasks.Get(c.Context(), auth.UserID(c), id)
	if err != nil {
		return apierr.Reply(c, err)
	}
	return c.JSON(task)
}

// Update godoc
//
//	@Summary		Изменить таску
//	@Description	Поля, которых нет в теле, остаются прежними, поэтому таску можно выключить телом {"active": false}
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"ID таски"
//	@Param			body	body		db.TaskParams	true	"Изменяемые параметры"
//	@Success		200		{object}	db.Task
//	@Failure		400		{object}	api.Error
//	@Failure		401		{object}	api.Error
//	@Failure		404		{object}	api.Error
//	@Failure		500		{object}	api.Error
//	@Router			/tasks/{id} [put]
func (s *Service) Update(c fiber.Ctx) error {
	id, err := apierr.ID(c)
	if err != nil {
		return apierr.Reply(c, err)
	}
	task, err := s.tasks.Get(c.Context(), auth.UserID(c), id)
	if err != nil {
		return apierr.Reply(c, err)
	}
	// тело разбирается поверх текущих значений: не пришедшие поля не меняются
	if err := c.Bind().JSON(&task.TaskParams); err != nil {
		return apierr.Fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if err := s.tasks.Save(c.Context(), task); err != nil {
		return apierr.Reply(c, err)
	}
	return c.JSON(task)
}

// Delete godoc
//
//	@Summary		Удалить таску
//	@Tags			tasks
//	@Param			id	path		int	true	"ID таски"
//	@Success		204
//	@Failure		400	{object}	api.Error
//	@Failure		401	{object}	api.Error
//	@Failure		404	{object}	api.Error
//	@Failure		500	{object}	api.Error
//	@Router			/tasks/{id} [delete]
func (s *Service) Delete(c fiber.Ctx) error {
	id, err := apierr.ID(c)
	if err != nil {
		return apierr.Reply(c, err)
	}
	if err := s.tasks.Delete(c.Context(), auth.UserID(c), id); err != nil {
		return apierr.Reply(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
