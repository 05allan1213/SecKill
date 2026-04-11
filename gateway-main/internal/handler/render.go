package handler

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/middleware"
)

func writeError(r *http.Request, w http.ResponseWriter, err error) {
	middleware.RecordAccessError(r.Context(), err)
	if httpErr, ok := err.(*middleware.HTTPError); ok {
		middleware.RecordAccessCode(r.Context(), httpErr.Status)
		middleware.WriteCodeMessage(w, httpErr.Status, httpErr.Status, httpErr.Message)
		return
	}

	middleware.RecordAccessCode(r.Context(), http.StatusInternalServerError)
	middleware.WriteCodeMessage(w, http.StatusInternalServerError, http.StatusInternalServerError, err.Error())
}
