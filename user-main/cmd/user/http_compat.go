package main

import (
	"net/http"

	"github.com/BitofferHub/user/internal/config"
	"github.com/BitofferHub/user/internal/service"
	"github.com/BitofferHub/user/internal/svc"
	"github.com/gin-gonic/gin"
)

var us *service.UserService

func StartCompatHTTPServer(c config.Config, svcCtx *svc.ServiceContext) (*http.Server, error) {
	us = svcCtx.UserService

	router := gin.Default()
	router.Use(InfoLog())
	router.GET("/get_user_info", GetUserInfo)
	router.GET("/get_user_info_by_name", GetUserInfoByName)

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
