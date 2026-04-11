package log

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type LogField = logx.LogField

const (
	AccessDetailSummary = "summary"
	AccessDetailRequest = "request"
	AccessDetailVerbose = "verbose"

	SQLModeSilent = "silent"
	SQLModeError  = "error"
	SQLModeSlow   = "slow"
	SQLModeAll    = "all"

	FieldSeverity = "severity"
	FieldTraceID  = "traceID"
	FieldAction   = "action"
	FieldUserID   = "userID"
	FieldGoodsNum = "goodsNum"
	FieldGoodsID  = "goodsID"
	FieldSecNum   = "secNum"
	FieldOrderNum = "orderNum"
	FieldCost     = "cost"
	FieldCode     = "code"
	FieldStatus   = "status"
	FieldError    = "err"
	FieldMethod   = "method"
	FieldPath     = "path"
	FieldRequest  = "request"
	FieldResponse = "response"
	FieldRows     = "rows"
	FieldSQL      = "sql"
	FieldRouteKey = "routeKey"
)

var sampledLogs sync.Map

type AccessEntry struct {
	Action   string
	Method   string
	Path     string
	Status   int
	Code     int
	Request  any
	Response any
	Err      error
	Cost     time.Duration
}

func Init(logPath string) {
	if logPath != "" {
		_ = os.MkdirAll(logPath, 0o755)
	}
}

func Field(key string, value any) LogField {
	return logx.Field(key, value)
}

func WithFields(ctx context.Context, fields ...LogField) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return logx.ContextWithFields(ctx, fields...)
}

func WithTrace(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return WithFields(ctx, Field(FieldTraceID, traceID))
}

func WithUser(ctx context.Context, userID any) context.Context {
	if userID == nil {
		return ctx
	}
	return WithFields(ctx, Field(FieldUserID, userID))
}

func WithAction(ctx context.Context, action string) context.Context {
	if action == "" {
		return ctx
	}
	return WithFields(ctx, Field(FieldAction, action))
}

func Info(ctx context.Context, msg string, fields ...LogField) {
	logger(ctx).Infow(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...LogField) {
	fields = append([]LogField{Field(FieldSeverity, "warn")}, fields...)
	logger(ctx).Infow(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...LogField) {
	logger(ctx).Errorw(msg, fields...)
}

func InfoEvery(ctx context.Context, key string, interval time.Duration, msg string, fields ...LogField) {
	if shouldLog(key, interval) {
		Info(ctx, msg, fields...)
	}
}

func WarnEvery(ctx context.Context, key string, interval time.Duration, msg string, fields ...LogField) {
	if shouldLog(key, interval) {
		Warn(ctx, msg, fields...)
	}
}

func InfoContextf(ctx context.Context, format string, v ...interface{}) {
	Info(ctx, fmt.Sprintf(format, v...))
}

func WarnContextf(ctx context.Context, format string, v ...interface{}) {
	Warn(ctx, fmt.Sprintf(format, v...))
}

func ErrorContextf(ctx context.Context, format string, v ...interface{}) {
	Error(ctx, fmt.Sprintf(format, v...))
}

func Infof(format string, v ...interface{}) {
	logx.Infow(fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) {
	logx.Errorw(fmt.Sprintf(format, v...))
}

func Access(ctx context.Context, detail string, entry AccessEntry) {
	fields := make([]LogField, 0, 8)
	if entry.Action != "" {
		fields = append(fields, Field(FieldAction, entry.Action))
	}
	if entry.Method != "" {
		fields = append(fields, Field(FieldMethod, entry.Method))
	}
	if entry.Path != "" {
		fields = append(fields, Field(FieldPath, entry.Path))
	}
	if entry.Status != 0 {
		fields = append(fields, Field(FieldStatus, entry.Status))
	}
	if entry.Code != 0 {
		fields = append(fields, Field(FieldCode, entry.Code))
	}
	if entry.Cost > 0 {
		fields = append(fields, Field(FieldCost, entry.Cost))
	}
	if detail != AccessDetailSummary && entry.Request != nil {
		fields = append(fields, Field(FieldRequest, entry.Request))
	}
	if entry.Response != nil {
		fields = append(fields, Field(FieldResponse, entry.Response))
	}
	if entry.Err != nil {
		fields = append(fields, Field(FieldError, entry.Err.Error()))
	}

	switch {
	case entry.Status >= 500:
		Error(ctx, "access", fields...)
	case entry.Status >= 400 || entry.Err != nil:
		Warn(ctx, "access", fields...)
	default:
		Info(ctx, "access", fields...)
	}
}

type GormLogger struct {
	mode          string
	slowThreshold time.Duration
}

func NewGormLogger(mode string, slowThresholdMillisecond int64) *GormLogger {
	mode = normalizeSQLMode(mode)
	if slowThresholdMillisecond < 0 {
		slowThresholdMillisecond = 0
	}
	return &GormLogger{
		mode:          mode,
		slowThreshold: time.Duration(slowThresholdMillisecond) * time.Millisecond,
	}
}

func (l *GormLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *GormLogger) Info(context.Context, string, ...interface{}) {}

func (l *GormLogger) Warn(context.Context, string, ...interface{}) {}

func (l *GormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	Error(ctx, fmt.Sprintf(msg, args...), Field(FieldAction, "sql.error"))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	executeTime := time.Since(begin)
	sql, rows := fc()
	fields := []LogField{
		Field(FieldSQL, sql),
		Field(FieldRows, rows),
		Field(FieldCost, executeTime),
	}

	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			if l.mode == SQLModeAll {
				Warn(ctx, "sql record not found", append(fields, Field(FieldAction, "sql.not_found"))...)
			}
			return
		}
		Error(ctx, "sql error", append(fields, Field(FieldAction, "sql.error"), Field(FieldError, err.Error()))...)
		return
	}

	switch l.mode {
	case SQLModeAll:
		if l.slowThreshold > 0 && executeTime >= l.slowThreshold {
			Info(ctx, "sql slow", append(fields, Field(FieldAction, "sql.slow"))...)
			return
		}
		Info(ctx, "sql query", append(fields, Field(FieldAction, "sql.query"))...)
	case SQLModeSlow:
		if l.slowThreshold > 0 && executeTime >= l.slowThreshold {
			Info(ctx, "sql slow", append(fields, Field(FieldAction, "sql.slow"))...)
		}
	}
}

func normalizeSQLMode(mode string) string {
	switch mode {
	case SQLModeSilent, SQLModeError, SQLModeSlow, SQLModeAll:
		return mode
	default:
		return SQLModeSlow
	}
}

func logger(ctx context.Context) logx.Logger {
	if ctx == nil {
		ctx = context.Background()
	}
	return logx.WithContext(ctx).WithCallerSkip(1)
}

func shouldLog(key string, interval time.Duration) bool {
	if key == "" || interval <= 0 {
		return true
	}
	now := time.Now().UnixNano()
	value, _ := sampledLogs.LoadOrStore(key, &atomic.Int64{})
	last := value.(*atomic.Int64)
	prev := last.Load()
	if prev != 0 && now-prev < interval.Nanoseconds() {
		return false
	}
	return last.CompareAndSwap(prev, now)
}
