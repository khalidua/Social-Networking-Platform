package apiresponse

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type SuccessEnvelope struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type ErrorEnvelope struct {
	Success   bool      `json:"success"`
	Error     ErrorBody `json:"error"`
	RequestID string    `json:"request_id,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Success(w http.ResponseWriter, status int, requestID string, data interface{}, message string) {
	resp := SuccessEnvelope{
		Success:   true,
		Data:      data,
		Message:   message,
		RequestID: requestID,
	}
	JSON(w, status, resp)
}

func Error(w http.ResponseWriter, status int, requestID string, code string, message string, details interface{}) {
	resp := ErrorEnvelope{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: requestID,
	}
	JSON(w, status, resp)
}