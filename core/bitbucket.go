package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	requestTimeout = 30 * time.Second // пуш не отменяется через ctx, таймаут не даёт запросу держать остановку
	maxErrorBody   = 512              // сколько байт ответа Bitbucket кладём в текст ошибки
)

// bitbucket это клиент REST API Bitbucket Server. Один коммит это один запрос
// PUT .../browse/{path}, клона репозитория нет
type bitbucket struct {
	client *http.Client
}

func newBitbucket() *bitbucket {
	return &bitbucket{client: &http.Client{Timeout: requestTimeout}}
}

// repoAPIURL превращает адрес репозитория из браузера в адрес его REST API:
//
//	https://host/projects/KEY/repos/slug/browse
//	https://host/rest/api/<apiVersion>/projects/KEY/repos/slug
//
// Другие виды адреса (клон /scm/..., личные репозитории /users/..., Cloud)
// не поддерживаются
func repoAPIURL(repoURL, apiVersion string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(repoURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", unsupportedURL(repoURL)
	}
	segs := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	if len(segs) < 4 || segs[0] != "projects" || segs[1] == "" || segs[2] != "repos" || segs[3] == "" {
		return "", unsupportedURL(repoURL)
	}
	return fmt.Sprintf("%s://%s/rest/api/%s/projects/%s/repos/%s", u.Scheme, u.Host, apiVersion, segs[1], segs[3]), nil
}

func unsupportedURL(repoURL string) error {
	return fmt.Errorf("unsupported repo_url %q: expected <scheme>://<host>/projects/KEY/repos/SLUG/browse", repoURL)
}

// commitFile создаёт новый файл path в ветке branch одним коммитом и возвращает
// id коммита. apiURL это результат repoAPIURL. Автором коммита Bitbucket
// назначает владельца токена
func (b *bitbucket) commitFile(apiURL, token, branch, path, message string, content []byte) (string, error) {
	// запись в bytes.Buffer не возвращает ошибок, поэтому их не проверяем
	var form bytes.Buffer
	mw := multipart.NewWriter(&form)
	mw.WriteField("message", message)
	mw.WriteField("branch", branch)
	fw, _ := mw.CreateFormFile("content", "file")
	fw.Write(content)
	mw.Close()

	req, err := http.NewRequest(http.MethodPut, apiURL+"/browse/"+path, &form)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Atlassian-Token", "no-check")
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return "", fmt.Errorf("bitbucket answered %d: %s", resp.StatusCode, cleanText(body))
	}

	// коммит уже создан, поэтому ответ без id не делает пуш неудачным
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || out.ID == "" {
		slog.Warn("bitbucket answered without a commit id", "status", resp.StatusCode, "err", err)
	}
	return out.ID, nil
}

// cleanText готовит кусок ответа к записи в базу. Обрезка по байтам может
// разрезать символ, а Postgres не принимает в text ни неверный UTF-8, ни нулевой
// байт: запись статуса упала бы, и коммит делался бы заново на каждом опросе
func cleanText(b []byte) string {
	s := strings.ToValidUTF8(string(b), "")
	return strings.TrimSpace(strings.ReplaceAll(s, "\x00", ""))
}
