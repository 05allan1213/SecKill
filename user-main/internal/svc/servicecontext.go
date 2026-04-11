package svc

import (
	"github.com/BitofferHub/user/internal/config"
	"github.com/BitofferHub/user/internal/data"
	zaplog "github.com/BitofferHub/user/internal/log"
)

type ServiceContext struct {
	Config   config.Config
	Data     *data.Data
	UserRepo *data.UserRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	zaplog.Init("./log/")

	dataLayer, err := data.NewDataFromConfig(c.Data)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:   c,
		Data:     dataLayer,
		UserRepo: data.NewUserRepo(dataLayer),
	}
}
