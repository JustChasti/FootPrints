package db

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
)

// Weekdays это дни запуска в ISO-нумерации: 1 понедельник, 7 воскресенье.
// В JSON массив чисел, в Postgres jsonb
type Weekdays []int16

// Clock это время суток в формате HH:MM
type Clock string

// minutes возвращает время в виде минут от полуночи
func (c Clock) minutes() (int, error) {
	t, err := time.Parse("15:04", string(c))
	if err != nil {
		return 0, fmt.Errorf("%q is not HH:MM", string(c))
	}
	return t.Hour()*60 + t.Minute(), nil
}

// TaskParams это поля таски, которые задаёт инженер. Одна структура служит
// и телом запроса, и частью сущности базы
type TaskParams struct {
	RepoURL       string   `json:"repo_url"        gorm:"not null" example:"https://bitbucket.org/team/repo"`
	Branch        string   `json:"branch"          gorm:"not null" example:"main"`
	CommitName    string   `json:"commit_name"     gorm:"not null" example:"PROJ-123 fix typo"`
	Days          Weekdays `json:"days"            gorm:"type:jsonb;serializer:json;not null" swaggertype:"array,integer" example:"1,2,3,4,5"`
	StartTime     Clock    `json:"start_time"      gorm:"not null" swaggertype:"string" example:"09:00"`
	EndTime       Clock    `json:"end_time"        gorm:"not null" swaggertype:"string" example:"18:00"`
	CommitsPerDay int      `json:"commits_per_day" gorm:"not null" example:"1"`
	Active        bool     `json:"active"          gorm:"not null" example:"true"`
}

// Task это таска в базе. Владелец (UserID) наружу не отдаётся
type Task struct {
	ID     int64 `json:"id" gorm:"primaryKey"`
	UserID int64 `json:"-"  gorm:"not null;index"`
	// User нужен только затем, чтобы AutoMigrate создал внешний ключ с каскадом
	User *User `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	TaskParams
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate проверяет параметры таски и приводит их к каноническому виду:
// обрезает пробелы, время переводит в HH:MM с нулём в начале, дни сортирует
func (p *TaskParams) Validate() error {
	p.RepoURL = strings.TrimSpace(p.RepoURL)
	u, err := url.Parse(p.RepoURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("repo_url must be an http(s) URL")
	}

	p.Branch = strings.TrimSpace(p.Branch)
	if p.Branch == "" {
		return errors.New("branch is required")
	}

	p.CommitName = strings.TrimSpace(p.CommitName)
	if p.CommitName == "" {
		return errors.New("commit_name is required")
	}

	if len(p.Days) == 0 {
		return errors.New("days must not be empty")
	}
	for _, d := range p.Days {
		if d < 1 || d > 7 {
			return fmt.Errorf("days: %d is out of range 1..7", d)
		}
	}
	slices.Sort(p.Days)
	for i := 1; i < len(p.Days); i++ {
		if p.Days[i] == p.Days[i-1] {
			return fmt.Errorf("days: %d is repeated", p.Days[i])
		}
	}

	start, err := p.StartTime.minutes()
	if err != nil {
		return fmt.Errorf("start_time: %w", err)
	}
	end, err := p.EndTime.minutes()
	if err != nil {
		return fmt.Errorf("end_time: %w", err)
	}
	if start >= end {
		return errors.New("start_time must be earlier than end_time")
	}
	p.StartTime = Clock(fmt.Sprintf("%02d:%02d", start/60, start%60))
	p.EndTime = Clock(fmt.Sprintf("%02d:%02d", end/60, end%60))

	if p.CommitsPerDay < 1 {
		return errors.New("commits_per_day must be at least 1")
	}
	return nil
}
