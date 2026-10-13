// Package apierr переводит ошибки доменов в ответы API
package apierr

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"footprints/domains"
	"footprints/models/api"
)

// Reply пишет ошибку в лог и отвечает клиенту подходящим кодом. Непредвиденные
// ошибки (в том числе недоступная база) дают 500 без внутренних деталей
func Reply(c fiber.Ctx, err error) error {
	var verr *domains.ValidationError
	switch {
	case errors.As(err, &verr):
		return Fail(c, fiber.StatusBadRequest, verr.Msg)
	case errors.Is(err, domains.ErrNotFound):
		return Fail(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, domains.ErrLoginTaken):
		return Fail(c, fiber.StatusConflict, err.Error())
	case errors.Is(err, domains.ErrBadCredentials), errors.Is(err, domains.ErrUnauthorized):
		return Fail(c, fiber.StatusUnauthorized, err.Error())
	}
	slog.Error("request failed", "method", c.Method(), "path", c.Path(), "err", err)
	return c.Status(fiber.StatusInternalServerError).JSON(api.Error{Error: "internal error"})
}

// Fail отвечает ошибкой клиента с заданным кодом и пишет её в лог
func Fail(c fiber.Ctx, status int, msg string) error {
	slog.Warn("request rejected", "method", c.Method(), "path", c.Path(), "status", status, "msg", msg)
	return c.Status(status).JSON(api.Error{Error: msg})
}

// ID читает числовой параметр пути :id
func ID(c fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, &domains.ValidationError{Msg: "id must be a positive integer"}
	}
	return id, nil
}
