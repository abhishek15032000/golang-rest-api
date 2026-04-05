package utils

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func RespondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: message,
		Data:    data,
	})
}

func RespondSuccessfullLogin(w http.ResponseWriter, code int, message string, data interface{}, cookie_data *http.Cookie) {
	w.Header().Set("Content-Type", "application/json")
	http.SetCookie(w, cookie_data)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: message,
		Data:    data,
	})
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Message: message,
	})
}

func RespondWithNotFound(w http.ResponseWriter) {
	RespondWithError(w, http.StatusNotFound, "Resource not found")
}
