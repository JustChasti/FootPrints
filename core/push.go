package core

import (
	"log/slog"

	"footprints/models/db"
)

// push делает коммит таски. Пока это заглушка: пишет в лог и ничего не меняет
// в репозитории. На её месте будет коммит в Bitbucket
func push(task db.Task) error {
	slog.Info("push (stub)",
		"task", task.ID,
		"repo", task.RepoURL,
		"branch", task.Branch,
		"commit_name", task.CommitName,
	)
	return nil
}
