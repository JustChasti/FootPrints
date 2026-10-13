package api

import "time"

type RegisterRequest struct {
	Login          string `json:"login"           example:"alice"`
	Password       string `json:"password"        example:"s3cret-pass"`
	BitbucketToken string `json:"bitbucket_token" example:"BBDC-xxxxxxxx"`
}

type LoginRequest struct {
	Login    string `json:"login"    example:"alice"`
	Password string `json:"password" example:"s3cret-pass"`
}

// AuthResult это ответ на регистрацию и вход. Сам JWT уходит в cookie token
type AuthResult struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login"`
	ExpiresAt time.Time `json:"expires_at"`
}
