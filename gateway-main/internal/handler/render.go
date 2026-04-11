package handler

import (
	"errors"
	"net/http"

	"github.com/BitofferHub/gateway/internal/middleware"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func writeError(r *http.Request, w http.ResponseWriter, err error) {
	middleware.RecordAccessError(r.Context(), err)
	httpErr := normalizeHTTPError(err)
	middleware.RecordAccessCode(r.Context(), httpErr.Status)
	middleware.WriteCodeMessage(w, httpErr.Status, httpErr.Status, httpErr.Message)
}

func normalizeHTTPError(err error) *middleware.HTTPError {
	if err == nil {
		return &middleware.HTTPError{Status: http.StatusInternalServerError, Message: "internal server error"}
	}

	var httpErr *middleware.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}

	if st, ok := grpcstatus.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			return &middleware.HTTPError{Status: http.StatusBadRequest, Message: "invalid request"}
		case codes.Unauthenticated:
			return &middleware.HTTPError{Status: http.StatusUnauthorized, Message: "no authentication"}
		case codes.NotFound:
			return &middleware.HTTPError{Status: http.StatusNotFound, Message: "resource not found"}
		case codes.DeadlineExceeded, codes.Unavailable:
			return &middleware.HTTPError{Status: http.StatusServiceUnavailable, Message: "service unavailable"}
		default:
			return &middleware.HTTPError{Status: http.StatusInternalServerError, Message: "internal server error"}
		}
	}

	return &middleware.HTTPError{Status: http.StatusInternalServerError, Message: "internal server error"}
}
