package domains

import (
	"context"

	"gorm.io/gorm"

	"footprints/models/api"
)

type Users struct {
	db *gorm.DB
}

func NewUsers(gdb *gorm.DB) *Users {
	return &Users{db: gdb}
}

// query отдаёт пользователей с числом их тасок
func (u *Users) query(ctx context.Context) *gorm.DB {
	return u.db.WithContext(ctx).Table("users").
		Select("users.id, users.login, count(tasks.id) AS tasks_count").
		Joins("LEFT JOIN tasks ON tasks.user_id = users.id").
		Group("users.id").
		Order("users.id")
}

func (u *Users) List(ctx context.Context) ([]api.UserInfo, error) {
	users := []api.UserInfo{}
	err := u.query(ctx).Scan(&users).Error
	return users, err
}

func (u *Users) Get(ctx context.Context, id int64) (*api.UserInfo, error) {
	users := []api.UserInfo{}
	if err := u.query(ctx).Where("users.id = ?", id).Scan(&users).Error; err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, ErrNotFound
	}
	return &users[0], nil
}
