package core

import (
	"fmt"
	"log/slog"
	"time"

	"footprints/models/db"
)

// push делает коммит таски в Bitbucket и возвращает id коммита. Токен владельца
// лежит в базе зашифрованным, расшифровывается на время одного запроса. Любая
// ошибка уходит наверх, execute запишет её в запись как неудачу
func (c *Core) push(task db.Task) (string, error) {
	var owner db.User
	if err := c.db.Select("bitbucket_token").First(&owner, task.UserID).Error; err != nil {
		return "", fmt.Errorf("load task owner: %w", err)
	}
	token, err := c.decrypt(owner.BitbucketToken)
	if err != nil {
		return "", err
	}
	apiURL, err := repoAPIURL(task.RepoURL, c.apiVersion)
	if err != nil {
		return "", err
	}

	path, body := newFile(time.Now())
	id, err := c.bb.commitFile(apiURL, token, task.Branch, path, task.CommitName, body)
	if err != nil {
		return "", err
	}
	slog.Info("committed", "task", task.ID, "repo", task.RepoURL, "branch", task.Branch, "file", path, "commit", id)
	return id, nil
}
