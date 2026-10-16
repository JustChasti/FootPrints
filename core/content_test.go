package core

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNewFile(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	pathRe := regexp.MustCompile(`^autocommit/2026-10-16/([0-9a-f]{12})\.txt$`)
	hashRe := regexp.MustCompile(`^[0-9a-f]{64}$`)

	seen := map[string]bool{}
	for range 100 {
		path, body := newFile(now)

		m := pathRe.FindStringSubmatch(path)
		if m == nil {
			t.Fatalf("unexpected path %q", path)
		}
		sum := sha256.Sum256(body)
		if want := hex.EncodeToString(sum[:])[:12]; m[1] != want {
			t.Fatalf("name %q is not the truncated hash of the body, want %q", m[1], want)
		}
		if seen[path] {
			t.Fatalf("path %q is repeated", path)
		}
		seen[path] = true

		text := string(body)
		if !strings.HasSuffix(text, "\n") {
			t.Fatal("file does not end with a newline")
		}
		lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
		if len(lines) < 32 || len(lines) > 64 {
			t.Fatalf("got %d lines, want 32..64", len(lines))
		}
		for _, l := range lines {
			if !hashRe.MatchString(l) {
				t.Fatalf("line %q is not a sha256 in hex", l)
			}
		}
	}
}
