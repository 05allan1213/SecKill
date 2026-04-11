package log

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogField = logx.LogField

const (
	AccessDetailSummary = "summary"
	AccessDetailRequest = "request"
	AccessDetailVerbose = "verbose"

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
	FieldRouteKey = "routeKey"
)

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

func logger(ctx context.Context) logx.Logger {
	if ctx == nil {
		ctx = context.Background()
	}
	return logx.WithContext(ctx).WithCallerSkip(1)
}
