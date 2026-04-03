package middleware

import (
	"context"
	"net/http"
	"rest-api/internal/utils"
	"strings"
)

type contextKey string

const UserClaimsKey contextKey = "claims"

func AuthMiddleware(next http.Handler) http.Handler {
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
		claims, err := utils.ParseJWT(token)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return // <-- was missing! Without this, nil claims get passed to the handler
		}
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
