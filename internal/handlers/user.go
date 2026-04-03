package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"rest-api/internal/dtos"
	"rest-api/internal/store"
	"rest-api/internal/utils"
	"rest-api/internal/validation"
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
