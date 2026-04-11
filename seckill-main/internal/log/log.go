package log

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/BitofferHub/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func InfoContextf(ctx context.Context, format string, v ...interface{}) {
	WithContext(ctx).Infof(withTracePrefix(ctx, format), prependTraceID(ctx, v...)...)
}

func ErrorContextf(ctx context.Context, format string, v ...interface{}) {
	WithContext(ctx).Errorf(withTracePrefix(ctx, format), prependTraceID(ctx, v...)...)
}

func WarnContextf(ctx context.Context, format string, v ...interface{}) {
	WithContext(ctx).Infof(withTracePrefix(ctx, format), prependTraceID(ctx, v...)...)
}

func Infof(format string, v ...interface{}) {
	logx.Infof(format, v...)
}

func Errorf(format string, v ...interface{}) {
	logx.Errorf(format, v...)
}

func WithContext(ctx context.Context) logx.Logger {
	return logx.WithContext(ctx)
}

func InfoContextw(ctx context.Context, msg string, fields ...logx.LogField) {
	WithContext(ctx).Infow(msg, append(traceField(ctx), fields...)...)
}

func ErrorContextw(ctx context.Context, msg string, fields ...logx.LogField) {
	WithContext(ctx).Errorw(msg, append(traceField(ctx), fields...)...)
}

func WarnContextw(ctx context.Context, msg string, fields ...logx.LogField) {
	WithContext(ctx).Sloww(msg, append(traceField(ctx), fields...)...)
}

type GormLogger struct {
	slowThreshold time.Duration
}

// NewGormLogger keeps the old package contract while staying inside the module.
func NewGormLogger(slowThresholdMillisecond int64) *GormLogger {
	return &GormLogger{
		slowThreshold: time.Duration(slowThresholdMillisecond) * time.Millisecond,
	}
}

func (l *GormLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	InfoContextf(ctx, s, i...)
}

func (l *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	WarnContextf(ctx, s, i...)
}

func (l *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	ErrorContextf(ctx, s, i...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	executeTime := time.Since(begin)
	sql, rows := fc()
	fields := []logx.LogField{
		logx.Field("sql", compactSQL(sql)),
		logx.Field("duration_ms", executeTime.Milliseconds()),
		logx.Field("rows", rows),
	}

	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			InfoContextw(ctx, "gorm record not found", fields...)
		} else {
			fields = append(fields, logx.Field("err", err))
			ErrorContextw(ctx, "gorm query failed", fields...)
		}
		return
	}

	if l.slowThreshold != 0 && executeTime > l.slowThreshold {
		WarnContextw(ctx, "gorm slow query", fields...)
		return
	}

	InfoContextw(ctx, "gorm query completed", fields...)
}

func traceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceValue := ctx.Value(constant.TraceID); traceValue != nil {
		if traceID, ok := traceValue.(string); ok {
			return traceID
		}
	}
	return ""
}

func compactSQL(sql string) string {
	return strings.Join(strings.Fields(sql), " ")
}

func traceField(ctx context.Context) []logx.LogField {
	traceID := traceIDFromContext(ctx)
	if traceID == "" {
		return nil
	}

	return []logx.LogField{logx.Field(constant.TraceID, traceID)}
}

func prependTraceID(ctx context.Context, args ...interface{}) []interface{} {
	return append([]interface{}{traceIDFromContext(ctx)}, args...)
}

func withTracePrefix(ctx context.Context, format string) string {
	if traceIDFromContext(ctx) == "" {
		return format
	}

	return fmt.Sprintf("%s:%%s, %s", constant.TraceID, format)
}
