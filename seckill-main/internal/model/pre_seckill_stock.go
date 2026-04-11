package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type PreSecKillRecord struct {
	SecNum     string
	UserID     int64
	GoodsID    int64
	OrderNum   string
	Price      float64
	Status     int
	CreateTime time.Time
	ModifyTime time.Time
}

type PreSecKillStockModel struct {
	store *Store
}

func NewPreSecKillStockModel(store *Store) *PreSecKillStockModel {
	return &PreSecKillStockModel{store: store}
}

func (m *PreSecKillStockModel) WithStore(store *Store) *PreSecKillStockModel {
	if store == nil {
		store = m.store
	}
	return &PreSecKillStockModel{store: store}
}

func (m *PreSecKillStockModel) PreDescStock(ctx context.Context, userID int64, goodsID int64,
	num int32, secNum string, record *PreSecKillRecord) (string, error) {
	rdb := m.store.GetCache()
	secRecord, err := marshalPreSecKillRecord(record)
	if err != nil {
		return secNum, fmt.Errorf("marshal pre seckill record failed: %w", err)
	}

	results, err := rdb.EvalResults(ctx, secKillLua, buildPreDescStockKeys(userID, goodsID, num, secNum, secRecord), []string{})
	if err != nil {
		return secNum, err
	}

	values := results.([]interface{})
	retCode := values[0].(int64)
	err = secKillRetCodeToError(int(retCode))
	if err != nil {
		if err.Error() == SecKillErrSecKilling.Error() {
			secNum = values[1].(string)
			return secNum, err
		}
		return secNum, err
	}
	return secNum, nil
}

func (m *PreSecKillStockModel) SetSuccessInPreSecKill(ctx context.Context, userID int64, goodsID int64,
	secNum string, record *PreSecKillRecord) (string, error) {
	rdb := m.store.GetCache()
	secRecord, err := marshalPreSecKillRecord(record)
	if err != nil {
		return secNum, fmt.Errorf("marshal pre seckill record failed: %w", err)
	}

	results, err := rdb.EvalResults(ctx, setSecKillSuccessLua, buildSetSuccessKeys(userID, goodsID, secNum, secRecord), []string{})
	if err != nil {
		return secNum, err
	}

	values := results.([]interface{})
	retCode := values[0].(int64)
	err = secKillRetCodeToError(int(retCode))
	if err != nil {
		if err.Error() == SecKillErrSecKilling.Error() {
			secNum = values[1].(string)
			return secNum, err
		}
		return secNum, err
	}
	return secNum, nil
}

func (m *PreSecKillStockModel) GetSecKillInfo(ctx context.Context, secNum string) (*PreSecKillRecord, error) {
	record := new(PreSecKillRecord)
	value, exist, err := m.store.GetCache().Get(ctx, secNum)
	if err != nil {
		return record, err
	}
	if !exist {
		return record, err
	}
	if err = json.Unmarshal([]byte(value), record); err != nil {
		return record, err
	}
	return record, nil
}

func marshalPreSecKillRecord(record *PreSecKillRecord) (string, error) {
	encoded, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func buildPreDescStockKeys(userID int64, goodsID int64, num int32, secNum string, encodedRecord string) []string {
	return []string{
		fmt.Sprintf("%d", userID),
		fmt.Sprintf("%d", goodsID),
		fmt.Sprintf("%d", num),
		secNum,
		encodedRecord,
	}
}

func buildSetSuccessKeys(userID int64, goodsID int64, secNum string, encodedRecord string) []string {
	return []string{
		fmt.Sprintf("%d", userID),
		fmt.Sprintf("%d", goodsID),
		secNum,
		encodedRecord,
	}
}

var (
	SecKillErrSecKilling        = errors.New("already in sec kill")
	SecKillErrUserGoodsOutLimit = errors.New("user out of limit on this goods")
	SecKillErrNotEnough         = errors.New("stock not enough")
	SecKillErrSelledOut         = errors.New("killed out")
)

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
redis.call('decrby',stockKey, KEYS[3])
redis.call('incrby', keyUserSecKilledNum, KEYS[3])
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
