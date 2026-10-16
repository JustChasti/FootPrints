package core

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

const (
	minLines = 32 // число строк в файле коммита, границы включены
	maxLines = 64
	nameLen  = 12 // сколько символов хэша содержимого идёт в имя файла
)

// newFile делает файл для коммита: от minLines до maxLines строк случайных
// хэшей. Имя это обрезанный хэш содержимого, путь autocommit/<yyyy-mm-dd>/<имя>.txt,
// дата по времени сервера
func newFile(now time.Time) (path string, body []byte) {
	n := minLines + rand.IntN(maxLines-minLines+1)
	var b strings.Builder
	for range n {
		b.WriteString(randomHash())
		b.WriteByte('\n')
	}
	body = []byte(b.String())

	sum := sha256.Sum256(body)
	name := hex.EncodeToString(sum[:])[:nameLen]
	return fmt.Sprintf("autocommit/%s/%s.txt", now.Format("2006-01-02"), name), body
}

// randomHash это sha256 от случайных байт в hex, 64 символа
func randomHash() string {
	var seed [32]byte
	crand.Read(seed[:]) // не возвращает ошибку, при сбое источника процесс падает сам
	sum := sha256.Sum256(seed[:])
	return hex.EncodeToString(sum[:])
}
