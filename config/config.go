package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"footprints/models/db"
)

type Config struct {
	Task TaskSettings
	Auth AuthSettings
	PG   PGSettings
	API  APISettings
}

// TaskSettings это значения по умолчанию для полей таски, которые не пришли в запросе
type TaskSettings struct {
	Days          db.Weekdays
	StartTime     db.Clock
	EndTime       db.Clock
	CommitsPerDay int
	Active        bool
}

// Params возвращает заготовку таски со значениями по умолчанию. Срез дней
// копируется: encoding/json дописывает элементы в существующий срез и затёр бы
// значения в конфиге
func (s TaskSettings) Params() db.TaskParams {
	return db.TaskParams{
		Days:          slices.Clone(s.Days),
		StartTime:     s.StartTime,
		EndTime:       s.EndTime,
		CommitsPerDay: s.CommitsPerDay,
		Active:        s.Active,
	}
}

type AuthSettings struct {
	TokenTTL  time.Duration
	JWTSecret string // AUTH_JWT_SECRET
	EncKey    []byte // AUTH_ENC_KEY, 32 байта в hex
}

type PGSettings struct {
	Host            string        // PG_HOST
	Port            int           // PG_PORT
	User            string        // PG_USER
	Password        string        // PG_PASSWORD
	DB              string        // PG_DB
	ConnectAttempts int           // PG_CONNECT_ATTEMPTS
	ConnectDelay    time.Duration // PG_CONNECT_DELAY
}

// DSN собирает строку подключения в виде URL, чтобы спецсимволы в пароле экранировались
func (s PGSettings) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(s.User, s.Password),
		Host:   net.JoinHostPort(s.Host, strconv.Itoa(s.Port)),
		Path:   s.DB,
	}
	return u.String()
}

type APISettings struct {
	Host string // API_HOST
	Port int    // API_PORT
}

func (s APISettings) Addr() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}

// Load читает .env (если файла нет, переменные берутся из окружения) и
// накладывает их на значения по умолчанию
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Task: TaskSettings{
			Days:          db.Weekdays{1, 2, 3, 4, 5},
			StartTime:     "09:00",
			EndTime:       "18:00",
			CommitsPerDay: 1,
			Active:        true,
		},
		Auth: AuthSettings{TokenTTL: 7 * 24 * time.Hour},
		PG: PGSettings{
			Host:            "localhost",
			Port:            5432,
			User:            "footprints",
			DB:              "footprints",
			ConnectAttempts: 10,
			ConnectDelay:    3 * time.Second,
		},
		API: APISettings{Host: "0.0.0.0", Port: 8080},
	}

	var errs []error
	collect := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	envString("PG_HOST", &cfg.PG.Host)
	collect(envInt("PG_PORT", &cfg.PG.Port))
	envString("PG_USER", &cfg.PG.User)
	envString("PG_PASSWORD", &cfg.PG.Password)
	envString("PG_DB", &cfg.PG.DB)
	collect(envInt("PG_CONNECT_ATTEMPTS", &cfg.PG.ConnectAttempts))
	collect(envDuration("PG_CONNECT_DELAY", &cfg.PG.ConnectDelay))

	envString("API_HOST", &cfg.API.Host)
	collect(envInt("API_PORT", &cfg.API.Port))

	envString("AUTH_JWT_SECRET", &cfg.Auth.JWTSecret)

	if cfg.PG.Password == "" {
		errs = append(errs, errors.New("PG_PASSWORD is required"))
	}
	if cfg.PG.ConnectAttempts < 1 {
		errs = append(errs, errors.New("PG_CONNECT_ATTEMPTS must be at least 1"))
	}
	if cfg.Auth.JWTSecret == "" {
		errs = append(errs, errors.New("AUTH_JWT_SECRET is required"))
	}
	key, err := hex.DecodeString(os.Getenv("AUTH_ENC_KEY"))
	if err != nil || len(key) != 32 {
		errs = append(errs, errors.New("AUTH_ENC_KEY is required: 32 bytes in hex (64 characters)"))
	}
	cfg.Auth.EncKey = key

	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

func envString(key string, dst *string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

func envInt(key string, dst *int) error {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("%s: %q is not an integer", key, v)
	}
	*dst = n
	return nil
}

func envDuration(key string, dst *time.Duration) error {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("%s: %q is not a duration like 3s", key, v)
	}
	*dst = d
	return nil
}
