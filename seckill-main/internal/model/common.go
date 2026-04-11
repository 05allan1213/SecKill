package model

import "github.com/zeromicro/go-zero/core/logx"

func bitlogField(key string, value interface{}) logx.LogField {
	return logx.Field(key, value)
}
