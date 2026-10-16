package core

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"footprints/models/db"
)

// Core это фоновая горутина: раз в период планирует коммиты тасок на день и
// запускает те, чьё время пришло
type Core struct {
	db         *gorm.DB
	step       time.Duration // период опроса и шаг сетки времён коммитов
	apiVersion string        // версия REST API Bitbucket в пути запроса
	decrypt    func([]byte) (string, error)
	bb         *bitbucket
}

// New принимает период опроса, версию REST API Bitbucket и функцию, которая
// расшифровывает токен Bitbucket из базы
func New(gdb *gorm.DB, step time.Duration, apiVersion string, decrypt func([]byte) (string, error)) *Core {
	return &Core{db: gdb, step: step, apiVersion: apiVersion, decrypt: decrypt, bb: newBitbucket()}
}

// Run крутит опросы, пока не отменён ctx. Первый опрос сразу, дальше на точках
// сетки (10:00, 10:10, ... при шаге 10 минут), поэтому запись на 10:20
// выполняется на опросе в 10:20
func (c *Core) Run(ctx context.Context) {
	slog.Info("core started", "tick", c.step)
	for {
		c.tick(ctx, time.Now())

		wait := time.Until(time.Now().Truncate(c.step).Add(c.step))
		select {
		case <-ctx.Done():
			slog.Info("core stopped")
			return
		case <-time.After(wait):
		}
	}
}

// tick это один опрос: сначала планируем сегодняшний день там, где его ещё
// нет, потом выполняем записи, чьё время пришло. Ошибки пишем в лог, опрос
// продолжится на следующей точке сетки
func (c *Core) tick(ctx context.Context, now time.Time) {
	if err := c.planDay(now); err != nil {
		slog.Error("plan the day", "err", err)
	}
	if err := c.runDue(ctx, now); err != nil {
		slog.Error("run planned commits", "err", err)
	}
}

// planDay планирует сегодняшний день активным таскам, у которых нет записей
// за сегодня. Сброса нет: со сменой суток вчерашние записи перестают считаться
// сегодняшними, и первый опрос после полуночи планирует таску заново
func (c *Core) planDay(now time.Time) error {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

	var plannedIDs []int64
	err := c.db.Model(&db.PlannedCommit{}).
		Where("planned_at >= ? AND planned_at < ?", dayStart, dayEnd).
		Distinct().
		Pluck("task_id", &plannedIDs).Error
	if err != nil {
		return err
	}
	planned := make(map[int64]bool, len(plannedIDs))
	for _, id := range plannedIDs {
		planned[id] = true
	}

	var tasks []db.Task
	if err := c.db.Where("active = ?", true).Order("id").Find(&tasks).Error; err != nil {
		return err
	}
	for _, task := range tasks {
		if planned[task.ID] {
			continue
		}
		free, err := slots(task, now, c.step)
		if err != nil {
			slog.Error("task window is broken", "task", task.ID, "err", err)
			continue
		}
		picked := pick(free, task.CommitsPerDay)
		if len(picked) == 0 {
			continue
		}
		rows := make([]db.PlannedCommit, len(picked))
		for i, at := range picked {
			rows[i] = db.PlannedCommit{TaskID: task.ID, PlannedAt: at, Status: db.StatusPlanned}
		}
		if err := c.db.Create(&rows).Error; err != nil {
			slog.Error("save planned commits", "task", task.ID, "err", err)
			continue
		}
		slog.Info("day planned", "task", task.ID, "commits", len(rows), "first", picked[0].Format("15:04"))
	}
	return nil
}

// runDue выполняет записи активных тасок, чьё время пришло. Между записями
// проверяет ctx: пуш, который уже идёт, доделывается до конца
func (c *Core) runDue(ctx context.Context, now time.Time) error {
	var due []db.PlannedCommit
	err := c.db.Joins("Task").
		Where("planned_commits.status = ? AND planned_commits.planned_at <= ?", db.StatusPlanned, now).
		Where(`"Task".active = ?`, true).
		Order("planned_commits.planned_at").
		Find(&due).Error
	if err != nil {
		return err
	}
	for _, rec := range due {
		if ctx.Err() != nil {
			return nil
		}
		c.execute(rec, now)
	}
	return nil
}

// execute обрабатывает одну запись: просроченную помечает неудачной без пуша,
// остальным делает пуш. Неудачная запись в этот день не повторяется
func (c *Core) execute(rec db.PlannedCommit, now time.Time) {
	task := *rec.Task

	if reason := overdueReason(task, rec.PlannedAt, now); reason != "" {
		slog.Warn("planned commit is overdue", "commit", rec.ID, "task", task.ID, "reason", reason)
		c.finish(rec.ID, map[string]any{"status": db.StatusFailed, "error": reason})
		return
	}

	executedAt := time.Now()
	commitID, err := c.push(task)
	if err != nil {
		slog.Error("push failed", "commit", rec.ID, "task", task.ID, "err", err)
		c.finish(rec.ID, map[string]any{"status": db.StatusFailed, "executed_at": executedAt, "error": err.Error()})
		return
	}
	c.finish(rec.ID, map[string]any{"status": db.StatusDone, "executed_at": executedAt, "commit_id": commitID})
}

// finish записывает итог по записи. Если запись не удалось сохранить, она
// останется planned и выполнится на следующем опросе
func (c *Core) finish(id int64, fields map[string]any) {
	if err := c.db.Model(&db.PlannedCommit{}).Where("id = ?", id).Updates(fields).Error; err != nil {
		slog.Error("save commit status", "commit", id, "err", err)
	}
}
