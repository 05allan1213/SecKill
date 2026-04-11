# user-main

`user-main` 在这个项目里不是重点模块，它主要提供两类能力：

- 登录时按用户名查询用户
- 网关鉴权后查询基础用户信息

## 关键文件

- [internal/logic/user/getuserlogic.go](/home/monody/project/SecKill/user-main/internal/logic/user/getuserlogic.go)
- [internal/logic/user/getuserbynamelogic.go](/home/monody/project/SecKill/user-main/internal/logic/user/getuserbynamelogic.go)
- [internal/data/user.go](/home/monody/project/SecKill/user-main/internal/data/user.go)

## 启动

```bash
cd user-main
go run ./cmd/user -f etc/user.yaml
```

默认测试账号：

- 用户名：`admin`
- 密码：`123321`

完整项目入口请看根目录 [README.md](/home/monody/project/SecKill/README.md)。
