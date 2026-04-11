package svc

import (
	"github.com/BitofferHub/seckill/internal/config"
	"github.com/BitofferHub/seckill/internal/data"
	projectlog "github.com/BitofferHub/seckill/internal/log"
)

type ServiceContext struct {
	Config config.Config
	Data   *data.Data
	*data.Repositories
}

func NewServiceContext(c config.Config) *ServiceContext {
	projectlog.Init("./log/")

	dataLayer, err := data.NewDataFromConfig(c.Data, c.Log)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:       c,
		Data:         dataLayer,
		Repositories: data.NewRepositories(dataLayer),
	}
}
