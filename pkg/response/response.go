package response

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, Envelope{Success: true, Data: data})
}

func Fail(w http.ResponseWriter, status int, code, message string) {
	write(w, status, Envelope{
		Success: false,
		Error:   &Error{Code: code, Message: message},
	})
}

func FailDetails(w http.ResponseWriter, status int, code, message string, details map[string]string) {
	write(w, status, Envelope{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func write(w http.ResponseWriter, status int, body Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
