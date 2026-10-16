package domains

import (
	"bytes"
	"testing"

	"footprints/config"
)

func newTestAuth(t *testing.T) *Auth {
	t.Helper()
	a, err := NewAuth(nil, config.AuthSettings{EncKey: bytes.Repeat([]byte{7}, 32)})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestDecryptTokenRoundTrip(t *testing.T) {
	a := newTestAuth(t)
	enc := a.encrypt([]byte("secret-token"))
	if bytes.Contains(enc, []byte("secret-token")) {
		t.Fatal("token is stored in the open")
	}
	got, err := a.DecryptToken(enc)
	if err != nil {
		t.Fatal(err)
	}
	if got != "secret-token" {
		t.Fatalf("got %q", got)
	}
}

func TestDecryptTokenUsesRandomNonce(t *testing.T) {
	a := newTestAuth(t)
	if bytes.Equal(a.encrypt([]byte("x")), a.encrypt([]byte("x"))) {
		t.Fatal("two encryptions of one string are equal")
	}
}

func TestDecryptTokenRejects(t *testing.T) {
	a := newTestAuth(t)
	broken := a.encrypt([]byte("secret-token"))
	broken[len(broken)-1] ^= 1

	cases := map[string][]byte{
		"nil":       nil,
		"too short": {1, 2, 3},
		"tampered":  broken,
	}
	for name, enc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := a.DecryptToken(enc); err == nil {
				t.Fatal("expected an error")
			}
		})
	}

	other, err := NewAuth(nil, config.AuthSettings{EncKey: bytes.Repeat([]byte{9}, 32)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.DecryptToken(a.encrypt([]byte("x"))); err == nil {
		t.Fatal("a different key must not decrypt")
	}
}
