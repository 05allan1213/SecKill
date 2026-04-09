package log

import (
	"context"
	stderrors "errors"
	"os"
	"time"

	"github.com/BitofferHub/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func Init(logPath string) {
	if logPath != "" {
		_ = os.MkdirAll(logPath, 0o755)
	}
}

func InfoContextf(ctx context.Context, format string, v ...interface{}) {
	traceID := traceIDFromContext(ctx)
	logx.WithContext(ctx).Infof(constant.TraceID+":%s, "+format, append([]interface{}{traceID}, v...)...)
}

func ErrorContextf(ctx context.Context, format string, v ...interface{}) {
	traceID := traceIDFromContext(ctx)
	logx.WithContext(ctx).Errorf(constant.TraceID+":%s, "+format, append([]interface{}{traceID}, v...)...)
}

func WarnContextf(ctx context.Context, format string, v ...interface{}) {
	traceID := traceIDFromContext(ctx)
	logx.WithContext(ctx).Infof(constant.TraceID+":%s, "+format, append([]interface{}{traceID}, v...)...)
}

func Infof(format string, v ...interface{}) {
	logx.Infof(format, v...)
}

func Errorf(format string, v ...interface{}) {
	logx.Errorf(format, v...)
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

	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			InfoContextf(ctx, "Database ErrRecordNotFound, sql: %s, time: %s, rows: %d", sql, executeTime.String(), rows)
		} else {
			ErrorContextf(ctx, "Database Error sql: %s, time: %s, rows: %d, err: %v", sql, executeTime.String(), rows, err)
		}
		return
	}

	if l.slowThreshold != 0 && executeTime > l.slowThreshold {
		InfoContextf(ctx, "Database Slow Log sql: %s, time: %s, rows: %d", sql, executeTime.String(), rows)
		return
	}

	InfoContextf(ctx, "Database Query: %s, time: %s, rows: %d, err: %v", sql, executeTime.String(), rows, err)
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
