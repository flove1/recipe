package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"flove/job/internal/user"
	"time"
)

type RefreshTokenModel struct {
	ID       string
	UserUUID string
	Token    string
	Expiry   time.Time
}

func NewRefreshToken(userUUID string, ttl time.Duration) (*RefreshTokenModel, error) {
	token := new(RefreshTokenModel)

	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	plaintext := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(plaintext))
	token.Token = string(base64.StdEncoding.EncodeToString(hash[:]))

	token.Expiry = time.Now().Add(ttl)
	token.UserUUID = userUUID

	return token, nil
}

func (t *RefreshTokenModel) IsExpired() bool {
	return t.Expiry.Before(time.Now())
}

type AccessTokenModel struct {
	UserUUID string
	Role     user.Role
	Token    string
}

func NewAccessToken(userUUID string, role user.Role) (*AccessTokenModel, error) {
	token := new(AccessTokenModel)

	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	plaintext := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(plaintext))
	token.Token = string(base64.StdEncoding.EncodeToString(hash[:]))

	token.UserUUID = userUUID
	token.Role = role

	return token, nil
}
