package domains

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"footprints/config"
	"footprints/models/api"
	"footprints/models/db"
)

// bcrypt принимает пароль не длиннее 72 байт
const maxPasswordLen = 72

// Session это результат регистрации или входа: пользователь и его JWT
type Session struct {
	UserID    int64
	Login     string
	Token     string
	ExpiresAt time.Time
}

type Auth struct {
	db   *gorm.DB
	cfg  config.AuthSettings
	aead cipher.AEAD
}

func NewAuth(gdb *gorm.DB, cfg config.AuthSettings) (*Auth, error) {
	block, err := aes.NewCipher(cfg.EncKey)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Auth{db: gdb, cfg: cfg, aead: aead}, nil
}

func (a *Auth) Register(ctx context.Context, req api.RegisterRequest) (*Session, error) {
	login := strings.TrimSpace(req.Login)
	switch {
	case login == "":
		return nil, &ValidationError{"login is required"}
	case req.Password == "":
		return nil, &ValidationError{"password is required"}
	case len(req.Password) > maxPasswordLen:
		return nil, &ValidationError{"password is longer than 72 bytes"}
	case strings.TrimSpace(req.BitbucketToken) == "":
		return nil, &ValidationError{"bitbucket_token is required"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := db.User{
		Login:          login,
		PasswordHash:   string(hash),
		BitbucketToken: a.encrypt([]byte(req.BitbucketToken)),
	}
	if err := a.db.WithContext(ctx).Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrLoginTaken
		}
		return nil, err
	}
	return a.session(user)
}

func (a *Auth) Login(ctx context.Context, req api.LoginRequest) (*Session, error) {
	var user db.User
	err := a.db.WithContext(ctx).First(&user, "login = ?", strings.TrimSpace(req.Login)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrBadCredentials
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, ErrBadCredentials
	}
	return a.session(user)
}

// Parse проверяет подпись и срок JWT и возвращает id пользователя
func (a *Auth) Parse(token string) (int64, error) {
	claims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (any, error) {
		return []byte(a.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return 0, ErrUnauthorized
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, ErrUnauthorized
	}
	return id, nil
}

func (a *Auth) session(user db.User) (*Session, error) {
	now := time.Now()
	exp := now.Add(a.cfg.TokenTTL)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(user.ID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}).SignedString([]byte(a.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}
	return &Session{UserID: user.ID, Login: user.Login, Token: token, ExpiresAt: exp}, nil
}

// encrypt шифрует токен Bitbucket для хранения в базе (AES-GCM), случайный
// nonce лежит в начале результата. Запросы к Bitbucket сервис не шифрует, их
// защищает HTTPS
func (a *Auth) encrypt(plain []byte) []byte {
	nonce := make([]byte, a.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		panic(err) // crypto/rand не возвращает ошибку на поддерживаемых платформах
	}
	return a.aead.Seal(nonce, nonce, plain, nil)
}

// DecryptToken расшифровывает токен Bitbucket, сохранённый в базе функцией
// encrypt. Нужна ядру на время одного запроса к Bitbucket
func (a *Auth) DecryptToken(enc []byte) (string, error) {
	n := a.aead.NonceSize()
	if len(enc) < n {
		return "", errors.New("encrypted token is too short")
	}
	plain, err := a.aead.Open(nil, enc[:n], enc[n:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt token: %w", err)
	}
	return string(plain), nil
}
