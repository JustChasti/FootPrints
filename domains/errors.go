package domains

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrLoginTaken     = errors.New("login is already taken")
	ErrBadCredentials = errors.New("invalid login or password")
	ErrUnauthorized   = errors.New("unauthorized")
)

// ValidationError это нарушенное правило входных данных, сервис отдаёт его
// текст клиенту с кодом 400
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }
