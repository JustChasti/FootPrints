package db

import "time"

// User это пользователь в базе. Хеш пароля и шифрованный токен Bitbucket
// наружу не отдаются
type User struct {
	ID             int64     `json:"id"    gorm:"primaryKey"`
	Login          string    `json:"login" gorm:"not null;uniqueIndex"`
	PasswordHash   string    `json:"-"     gorm:"not null"`
	BitbucketToken []byte    `json:"-"     gorm:"not null"`
	CreatedAt      time.Time `json:"created_at"`
}
