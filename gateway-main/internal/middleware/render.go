package middleware

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteCodeMessage(w http.ResponseWriter, status int, code int, message string) {
	WriteJSON(w, status, map[string]interface{}{
		"code":    code,
		"message": message,
	})
}
