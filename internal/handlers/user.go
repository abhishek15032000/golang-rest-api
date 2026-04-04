package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"rest-api/internal/dtos"
	"rest-api/internal/middleware"
	"rest-api/internal/store"
	"rest-api/internal/utils"
	"rest-api/internal/validation"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// login a user.

func (h *Handler) LoginUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req dtos.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		// validate the request
		if err := validation.Validate(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		user, err := h.Queries.GetUserByUsername(ctx, req.Username)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}

		if !utils.ComparePassword(req.Password, user.Password) {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}

		token, err := utils.GenerateJWT(int(user.ID), user.Username)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, "User logged in successfully", map[string]string{
			"token": token,
		})
	}
}

// createUser with transaction

func (h *Handler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dtos.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate the request
		if err := validation.Validate(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Start a transaction
		tx, err := h.DB.BeginTx(ctx, nil)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to start transaction")
			return
		}
		defer tx.Rollback() // Ensures rollback if any error occurs

		// Create a new Queries instance bound to the transaction
		qtx := store.New(tx)

		// Check if the username already exists
		_, err = qtx.GetUserByUsername(ctx, req.Username)
		if err == nil {
			utils.RespondWithError(w, http.StatusConflict, "Username already taken")
			return
		}

		// Check if the email already exists
		_, err = qtx.GetUserByEmail(ctx, req.Email)
		if err == nil {
			utils.RespondWithError(w, http.StatusConflict, "Email already taken")
			return
		}

		// Check if it's a real DB error or just "Not Found"
		if err != sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}

		// Hash password
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		// Create user within the transaction
		_, err = qtx.CreateUser(ctx, store.CreateUserParams{
			Username: req.Username,
			Email:    req.Email,
			Password: string(hashedPassword),
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Commit the transaction if all operations succeed
		if err := tx.Commit(); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
			return
		}

		utils.RespondWithSuccess(w, http.StatusCreated, "User created successfully", nil)
	}
}

func (h *Handler) GetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := r.Context().Value(middleware.UserClaimsKey).(jwt.MapClaims)
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "please login to continue")
			return
		}
		userID := int32(claims["user_id"].(float64))
		// check the redis first
		cacheKey := fmt.Sprintf("user:%d", userID)
		if cached, err := h.Redis.Get(ctx, cacheKey).Result(); err == nil {
			var user store.User
			if err := json.Unmarshal([]byte(cached), &user); err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to unmarshal cached data")
				return
			}
			utils.RespondWithSuccess(w, http.StatusOK, "User profile fetched successfully (success from cache/redis)", user)
			return
		}
		// fall back to db;
		user, err := h.Queries.GetUser(ctx, userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		// set in redis.
		userJSON, _ := json.Marshal(user) // converting into json string
		// marshal means - converting the struct into json string
		// unmarshal means - converting the json to struct
		h.Redis.Set(ctx, cacheKey, userJSON, 5*time.Minute)
		utils.RespondWithSuccess(w, http.StatusOK, "User profile fetched successfully", user)
	}
}

func (h *Handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. extract jwt claims from the request context
		ctx := r.Context()
		claims, ok := r.Context().Value(middleware.UserClaimsKey).(jwt.MapClaims)
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "please login to continue")
			return
		}
		// extract token from auth header
		tokenString := extractTokenFromHander(r)
		if tokenString == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "token not found")
			return
		}
		// 2. convert expiresAt to time.Time
		expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
		currentTIme := time.Now()
		ttl := expirationTime.Sub(currentTIme)
		if ttl <= 0 {
			ttl = 5 * time.Minute // fallback ttl
		}
		// 3. add token to blacklist in redis
		// it is setting tokenstring:blacklisted like this in redis
		if err := h.Redis.Set(ctx, tokenString, "blacklisted", ttl).Err(); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to blacklist token")
			return
		}
		// clean user cache in redis:-
		cacheKey := fmt.Sprintf("user:%d", int32(claims["user_id"].(float64)))
		if err := h.Redis.Del(ctx, cacheKey).Err(); err != nil {
			fmt.Printf("Error Cleaning user cache for %s in redis: %v\n", cacheKey, err)
		}
		utils.RespondWithSuccess(w, http.StatusOK, "User logged out successfully", nil)
	}
}

func extractTokenFromHander(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}
