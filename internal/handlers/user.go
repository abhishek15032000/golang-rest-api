package handlers

import (
	"encoding/json"
	"net/http"
	"rest-api/internal/dtos"
	"rest-api/internal/store"
	"rest-api/internal/utils"
)

func (h *Handler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// create a context.
		ctx := r.Context()
		// user request;
		var req dtos.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error hashing password")
			return
		}
		_, err = h.Queries.CreateUser(ctx, store.CreateUserParams{
			Username: req.Username,
			Email:    req.Email,
			Password: hashedPassword,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error creating user")
			return
		}
		utils.RespondWithSuccess(w, http.StatusCreated, "User created successfully", req.Username)
	}
}
