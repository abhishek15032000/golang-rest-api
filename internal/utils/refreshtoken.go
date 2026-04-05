package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"rest-api/internal/store"
	"time"
)

type Refresh_token struct {
	Sub    int       `json:"sub"`
	Token  string    `json:"token"`
	Expiry time.Time `json:"expiry"`
}

func Generaterefreshtoken(sub int32, qtx *store.Queries, ctx *context.Context) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)

	hash := sha256.Sum256([]byte(token))

	// encode hash before storing
	hashStr := hex.EncodeToString(hash[:])

	// but here i also have to insert the hash along with user_id for lookup.
	_, err = qtx.InsertRefreshToken(*ctx, store.InsertRefreshTokenParams{
		UserID:    sub,
		TokenHash: hashStr,
	})
	if err != nil {
		return "could not insert the hash into token", err
	}
	return hashStr, nil
}
