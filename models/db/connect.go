package db

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect подключается к базе и создаёт таблицы. Пока база недоступна, пробует
// ещё раз через delay, всего attempts попыток
func Connect(dsn string, attempts int, delay time.Duration) (*gorm.DB, error) {
	var lastErr error
	for i := 1; i <= attempts; i++ {
		gdb, err := open(dsn)
		if err == nil {
			return gdb, nil
		}
		lastErr = err
		slog.Warn("database is not ready", "attempt", i, "of", attempts, "err", err)
		if i < attempts {
			time.Sleep(delay)
		}
	}
	return nil, fmt.Errorf("database is unavailable after %d attempts: %w", attempts, lastErr)
}

func open(dsn string) (*gorm.DB, error) {
	// на время подключения логгер GORM молчит: ошибку и номер попытки пишет Connect
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := gdb.AutoMigrate(&User{}, &Task{}, &PlannedCommit{}); err != nil {
		if sqlDB, e := gdb.DB(); e == nil {
			sqlDB.Close()
		}
		return nil, err
	}
	gdb.Logger = logger.Default
	return gdb, nil
}
