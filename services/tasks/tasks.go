package tasks

import "github.com/gofiber/fiber/v3"

// Create godoc
//
//	@Summary		Создать таску
//	@Description	Репозиторий, ветка, задача в Jira, рабочее окно и число коммитов в день
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Success		201
//	@Failure		501
//	@Router			/tasks [post]
func Create(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

// List godoc
//
//	@Summary		Список тасок
//	@Description	Таски текущего пользователя со статусом: число запусков и время последнего
//	@Tags			tasks
//	@Produce		json
//	@Success		200
//	@Failure		501
//	@Router			/tasks [get]
func List(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

// Get godoc
//
//	@Summary		Одна таска
//	@Description	Таска со статусом: число запусков и время последнего
//	@Tags			tasks
//	@Produce		json
//	@Param			id	path	int	true	"ID таски"
//	@Success		200
//	@Failure		501
//	@Router			/tasks/{id} [get]
func Get(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

// Update godoc
//
//	@Summary		Изменить таску
//	@Description	Заменяет параметры таски
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"ID таски"
//	@Success		200
//	@Failure		501
//	@Router			/tasks/{id} [put]
func Update(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

// Delete godoc
//
//	@Summary		Удалить таску
//	@Tags			tasks
//	@Param			id	path	int	true	"ID таски"
//	@Success		204
//	@Failure		501
//	@Router			/tasks/{id} [delete]
func Delete(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}
