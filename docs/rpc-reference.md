# RPC 参考

## user-main

### CreateUser

请求字段：

- `userName`
- `pwd`
- `sex`
- `age`
- `email`
- `contact`
- `mobile`
- `idCard`

示例：
```json
{
  "userName": "alice",
  "pwd": "123456",
  "sex": 1,
  "age": 18,
  "email": "alice@example.com",
  "contact": "Shanghai",
  "mobile": "13800000000",
  "idCard": "510681199901010001"
}
```

### GetUser

请求字段：

- `userID`

### GetUserByName

请求字段：

- `userName`

返回字段：

- `code`
- `message`
- `data.userID`
- `data.userName`
- `data.pwd`
- `data.sex`
- `data.age`
- `data.email`
- `data.contact`
- `data.mobile`
- `data.idCard`

## seckill-main

### SecKillV1 / SecKillV2 / SecKillV3

请求字段：

- `userID`
- `goodsNum`
- `num`

示例：
```json
{
  "userID": 1,
  "goodsNum": "abc123",
  "num": 1
}
```

### GetGoodsList

请求字段：

- `userID`
- `offset`
- `limit`

### GetSecKillInfo

请求字段：

- `userID`
- `secNum`

返回字段：

- `code`
- `message`
- `data.status`
- `data.orderNum`
- `data.secNum`
- `data.goodsNum`
