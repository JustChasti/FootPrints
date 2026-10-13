package api

// UserInfo это пользователь для чтения: имя и число его тасок
type UserInfo struct {
	ID         int64  `json:"id"`
	Login      string `json:"login"`
	TasksCount int64  `json:"tasks_count"`
}
