package handlers

import (
	"encoding/json"
	"net/http"
)

// APIResponse adalah struktur response JSON seragam untuk semua API routes
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError adalah struktur error yang user-friendly
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SendSuccess mengirim response sukses
func SendSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

// SendError mengirim response error
func SendError(w http.ResponseWriter, code string, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}
