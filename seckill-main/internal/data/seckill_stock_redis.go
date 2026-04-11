package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BitofferHub/seckill/internal/log"
)

type PreSecKillStockRepo struct {
	data *Data
}

func NewPreSecKillStockRepo(data *Data) *PreSecKillStockRepo {
	return &PreSecKillStockRepo{
		data: data,
	}
}

func (r *PreSecKillStockRepo) PreDescStock(ctx context.Context, data *Data,
	userID int64, goodsID int64, num int32, secNum string, record *PreSecKillRecord) (string, error) {
	rdb := data.GetCache()
	keys := make([]string, 0)
	userIDStr := fmt.Sprintf("%d", userID)
	goodsIDStr := fmt.Sprintf("%d", goodsID)
	numStr := fmt.Sprintf("%d", num)
	secRecord, err := json.Marshal(record)
	if err != nil {
		log.ErrorContextf(ctx, "json marshal err ", err)
		return secNum, fmt.Errorf("marshal pre seckill record failed: %w", err)
	}
	log.InfoContextf(ctx, "secNum is %s, secRecord is %s",
		secNum, string(secRecord))
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
			log.InfoContextf(ctx, "already in seckill, secnum is %s", secNum)
			return secNum, err
		}
		return secNum, err
	}
	return secNum, err
}

func (r *PreSecKillStockRepo) SetSuccessInPreSecKill(ctx context.Context, data *Data,
	userID int64, goodsID int64, secNum string, record *PreSecKillRecord) (string, error) {
	rdb := data.GetCache()
	keys := make([]string, 0)
	userIDStr := fmt.Sprintf("%d", userID)
	goodsIDStr := fmt.Sprintf("%d", goodsID)
	secRecord, err := json.Marshal(record)
	if err != nil {
		log.ErrorContextf(ctx, "json marshal err ", err)
		return secNum, fmt.Errorf("marshal pre seckill record failed: %w", err)
	}
	log.InfoContextf(ctx, "secNum is %s, secRecord is %s",
		secNum, string(secRecord))
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
			log.InfoContextf(ctx, "already in seckill, secnum is %s", secNum)
			return secNum, err
		}
		return secNum, err
	}
	return secNum, err
}

func (r *PreSecKillStockRepo) GetSecKillInfo(ctx context.Context, data *Data, secNum string) (*PreSecKillRecord, error) {
	rdb := data.GetCache()
	var record = new(PreSecKillRecord)
	value, exist, err := rdb.Get(ctx, secNum)
	if err != nil {
		return record, err
	}
	if !exist {
		return record, err
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
