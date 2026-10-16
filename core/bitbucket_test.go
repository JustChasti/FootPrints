package core

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestRepoAPIURL(t *testing.T) {
	cases := []struct {
		name    string
		repoURL string
		version string
		want    string // пусто, если ждём ошибку
	}{
		{"url from the spec", "https://bitbucket.example.com/projects/PROJ/repos/repo-name/browse", "1.0",
			"https://bitbucket.example.com/rest/api/1.0/projects/PROJ/repos/repo-name"},
		{"without browse", "https://bitbucket.example.com/projects/PROJ/repos/repo-name", "1.0",
			"https://bitbucket.example.com/rest/api/1.0/projects/PROJ/repos/repo-name"},
		{"trailing slash", "https://bitbucket.example.com/projects/PROJ/repos/repo-name/", "1.0",
			"https://bitbucket.example.com/rest/api/1.0/projects/PROJ/repos/repo-name"},
		{"browse with a directory and a ref", "https://h/projects/K/repos/s/browse/dir?at=refs/heads/x", "1.0",
			"https://h/rest/api/1.0/projects/K/repos/s"},
		{"port and http scheme are kept", "http://127.0.0.1:7990/projects/K/repos/s/browse", "1.0",
			"http://127.0.0.1:7990/rest/api/1.0/projects/K/repos/s"},
		{"version comes from the argument", "https://h/projects/K/repos/s/browse", "2.0",
			"https://h/rest/api/2.0/projects/K/repos/s"},
		{"spaces around", "  https://h/projects/K/repos/s/browse  ", "1.0",
			"https://h/rest/api/1.0/projects/K/repos/s"},

		{"bitbucket cloud", "https://bitbucket.org/team/repo", "1.0", ""},
		{"clone url", "https://h/scm/KEY/slug.git", "1.0", ""},
		{"personal repository", "https://h/users/name/repos/slug/browse", "1.0", ""},
		{"context path before projects", "https://h/bitbucket/projects/K/repos/s/browse", "1.0", ""},
		{"empty project key", "https://h/projects//repos/s/browse", "1.0", ""},
		{"empty slug", "https://h/projects/K/repos/", "1.0", ""},
		{"too short", "https://h/projects/K", "1.0", ""},
		{"no path", "https://h", "1.0", ""},
		{"no scheme", "h/projects/K/repos/s", "1.0", ""},
		{"empty", "", "1.0", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := repoAPIURL(c.repoURL, c.version)
			if c.want == "" {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				if !strings.Contains(err.Error(), "unsupported repo_url") {
					t.Fatalf("unexpected error text: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

// recorded это запрос, который получил fakeServer
type recorded struct {
	method, path, auth, xsrf   string
	message, branch            string
	content                    string
	contentTypeIsMultipartForm bool
}

// fakeServer поднимает httptest-сервер с заданным ответом и запоминает запрос
func fakeServer(t *testing.T, status int, answer string) (*httptest.Server, *recorded) {
	t.Helper()
	got := &recorded{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.auth = r.Header.Get("Authorization")
		got.xsrf = r.Header.Get("X-Atlassian-Token")
		if err := r.ParseMultipartForm(1 << 20); err == nil {
			got.contentTypeIsMultipartForm = true
			got.message = r.FormValue("message")
			got.branch = r.FormValue("branch")
			if f, _, err := r.FormFile("content"); err == nil {
				b, _ := io.ReadAll(f)
				got.content = string(b)
			}
		}
		w.WriteHeader(status)
		io.WriteString(w, answer)
	}))
	t.Cleanup(srv.Close)
	return srv, got
}

func TestCommitFile(t *testing.T) {
	srv, got := fakeServer(t, http.StatusOK, `{"id":"abc123","displayId":"abc"}`)
	bb := &bitbucket{client: srv.Client()}

	id, err := bb.commitFile(srv.URL+"/rest/api/1.0/projects/KEY/repos/slug", "tok-1", "feature/x",
		"autocommit/2026-10-16/aaa.txt", "PROJ-1 work", []byte("line1\nline2\n"))
	if err != nil {
		t.Fatal(err)
	}
	if id != "abc123" {
		t.Fatalf("got id %q", id)
	}
	if got.method != http.MethodPut {
		t.Errorf("method %q", got.method)
	}
	if want := "/rest/api/1.0/projects/KEY/repos/slug/browse/autocommit/2026-10-16/aaa.txt"; got.path != want {
		t.Errorf("path %q, want %q", got.path, want)
	}
	if got.auth != "Bearer tok-1" {
		t.Errorf("Authorization %q", got.auth)
	}
	if got.xsrf != "no-check" {
		t.Errorf("X-Atlassian-Token %q", got.xsrf)
	}
	if !got.contentTypeIsMultipartForm {
		t.Fatal("request is not multipart/form-data")
	}
	if got.message != "PROJ-1 work" || got.branch != "feature/x" || got.content != "line1\nline2\n" {
		t.Errorf("form: message %q, branch %q, content %q", got.message, got.branch, got.content)
	}
}

func TestCommitFileAnswers(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		body      string
		wantID    string
		wantErr   string // подстрока ошибки, пусто если ошибки нет
		maxErrLen int
	}{
		{"created", http.StatusCreated, `{"id":"c1"}`, "c1", "", 0},
		{"forbidden with a body", http.StatusForbidden, `{"errors":[{"message":"branch is protected"}]}`, "",
			`bitbucket answered 403: {"errors":[{"message":"branch is protected"}]}`, 0},
		{"conflict", http.StatusConflict, "file exists", "", "bitbucket answered 409: file exists", 0},
		{"long body is cut", http.StatusInternalServerError, strings.Repeat("x", 5000), "",
			"bitbucket answered 500: xxx", len("bitbucket answered 500: ") + 512},
		{"success without id", http.StatusOK, `{}`, "", "", 0},
		{"success with garbage", http.StatusOK, `not json`, "", "", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, _ := fakeServer(t, c.status, c.body)
			bb := &bitbucket{client: srv.Client()}

			id, err := bb.commitFile(srv.URL+"/rest/api/1.0/projects/K/repos/s", "t", "main", "f.txt", "m", []byte("x"))
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("got error %v, want it to contain %q", err, c.wantErr)
				}
				if c.maxErrLen > 0 && len(err.Error()) > c.maxErrLen {
					t.Fatalf("error is %d bytes, want at most %d", len(err.Error()), c.maxErrLen)
				}
			}
			if id != c.wantID {
				t.Fatalf("got id %q, want %q", id, c.wantID)
			}
		})
	}
}

// Обрезка тела по байтам может разрезать символ посередине. Текст ошибки
// уходит в Postgres, а он не принимает неверный UTF-8
func TestCommitFileErrorIsValidUTF8(t *testing.T) {
	srv, _ := fakeServer(t, http.StatusBadGateway, strings.Repeat("€", 400)) // 3 байта, 512 не кратно 3
	bb := &bitbucket{client: srv.Client()}

	_, err := bb.commitFile(srv.URL+"/rest/api/1.0/projects/K/repos/s", "t", "main", "f.txt", "m", []byte("x"))
	if err == nil {
		t.Fatal("expected an error")
	}
	if !utf8.ValidString(err.Error()) {
		t.Fatalf("error text is not valid UTF-8: %q", err.Error())
	}
}

func TestCommitFileTimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release // сервер молчит, пока тест не закончится
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(release) }) // выполняется раньше srv.Close
	bb := &bitbucket{client: &http.Client{Timeout: 50 * time.Millisecond}}

	start := time.Now()
	_, err := bb.commitFile(srv.URL+"/rest/api/1.0/projects/K/repos/s", "t", "main", "f.txt", "m", []byte("x"))
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("request took %v, the client timeout did not work", time.Since(start))
	}
}
