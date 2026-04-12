package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/BitofferHub/seckill/internal/log"
	"github.com/redis/go-redis/v9"
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

// GetSecKillInfoBatch 批量获取秒杀信息，使用 Pipeline 优化
func (r *PreSecKillStockRepo) GetSecKillInfoBatch(ctx context.Context, data *Data, secNums []string) (map[string]*PreSecKillRecord, error) {
	if len(secNums) == 0 {
		return make(map[string]*PreSecKillRecord), nil
	}

	// 使用原生 Redis 客户端的 Pipeline
	redisClient := data.GetRedisClient()
	if redisClient == nil {
		// 降级为逐个获取
		return r.getSecKillInfoFallback(ctx, data, secNums)
	}

	// 构建 Pipeline 命令
	pipe := redisClient.Pipeline()
	cmds := make(map[string]*redis.StringCmd, len(secNums))

	for _, secNum := range secNums {
		cmds[secNum] = pipe.Get(ctx, secNum)
	}

	// 执行 Pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("pipeline exec failed: %w", err)
	}

	// 解析结果
	results := make(map[string]*PreSecKillRecord, len(secNums))
	for secNum, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				continue // Key 不存在，跳过
			}
			continue // 其他错误，跳过
		}

		var record PreSecKillRecord
		if err := json.Unmarshal([]byte(val), &record); err == nil {
			results[secNum] = &record
		}
	}

	return results, nil
}

// getSecKillInfoFallback 降级逐个获取
func (r *PreSecKillStockRepo) getSecKillInfoFallback(ctx context.Context, data *Data, secNums []string) (map[string]*PreSecKillRecord, error) {
	results := make(map[string]*PreSecKillRecord, len(secNums))
	for _, secNum := range secNums {
		record, err := r.GetSecKillInfo(ctx, data, secNum)
		if err == nil && record != nil {
			results[secNum] = record
		}
	}
	return results, nil
}

// BatchSetSecKillRecords 批量设置秒杀记录，使用 Pipeline 优化
func (r *PreSecKillStockRepo) BatchSetSecKillRecords(ctx context.Context, data *Data, records map[string]*PreSecKillRecord, ttl time.Duration) error {
	if len(records) == 0 {
		return nil
	}

	redisClient := data.GetRedisClient()
	if redisClient == nil {
		// 降级为逐个设置
		return r.batchSetSecKillRecordsFallback(ctx, data, records, ttl)
	}

	pipe := redisClient.Pipeline()

	for secNum, record := range records {
		recordJSON, err := json.Marshal(record)
		if err != nil {
			continue
		}
		pipe.Set(ctx, secNum, recordJSON, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// batchSetSecKillRecordsFallback 降级逐个设置
func (r *PreSecKillStockRepo) batchSetSecKillRecordsFallback(ctx context.Context, data *Data, records map[string]*PreSecKillRecord, ttl time.Duration) error {
	// 使用原生 Redis 客户端
	redisClient := data.GetRedisClient()
	if redisClient != nil {
		for secNum, record := range records {
			recordJSON, err := json.Marshal(record)
			if err != nil {
				continue
			}
			if err := redisClient.Set(ctx, secNum, recordJSON, ttl).Err(); err != nil {
				return err
			}
		}
		return nil
	}

	// 如果没有原生客户端，尝试使用 cache 包
	rdb := data.GetCache()
	for secNum, record := range records {
		recordJSON, err := json.Marshal(record)
		if err != nil {
			continue
		}
		// cache 包可能不支持 SetEX，使用 Eval 替代
		setScript := "return redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[2])"
		if _, err := rdb.EvalResults(ctx, setScript, []string{secNum}, []string{string(recordJSON), fmt.Sprintf("%d", int(ttl.Seconds()))}); err != nil {
			return err
		}
	}
	return nil
}

// BatchCheckStock 批量检查库存，使用 Pipeline 优化
func (r *PreSecKillStockRepo) BatchCheckStock(ctx context.Context, data *Data, goodsIDs []int64) (map[int64]int64, error) {
	if len(goodsIDs) == 0 {
		return make(map[int64]int64), nil
	}

	redisClient := data.GetRedisClient()
	if redisClient == nil {
		return nil, errors.New("redis client not available")
	}

	pipe := redisClient.Pipeline()
	cmds := make(map[int64]*redis.StringCmd, len(goodsIDs))

	for _, goodsID := range goodsIDs {
		stockKey := fmt.Sprintf("SK:Stock:%d", goodsID)
		cmds[goodsID] = pipe.Get(ctx, stockKey)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("pipeline exec failed: %w", err)
	}

	results := make(map[int64]int64, len(goodsIDs))
	for goodsID, cmd := range cmds {
		val, err := cmd.Int64()
		if err == nil {
			results[goodsID] = val
		}
	}

	return results, nil
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
-- secKillLua: Redis 预扣脚本
-- KEYS[1] = userID
-- KEYS[2] = goodsID
-- KEYS[3] = 秒杀数量
-- KEYS[4] = secNum
-- KEYS[5] = 预秒杀记录 JSON
--
-- 相关 key:
--   SK:Limit{goodsID}                 商品限购阈值
--   SK:UserGoodsSecNum:{userID}:{goodsID}   用户当前占用中的 secNum
--   SK:UserSecKilledNum:{userID}:{goodsID}  用户已经预扣/成功的数量
--   SK:Stock:{goodsID}                Redis 热点库存
--   {secNum}                          预秒杀记录
--
-- 返回:
--   {0, ""}     预扣成功
--   {-1, secNum} 用户已有进行中的秒杀请求
--   {-2, ""}    用户超出限购
--   {-3, ""}    Redis 库存不足
--
-- 流程:
--   1. 检查用户是否已有进行中的 secNum，占位存在则直接返回旧 secNum。
--   2. 检查用户累计秒杀数量是否超过商品限额。
--   3. 检查 Redis 热点库存是否足够。
--   4. 一次性完成库存预扣、用户已秒杀数量递增、进行中占位写入、预秒杀记录写入。
-- keyLimit 是 SK:Limit:goodsID
local keyLimit = "SK:Limit" .. KEYS[2]
-- keyUserGoodsSecNum 是 SK:UserGoodsSecNum:userID:goodsID
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
-- userRebackStockLua: 旧版“回退库存”脚本
-- 当前主流程已经统一改为 setSecKillFailedLua 做失败写回 + 回补。
-- 这个脚本保留的原因是方便对照早期实现，不参与现在的 V2 / V3 主链路。
--
-- KEYS[1] = userID
-- KEYS[2] = goodsID
-- KEYS[3] = 秒杀数量
-- KEYS[4] = secNum
-- KEYS[5] = 回补后的记录 JSON
--
-- 注意:
--   1. 这个脚本名字虽然叫“reback”，但它的写回行为保留了旧版占位逻辑。
--   2. 新版本失败闭环请看 setSecKillFailedLua 的注释和实现。
-- keyLimit 是 SK:Limit:goodsID
local keyLimit = "SK:Limit" .. KEYS[2]
-- keyUserGoodsSecNum 是 SK:UserGoodsSecNum:userID:goodsID
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
-- setSecKillSuccessLua: 秒杀成功写回脚本
-- KEYS[1] = userID
-- KEYS[2] = goodsID
-- KEYS[3] = secNum
-- KEYS[4] = 成功后的预秒杀记录 JSON
--
-- 流程:
--   1. 清空“用户进行中秒杀”占位，允许后续新的请求进入。
--   2. 用最终成功记录覆盖 secNum 对应的预秒杀记录。
--
-- 这里不会动库存或用户累计数量，因为这两个值已经在预扣阶段扣掉，
-- 成功场景只需要把“进行中”切换为“成功”即可。
local keyUserGoodsSecNum = "SK:UserGoodsSecNum:" .. KEYS[1] .. ":" .. KEYS[2]
local retAry = {0, ""}
redis.call('set', keyUserGoodsSecNum, "") 
redis.call('set', KEYS[3], KEYS[4]) 
return retAry
`

var setSecKillFailedLua = `
-- setSecKillFailedLua: 失败写回 + Redis 回补脚本
-- KEYS[1] = userID
-- KEYS[2] = goodsID
-- KEYS[3] = 秒杀数量
-- KEYS[4] = secNum
-- KEYS[5] = 失败后的预秒杀记录 JSON
--
-- 相关 key:
--   SK:Stock:{goodsID}                      Redis 热点库存
--   SK:UserGoodsSecNum:{userID}:{goodsID}  用户进行中 secNum 占位
--   SK:UserSecKilledNum:{userID}:{goodsID} 用户累计预扣数量
--
-- 流程:
--   1. 只在“当前进行中 secNum 和传入 secNum 一致”时执行回补，避免误回补别的请求。
--   2. 清空用户进行中占位。
--   3. 把 Redis 库存加回去。
--   4. 把用户累计预扣数量减回去，最低回到 0。
--   5. 用失败记录覆盖 secNum 对应的预秒杀记录，供 GetSecKillInfo 查询。
--
-- 幂等说明:
--   - 第一次失败回补后，keyUserGoodsSecNum 会被清空。
--   - 同一个 secNum 再次调用时，第 1 步不会命中，库存和用户计数不会再次回补。
--   - 但 secNum 对应的失败记录仍会被重复写入，保证查询结果稳定可见。
--   - 占位被清空后，用户可以重新发起新的秒杀请求。
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
