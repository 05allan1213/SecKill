package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BitofferHub/seckill/internal/log"
	"time"
)

type PreSecKillStockRepo struct {
	data *Data
}

var ErrPreSecKillInfoNotFound = errors.New("pre seckill info not found")

func NewPreSecKillStockRepo(data *Data) *PreSecKillStockRepo {
	return &PreSecKillStockRepo{
		data: data,
	}
}

func (r *PreSecKillStockRepo) PreDescStock(ctx context.Context, data *Data,
	userID int64, goodsID int64, num int32, secNum string, record *PreSecKillRecord) (string, error) {
	ctx = log.WithUser(ctx, userID)
	ctx = log.WithFields(ctx,
		log.Field(log.FieldGoodsID, goodsID),
		log.Field(log.FieldSecNum, secNum),
	)
	rdb := data.GetCache()
	keys := make([]string, 0)
	userIDStr := fmt.Sprintf("%d", userID)
	goodsIDStr := fmt.Sprintf("%d", goodsID)
	numStr := fmt.Sprintf("%d", num)
	secRecord, err := json.Marshal(record)
	if err != nil {
		log.Error(ctx, "pre-seckill record marshal failed",
			log.Field(log.FieldAction, "seckill.pre_desc_stock"),
			log.Field(log.FieldError, err.Error()),
		)
		return secNum, fmt.Errorf("marshal pre seckill record failed: %w", err)
	}
	keys = append(keys, userIDStr, goodsIDStr, numStr, secNum, string(secRecord))
	results, err := rdb.EvalResults(ctx, secKillLua, keys, []string{})
	if err != nil {
		return secNum, err
	}
	values := results.([]interface{})
	retCode := values[0].(int64)
	err = secKillRetCodeToError(int(retCode))
	if err != nil {
		if err.Error() == SecKillErrSecKilling.Error() {
			secNum = values[1].(string)
			log.WarnEvery(ctx, "seckill.pre_desc_stock.duplicate", 2*time.Second, "user already in seckill",
				log.Field(log.FieldAction, "seckill.pre_desc_stock"),
				log.Field(log.FieldSecNum, secNum),
			)
			return secNum, err
		}
		return secNum, err
	}
	return secNum, err
}

func (r *PreSecKillStockRepo) SetSuccessInPreSecKill(ctx context.Context, data *Data,
	userID int64, goodsID int64, secNum string, record *PreSecKillRecord) (string, error) {
	ctx = log.WithUser(ctx, userID)
	ctx = log.WithFields(ctx,
		log.Field(log.FieldGoodsID, goodsID),
		log.Field(log.FieldSecNum, secNum),
	)
	rdb := data.GetCache()
	keys := make([]string, 0)
	userIDStr := fmt.Sprintf("%d", userID)
	goodsIDStr := fmt.Sprintf("%d", goodsID)
	secRecord, err := json.Marshal(record)
	if err != nil {
		log.Error(ctx, "pre-seckill success record marshal failed",
			log.Field(log.FieldAction, "seckill.set_success"),
			log.Field(log.FieldError, err.Error()),
		)
		return secNum, fmt.Errorf("marshal pre seckill record failed: %w", err)
	}
	keys = append(keys, userIDStr, goodsIDStr, secNum, string(secRecord))
	results, err := rdb.EvalResults(ctx, setSecKillSuccessLua, keys, []string{})
	if err != nil {
		return secNum, err
	}
	values := results.([]interface{})
	retCode := values[0].(int64)
	err = secKillRetCodeToError(int(retCode))
	if err != nil {
		if err.Error() == SecKillErrSecKilling.Error() {
			secNum = values[1].(string)
			log.WarnEvery(ctx, "seckill.set_success.duplicate", 2*time.Second, "user already in seckill",
				log.Field(log.FieldAction, "seckill.set_success"),
				log.Field(log.FieldSecNum, secNum),
			)
			return secNum, err
		}
		return secNum, err
	}
	return secNum, err
}

func (r *PreSecKillStockRepo) SetFailedInPreSecKill(ctx context.Context, data *Data,
	userID int64, goodsID int64, num int32, secNum string, record *PreSecKillRecord) (string, error) {
	ctx = log.WithUser(ctx, userID)
	ctx = log.WithFields(ctx,
		log.Field(log.FieldGoodsID, goodsID),
		log.Field(log.FieldSecNum, secNum),
	)
	rdb := data.GetCache()
	userIDStr := fmt.Sprintf("%d", userID)
	goodsIDStr := fmt.Sprintf("%d", goodsID)
	numStr := fmt.Sprintf("%d", num)
	recordJSON, err := json.Marshal(record)
	if err != nil {
		log.Error(ctx, "pre-seckill failed record marshal failed",
			log.Field(log.FieldAction, "seckill.set_failed"),
			log.Field(log.FieldError, err.Error()),
		)
		return secNum, fmt.Errorf("marshal pre seckill record failed: %w", err)
	}

	results, err := rdb.EvalResults(ctx, setSecKillFailedLua, []string{
		userIDStr,
		goodsIDStr,
		numStr,
		secNum,
		string(recordJSON),
	})
	if err != nil {
		return secNum, err
	}

	values := results.([]interface{})
	retCode := values[0].(int64)
	if err = secKillRetCodeToError(int(retCode)); err != nil {
		return secNum, err
	}
	return secNum, nil
}

func (r *PreSecKillStockRepo) GetSecKillInfo(ctx context.Context, data *Data, secNum string) (*PreSecKillRecord, error) {
	rdb := data.GetCache()
	var record = new(PreSecKillRecord)
	value, exist, err := rdb.Get(ctx, secNum)
	if err != nil {
		return record, err
	}
	if !exist {
		return record, ErrPreSecKillInfoNotFound
	}
	err = json.Unmarshal([]byte(value), record)
	if err != nil {
		return record, err
	}
	return record, err
}

var (
	SecKillErrSecKilling        = errors.New("already in sec kill")
	SecKillErrUserGoodsOutLimit = errors.New("user out of limit on this goods")
	SecKillErrNotEnough         = errors.New("stock not enough")
	SecKillErrSelledOut         = errors.New("killed out")
)

// SecKillRetCodeToError func get code to err
func secKillRetCodeToError(retCode int) error {
	switch retCode {
	case -1:
		return SecKillErrSecKilling
	case -2:
		return SecKillErrUserGoodsOutLimit
	case -3:
		return SecKillErrNotEnough
	case -4:
		return SecKillErrSelledOut
	}
	return nil
}

// userGoodsSecNum key: userid+goodsid, value:secNum
// keyUserSecKilledNum: key: userid+goodsid, value:killedNum

var secKillLua = `
-- key1：用户id，key2：商品id key3：抢购多少个 key4：秒杀单号, keys5:秒杀记录
-- keyLimit是 SK:Limit:goodsID
local keyLimit = "SK:Limit" .. KEYS[2]
-- keyUserGoodsSecNum 是 SK:UserGoodsSecNum:goodsID:userID
local keyUserGoodsSecNum = "SK:UserGoodsSecNum:" .. KEYS[1] .. ":" .. KEYS[2]
-- keyUserSecKilledNum 是SK:UserSecKilledNum:userID:goodsID
local keyUserSecKilledNum = "SK:UserSecKilledNum:" .. KEYS[1] .. ":" .. KEYS[2]

--1.判断这个用户是不是已经在秒杀中，是的话返回secNum
local alreadySecNum = redis.call('get', keyUserGoodsSecNum)

local retAry = {0, ""}
if alreadySecNum and string.len(alreadySecNum) ~= 0 then
   retAry[1] = -1
   retAry[2] = alreadySecNum
   return retAry
end

--2.判断这个用户是不是已经超过限额了
local limit = redis.call('get', keyLimit)
local userSecKilledNum  = redis.call('get', keyUserSecKilledNum)
if limit and userSecKilledNum and tonumber(userSecKilledNum) + tonumber(KEYS[3]) > tonumber(limit) then 
   retAry[1] = -2
   return retAry
end

--3.判断查询活动库存
local stockKey = "SK:Stock:" .. KEYS[2]
local stock = redis.call('get', stockKey)
if not stock or tonumber(stock) < tonumber(KEYS[3]) then
   retAry[1] = -3
   return retAry
end

-- 4.活动库存充足，进行扣减操作
redis.call('decrby',stockKey, KEYS[3])
redis.call('incrby', keyUserSecKilledNum, KEYS[3])
redis.call('set', keyUserGoodsSecNum, KEYS[4]) 
redis.call('set', KEYS[4], KEYS[5]) 
return retAry
`

var userRebackStockLua = `
-- key1：用户id，key2：商品id key3：抢购多少个 key4：秒杀单号, keys5:秒杀记录
-- keyLimit是 SK:Limit:goodsID
local keyLimit = "SK:Limit" .. KEYS[2]
-- keyUserGoodsSecNum 是 SK:UserGoodsSecNum:goodsID:userID
local keyUserGoodsSecNum = "SK:UserGoodsSecNum:" .. KEYS[1] .. ":" .. KEYS[2]
-- keyUserSecKilledNum 是SK:UserSecKilledNum:userID:goodsID
local keyUserSecKilledNum = "SK:UserSecKilledNum:" .. KEYS[1] .. ":" .. KEYS[2]

--1.判断这个用户是不是已经在秒杀中，是的话返回secNum
local alreadySecNum = redis.call('get', keyUserGoodsSecNum)

local retAry = {0, ""}
if alreadySecNum and string.len(alreadySecNum) ~= 0 then
   retAry[1] = -1
   retAry[2] = alreadySecNum
   return retAry
end

--2.判断这个用户是不是已经超过限额了
local limit = redis.call('get', keyLimit)
local userSecKilledNum  = redis.call('get', keyUserSecKilledNum)
if limit and userSecKilledNum and tonumber(userSecKilledNum) + tonumber(KEYS[3]) > tonumber(limit) then 
   retAry[1] = -2
   return retAry
end

--3.判断查询活动库存
local stockKey = "SK:Stock:" .. KEYS[2]
local stock = redis.call('get', stockKey)
if not stock or tonumber(stock) < tonumber(KEYS[3]) then
   retAry[1] = -3
   return retAry
end

-- 4.活动库存充足，进行扣减操作
redis.call('incrby',stockKey, KEYS[3])
redis.call('decrby', keyUserSecKilledNum, KEYS[3])
redis.call('set', keyUserGoodsSecNum, KEYS[4]) 
redis.call('set', KEYS[4], KEYS[5]) 
return retAry
`

var setSecKillSuccessLua = `
-- key1：用户id，key2：商品id key3：秒杀单号, keys4:秒杀记录
-- keyUserGoodsSecNum 是 SK:UserGoodsSecNum:goodsID:userID
local keyUserGoodsSecNum = "SK:UserGoodsSecNum:" .. KEYS[1] .. ":" .. KEYS[2]
local retAry = {0, ""}
redis.call('set', keyUserGoodsSecNum, "") 
redis.call('set', KEYS[3], KEYS[4]) 
return retAry
`

var setSecKillFailedLua = `
-- key1：用户id，key2：商品id key3：抢购数量 key4：秒杀单号 key5：失败后的秒杀记录
local stockKey = "SK:Stock:" .. KEYS[2]
local keyUserGoodsSecNum = "SK:UserGoodsSecNum:" .. KEYS[1] .. ":" .. KEYS[2]
local keyUserSecKilledNum = "SK:UserSecKilledNum:" .. KEYS[1] .. ":" .. KEYS[2]

local currentSecNum = redis.call('get', keyUserGoodsSecNum)
if currentSecNum and currentSecNum == KEYS[4] then
   redis.call('set', keyUserGoodsSecNum, "")
   if redis.call('exists', stockKey) == 1 then
      redis.call('incrby', stockKey, KEYS[3])
   end

   local userKilledNum = redis.call('get', keyUserSecKilledNum)
   if userKilledNum then
      local nextNum = tonumber(userKilledNum) - tonumber(KEYS[3])
      if nextNum > 0 then
         redis.call('set', keyUserSecKilledNum, nextNum)
      else
         redis.call('set', keyUserSecKilledNum, 0)
      end
   end
end

redis.call('set', KEYS[4], KEYS[5])
local retAry = {0, ""}
return retAry
`
