package domains

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"footprints/config"
	"footprints/models/db"
)

type Tasks struct {
	db       *gorm.DB
	defaults config.TaskSettings
}

func NewTasks(gdb *gorm.DB, defaults config.TaskSettings) *Tasks {
	return &Tasks{db: gdb, defaults: defaults}
}

// NewParams возвращает заготовку таски с дефолтами, поверх неё сервис
// разбирает тело запроса
func (t *Tasks) NewParams() db.TaskParams {
	return t.defaults.Params()
}

func (t *Tasks) Create(ctx context.Context, userID int64, params db.TaskParams) (*db.Task, error) {
	task := db.Task{UserID: userID, TaskParams: params}
	if err := t.Save(ctx, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (t *Tasks) List(ctx context.Context, userID int64) ([]db.Task, error) {
	tasks := []db.Task{}
	err := t.db.WithContext(ctx).Where("user_id = ?", userID).Order("id").Find(&tasks).Error
	return tasks, err
}

func (t *Tasks) Get(ctx context.Context, userID, id int64) (*db.Task, error) {
	var task db.Task
	err := t.db.WithContext(ctx).First(&task, "id = ? AND user_id = ?", id, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Save проверяет параметры и записывает таску целиком. Целую строку пишем
// через Save, чтобы нулевые значения (active=false) не пропадали
func (t *Tasks) Save(ctx context.Context, task *db.Task) error {
	if err := task.TaskParams.Validate(); err != nil {
		return &ValidationError{err.Error()}
	}
	return t.db.WithContext(ctx).Save(task).Error
}

func (t *Tasks) Delete(ctx context.Context, userID, id int64) error {
	res := t.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&db.Task{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
