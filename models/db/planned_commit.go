package db

import "time"

// CommitStatus это статус записи PlannedCommit. В базе лежит обычным текстом
type CommitStatus string

const (
	StatusPlanned CommitStatus = "planned" // ждёт времени
	StatusDone    CommitStatus = "done"    // пуш прошёл
	StatusFailed  CommitStatus = "failed"  // коммита нет: пуш вернул ошибку или запись просрочена, причина в Error
)

// PlannedCommit это один запланированный коммит таски. PlannedAt плановое время
// слота и не меняется, ExecutedAt фактическое время запуска пуша (пусто, если
// пуш не делался)
type PlannedCommit struct {
	ID         int64 `gorm:"primaryKey"`
	TaskID     int64 `gorm:"not null;uniqueIndex:idx_planned_task_at"`
	Task       *Task
	PlannedAt  time.Time    `gorm:"not null;uniqueIndex:idx_planned_task_at;index"`
	Status     CommitStatus `gorm:"not null;default:planned"`
	ExecutedAt *time.Time
	Error      string `gorm:"not null;default:''"`
}
