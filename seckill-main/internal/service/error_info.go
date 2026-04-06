package service

import (
	"fmt"
)

var (
	SUCCESS                           = 0
	ERR_INPUT_INVALID                 = 8020
	ERR_SHOULD_BIND                   = 8021
	ERR_JSON_MARSHAL                  = 8022
	ERR_FIND_GOODS_FAILED             = 8101
	ERR_GOODS_STOCK_NOT_ENOUGH        = 8102
	ERR_CREATE_ORDER_FAILED           = 8103
	ERR_CREATE_SECKILL_RECORD_FAILED  = 8104
	ERR_RECORD_USER_KILLED_NUM_FAILED = 8105
	ERR_DESC_STOCK_FAILED             = 8106
	ERR_CREATER_USER_QUOTA_FAILED     = 8107
	ERR_USER_QUOTA_NOT_ENOUGH         = 8108
	ERR_FIND_USER_QUOTA_FAILED        = 8109
)

var errMsgDic = map[int]string{
	SUCCESS:                           "success",
	ERR_INPUT_INVALID:                 "input invalid",
	ERR_SHOULD_BIND:                   "should bind failed",
	ERR_JSON_MARSHAL:                  "json marshal failed",
	ERR_FIND_GOODS_FAILED:             "商品查询失败",
	ERR_GOODS_STOCK_NOT_ENOUGH:        "商品库存不足",
	ERR_CREATE_ORDER_FAILED:           "订单创建失败",
	ERR_CREATE_SECKILL_RECORD_FAILED:  "秒杀记录创建失败",
	ERR_RECORD_USER_KILLED_NUM_FAILED: "记录用户已经秒杀到的名额数失败",
	ERR_DESC_STOCK_FAILED:             "库存扣减失败",
	ERR_CREATER_USER_QUOTA_FAILED:     "插入用户限额记录失败",
	ERR_USER_QUOTA_NOT_ENOUGH:         "用户额度不足",
	ERR_FIND_USER_QUOTA_FAILED:        "查询用户额度失败",
}

// GetErrMsg 获取错误描述
func GetErrMsg(code int) string {
	if msg, ok := errMsgDic[code]; ok {
		return msg
	}
	return fmt.Sprintf("unknown error code %d", code)
}
