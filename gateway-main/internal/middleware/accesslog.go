package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/BitofferHub/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
)

const maxLoggedBodySize = 512

type AccessLogMiddleware struct{}

func NewAccessLogMiddleware() *AccessLogMiddleware {
	return &AccessLogMiddleware{}
}

func (m *AccessLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqSummary := summarizeRequest(r)
		recorder := newLoggingResponseWriter(w)
		begin := time.Now()

		next(recorder, r)

		statusCode := recorder.StatusCode()
		fields := []logx.LogField{
			logx.Field(constant.TraceID, TraceIDFromContext(r.Context())),
			logx.Field("method", r.Method),
			logx.Field("path", r.URL.Path),
			logx.Field("status", statusCode),
			logx.Field("duration_ms", time.Since(begin).Milliseconds()),
			logx.Field("user_id", UserIDFromContext(r.Context())),
			logx.Field("req_summary", reqSummary),
			logx.Field("reply_summary", recorder.BodySummary()),
		}

		if statusCode >= http.StatusBadRequest {
			logx.WithContext(r.Context()).Errorw("http request failed", fields...)
			return
		}

		logx.WithContext(r.Context()).Infow("http request completed", fields...)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggingResponseWriter) Write(p []byte) (int, error) {
	if w.body.Len() < maxLoggedBodySize {
		remain := maxLoggedBodySize - w.body.Len()
		if remain > len(p) {
			remain = len(p)
		}
		_, _ = w.body.Write(p[:remain])
	}

	return w.ResponseWriter.Write(p)
}

func (w *loggingResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *loggingResponseWriter) BodySummary() string {
	body := strings.TrimSpace(w.body.String())
	if body == "" {
		return "<empty>"
	}
	return body
}

func summarizeRequest(r *http.Request) string {
	if r == nil {
		return "<nil>"
	}
	if r.URL.RawQuery != "" {
		return r.URL.RawQuery
	}
	if r.Body == nil {
		return "<empty>"
	}

	body, truncated, err := snapshotBody(r)
	if err != nil {
		return "<read_body_error>"
	}
	if len(body) == 0 {
		return "<empty>"
	}
	if truncated {
		return string(body) + "...(truncated)"
	}
	return string(body)
}

func snapshotBody(r *http.Request) ([]byte, bool, error) {
	limited, err := io.ReadAll(io.LimitReader(r.Body, maxLoggedBodySize+1))
	if err != nil {
		return nil, false, err
	}
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(limited), r.Body))

	if len(limited) > maxLoggedBodySize {
		return limited[:maxLoggedBodySize], true, nil
	}
	return limited, false, nil
}
