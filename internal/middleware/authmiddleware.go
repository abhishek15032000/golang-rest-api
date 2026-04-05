package middleware

import (
	"context"
	"net/http"
	"rest-api/internal/utils"
	"rest-api/redisconfig"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type contextKey string

const UserClaimsKey contextKey = "claims"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// retrieve the Authorization Header from request
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "No token provided")
			return // <-- was missing! Without this, code continues even with no token
		}
		// extract Bearer Token from it, so we split and the second index value is our thing
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token format")
			return // <-- was missing! Without this, parts[1] panics on bad format
		}
		token := parts[1]
		// check for blacklisted token before moving to the next function
		blacklisted, err := redisconfig.RedisClient.Get(r.Context(), token).Result()
		if err == nil && blacklisted == "blacklisted" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Token revoked")
			return
		} else if err != nil && err != redis.Nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check token status")
			return
		}
		claims, err := utils.ParseJWT(token)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return // <-- was missing! Without this, nil claims get passed to the handler
		}
		// check if the token has expired or not.
		if claims["exp"].(float64) < float64(time.Now().Unix()) {
			// set int redis as blacklisted token
			redisconfig.RedisClient.Set(r.Context(), token, "blacklisted", time.Duration(claims["exp"].(float64))*time.Second)
			utils.RespondWithError(w, http.StatusUnauthorized, "Token expired")
			return
		}
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
