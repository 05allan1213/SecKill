package main

import (
	"net/http"

	"github.com/BitofferHub/seckill/internal/config"
	"github.com/BitofferHub/seckill/internal/service"
	"github.com/BitofferHub/seckill/internal/svc"
	"github.com/gin-gonic/gin"
)

var us *service.SecKillService

func StartCompatHTTPServer(c config.Config, svcCtx *svc.ServiceContext) (*http.Server, error) {
	us = svcCtx.SecKillService

	router := gin.Default()
	router.Use(InfoLog())

	v1 := router.Group("v1")
	{
		v1.POST("/sec_kill", SecKill)
	}
	v2 := router.Group("v2")
	{
		v2.POST("/sec_kill", SecKillV2)
	}
	v3 := router.Group("v3")
	{
		v3.POST("/sec_kill", SecKillV3)
		v3.GET("/get_sec_kill_info", GetSecKillInfo)
	}
	router.GET("/get_goods_info", GetGoodsInfo)
	router.GET("/get_goods_list", GetGoodsList)

	server := &http.Server{
		Addr:    c.CompatHttp.Addr,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return server, nil
}
