package domains

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"footprints/models/api"
	"footprints/models/db"
)

type Users struct {
	db *gorm.DB
}

func NewUsers(gdb *gorm.DB) *Users {
	return &Users{db: gdb}
}

func (u *Users) List(ctx context.Context) ([]api.UserInfo, error) {
	var users []db.User
	if err := u.db.WithContext(ctx).Order("id").Find(&users).Error; err != nil {
		return nil, err
	}

	// число тасок считаем в Go: достаточно списка владельцев всех тасок
	var owners []int64
	if err := u.db.WithContext(ctx).Model(&db.Task{}).Pluck("user_id", &owners).Error; err != nil {
		return nil, err
	}
	counts := make(map[int64]int64, len(users))
	for _, id := range owners {
		counts[id]++
	}

	infos := make([]api.UserInfo, 0, len(users))
	for _, user := range users {
		infos = append(infos, api.UserInfo{ID: user.ID, Login: user.Login, TasksCount: counts[user.ID]})
	}
	return infos, nil
}

func (u *Users) Get(ctx context.Context, id int64) (*api.UserInfo, error) {
	var user db.User
	err := u.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var count int64
	if err := u.db.WithContext(ctx).Model(&db.Task{}).Where("user_id = ?", id).Count(&count).Error; err != nil {
		return nil, err
	}
	return &api.UserInfo{ID: user.ID, Login: user.Login, TasksCount: count}, nil
}
