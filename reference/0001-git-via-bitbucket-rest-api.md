# reference 0001: Коммиты через REST API Bitbucket Server

Статус: принято (путь 1 из 2, второй путь описан в `0002-git-clone-single-branch.md`)

Дата: 2026-10-09

Уточнение от 2026-10-16 (спека 004, реализовано):
- Выбран этот путь. Версия API в пути `1.0` вместо `latest`, она задаётся настройкой `BITBUCKET_API_VERSION`
- Инженер хранит в `repo_url` адрес из браузера (`https://host/projects/KEY/repos/slug/browse`), сервис превращает его в адрес API `https://host/rest/api/1.0/projects/KEY/repos/slug` и дописывает `/browse/<путь файла>`. Другие виды адреса не поддерживаются
- Файл коммита: `autocommit/<yyyy-mm-dd>/<12 символов sha256 содержимого>.txt`, внутри 32-64 строки случайных sha256 в hex. Сообщение коммита это `commit_name` таски
- Id коммита из ответа хранится в `planned_commits.commit_id`. Текст ответа при ошибке (до 512 байт) попадает в `planned_commits.error`
- Шифруется токен Bitbucket, который лежит в базе (AES-GCM, `AUTH_ENC_KEY`). Сами запросы к Bitbucket сервис не шифрует, на проде их защищает HTTPS
- Запрос идёт с заголовком `X-Atlassian-Token: no-check`, таймаут 30 секунд, повторов нет
- Токен из `api-key.txt` оказался личным API-токеном Atlassian Cloud (`ATATT3...`): запросы к `api.bitbucket.org` с Bearer дают 401, Cloud принимает его только как Basic `email:токен`. Bearer-токены Cloud требуют платного тарифа, проверить их нельзя. Cloud не реализован
- Проверено на локальной подделке Bitbucket Server и в тестах на `httptest`. Живая проверка на настоящем Server за пользователем, после неё этот документ нужно обновить (поддержка REST на конкретной версии, защита ветки, формат `repo_url`)

## Контекст

Сервис автоматизации по расписанию создаёт коммиты со случайно сгенерированными файлами от имени инженера: в его репозиторий, в его ветку, с привязкой к story в Jira. Инженер регистрируется в сервисе через Swagger и оставляет id, персональный токен Bitbucket, репозиторий, ветку, story, число коммитов в день и cron

Сервис разворачивается на платформе Сбера в dev-контуре. Постоянные тома там запрещены. В чартах есть только `emptyDir`, корневая файловая система пода доступна только на чтение, процесс идёт под uid 10001, писать можно в `/tmp` (`emptyDir` на диске ноды). Состояние (пользователи, план коммитов, история) лежит во внешней БД

Нужно выбрать способ попадания в Bitbucket. Этот документ описывает путь через REST API, путь через клон описан в `0002-git-clone-single-branch.md`

## Описание пути

Один HTTP-запрос на один коммит, никакого клона:

```
PUT https://<bitbucket-host>/rest/api/latest/projects/<KEY>/repos/<slug>/browse/<path>
Authorization: Bearer <персональный токен>
Content-Type: multipart/form-data
  content = <тело файла>
  message = PROJ-123: <текст>
  branch  = <ветка>
```

Для нового файла параметр `sourceCommitId` не нужен. Ответ 200 содержит id созданного коммита. Его сохраняем в таблицу `commits`: это статистика и доказательство, что коммит сделан

Каждый раз создаётся новый файл вида `autocommit/<yyyy-mm-dd>/<random>.txt`, поэтому конфликтов с чужими коммитами в ветке не возникает

Клиент на стандартной библиотеке Go:

```go
func commitFile(ctx context.Context, host, key, slug, branch, token, path, msg string, body []byte) (string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("message", msg)
	mw.WriteField("branch", branch)
	fw, _ := mw.CreateFormFile("content", "f")
	fw.Write(body)
	mw.Close()

	url := fmt.Sprintf("https://%s/rest/api/latest/projects/%s/repos/%s/browse/%s", host, key, slug, path)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("bitbucket %d: %s", resp.StatusCode, b)
	}
	var out struct{ ID string `json:"id"` }
	json.NewDecoder(resp.Body).Decode(&out)
	return out.ID, nil
}
```

## Плюсы

- Нет клона, диска, бинарника `git` в образе и очистки после коммита
- Нет push-конфликтов и ребейзов, если кто-то двинул ветку
- Автор коммита берётся из профиля владельца токена, имя и email в сервисе хранить не нужно, KPI засчитается владельцу
- Память: тело запроса измеряется килобайтами
- Одна зависимость, `net/http`
- Egress: достаточно записи `http_servers` на хост Bitbucket

## Минусы и риски

- Наличие и поведение endpoint на конкретной версии Bitbucket не проверены (уверенность в работоспособности около 80%)
- Путь привязан к Bitbucket Server / Data Center, на GitHub и GitLab он не переносится
- Один коммит содержит ровно один файл
- Pre-receive хуки и защита ветки применяются и к REST-коммитам: при обязательном PR или проверке формата сообщения будет ответ 4xx
- Запросы идут под токеном инженера, возможен rate limit на стороне Bitbucket

## Проверка до разработки

Один запрос на тестовую ветку своего репозитория:

```bash
echo test > f.txt
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -F content=@f.txt -F message='PROJ-123 test' -F branch=my-branch \
  "https://<bitbucket-host>/rest/api/latest/projects/<KEY>/repos/<slug>/browse/autocommit/f.txt"
```

Ответ 200 с json, где есть `id`, означает, что путь рабочий. Ответ 4xx с текстом хука или защиты ветки объясняет причину отказа, текст ответа сервис должен сохранять в `commits.status`

## Что нужно от окружения

- Запись `http_servers` на хост Bitbucket в values неймспейса
- Персональный токен инженера с правом записи в репозиторий
- Ключ шифрования токенов в БД, подаётся через `app.secrets`

## Открытые вопросы

- Версия Bitbucket и доступность `PUT .../browse/{path}`
- Защищена ли целевая ветка и есть ли проверка формата сообщения коммита
