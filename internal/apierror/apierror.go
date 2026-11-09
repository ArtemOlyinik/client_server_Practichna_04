package apierror

import (
	"encoding/json"
	"net/http"
	"time"
)

// Error - єдиний формат помилкової відповіді API.
type Error struct {
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func Write(w http.ResponseWriter, status int, code, message string) {
	payload := Error{
		Error:     code,
		Message:   message,
		Status:    status,
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
