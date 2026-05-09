package apiresponse

import (
	"encoding/json"
	"net/http"
)

type SuccessEnvelope struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type ErrorEnvelope struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Success(w http.ResponseWriter, status int, requestID string, data interface{}, message string) {
	JSON(w, status, SuccessEnvelope{
		Success:   true,
		Data:      data,
		Message:   message,
		RequestID: requestID,
	})
}

func Error(w http.ResponseWriter, status int, code string, message string) {
	JSON(w, status, ErrorEnvelope{
		Error:   code,
		Message: message,
		Status:  status,
	})
}
