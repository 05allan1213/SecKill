package handler

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/middleware"
)

func writeError(w http.ResponseWriter, err error) {
	if httpErr, ok := err.(*middleware.HTTPError); ok {
		middleware.WriteCodeMessage(w, httpErr.Status, httpErr.Status, httpErr.Message)
		return
	}

	middleware.WriteCodeMessage(w, http.StatusInternalServerError, http.StatusInternalServerError, err.Error())
}
