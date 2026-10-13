package api

// Error это тело ответа с ошибкой
type Error struct {
	Error string `json:"error" example:"branch is required"`
}
