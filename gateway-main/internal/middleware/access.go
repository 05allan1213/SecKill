package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gwlog "github.com/BitofferHub/gateway/internal/log"
)

type accessLogState struct {
	action   string
	userID   string
	request  any
	response any
	code     int
	err      error
}

type accessLogStateKey struct{}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

type AccessLogMiddleware struct {
	detail string
}

func NewAccessLogMiddleware(detail string) *AccessLogMiddleware {
	return &AccessLogMiddleware{detail: detail}
}

func (m *AccessLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := &accessLogState{}
		ctx := withAccessLogState(r.Context(), state)
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		begin := time.Now()

		next(recorder, r.WithContext(ctx))

		if state.userID != "" {
			ctx = gwlog.WithUser(ctx, state.userID)
		}
		gwlog.Access(ctx, m.detail, gwlog.AccessEntry{
			Action:   defaultAction(state.action, r.URL.Path),
			Method:   r.Method,
			Path:     r.URL.Path,
			Status:   recorder.status,
			Code:     defaultCode(state.code, recorder.status),
			Request:  state.request,
			Response: state.response,
			Err:      state.err,
			Cost:     time.Since(begin),
		})
	}
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func withAccessLogState(ctx context.Context, state *accessLogState) context.Context {
	return context.WithValue(ctx, accessLogStateKey{}, state)
}

func accessLogStateFromContext(ctx context.Context) *accessLogState {
	state, _ := ctx.Value(accessLogStateKey{}).(*accessLogState)
	return state
}

func RecordAccessAction(ctx context.Context, action string) {
	if state := accessLogStateFromContext(ctx); state != nil && action != "" {
		state.action = action
	}
}

func RecordAccessUser(ctx context.Context, userID string) {
	if state := accessLogStateFromContext(ctx); state != nil && userID != "" {
		state.userID = userID
	}
}

func RecordAccessRequest(ctx context.Context, request any) {
	if state := accessLogStateFromContext(ctx); state != nil {
		state.request = request
	}
}

func RecordAccessResponse(ctx context.Context, response any) {
	if state := accessLogStateFromContext(ctx); state != nil {
		state.response = response
	}
}

func RecordAccessCode(ctx context.Context, code int) {
	if state := accessLogStateFromContext(ctx); state != nil && code != 0 {
		state.code = code
	}
}

func RecordAccessError(ctx context.Context, err error) {
	if state := accessLogStateFromContext(ctx); state != nil && err != nil {
		state.err = err
	}
}

func NewAccessError(message string) error {
	return fmt.Errorf("%s", message)
}

func defaultAction(action string, path string) string {
	if action != "" {
		return action
	}
	return path
}

func defaultCode(code int, status int) int {
	if code != 0 {
		return code
	}
	return status
}
