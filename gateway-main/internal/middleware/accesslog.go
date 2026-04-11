package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	obs "github.com/BitofferHub/observability"
	"github.com/BitofferHub/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
)

type AccessLogMiddleware struct {
	summaryMaxBytes int
}

func NewAccessLogMiddleware(summaryMaxBytes int) *AccessLogMiddleware {
	if summaryMaxBytes <= 0 {
		summaryMaxBytes = 128
	}
	return &AccessLogMiddleware{summaryMaxBytes: summaryMaxBytes}
}

func (m *AccessLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqSummary := summarizeRequest(r, m.summaryMaxBytes)
		recorder := newLoggingResponseWriter(w, m.summaryMaxBytes)
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
	statusCode      int
	summaryMaxBytes int
	body            bytes.Buffer
	truncated       bool
}

func newLoggingResponseWriter(w http.ResponseWriter, summaryMaxBytes int) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter:  w,
		statusCode:      http.StatusOK,
		summaryMaxBytes: summaryMaxBytes,
	}
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggingResponseWriter) Write(p []byte) (int, error) {
	captureLimit := w.summaryMaxBytes + 1
	if w.body.Len() < captureLimit {
		remain := captureLimit - w.body.Len()
		if remain > len(p) {
			remain = len(p)
		}
		_, _ = w.body.Write(p[:remain])
		if remain < len(p) {
			w.truncated = true
		}
	} else if len(p) > 0 {
		w.truncated = true
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
	if w.truncated {
		return obs.TruncateString(body, w.summaryMaxBytes)
	}
	return obs.TruncateString(body, w.summaryMaxBytes)
}

func summarizeRequest(r *http.Request, summaryMaxBytes int) string {
	if r == nil {
		return "<nil>"
	}
	if r.URL.RawQuery != "" {
		return obs.TruncateString(r.URL.RawQuery, summaryMaxBytes)
	}
	if r.Body == nil {
		return "<empty>"
	}

	body, err := snapshotBody(r, summaryMaxBytes)
	if err != nil {
		return "<read_body_error>"
	}
	if len(body) == 0 {
		return "<empty>"
	}

	return obs.TruncateString(string(body), summaryMaxBytes)
}

func snapshotBody(r *http.Request, summaryMaxBytes int) ([]byte, error) {
	limited, err := io.ReadAll(io.LimitReader(r.Body, int64(summaryMaxBytes+1)))
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(limited), r.Body))
	return limited, nil
}
