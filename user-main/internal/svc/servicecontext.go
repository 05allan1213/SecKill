package svc

import (
	zaplog "github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/user/internal/biz"
	"github.com/BitofferHub/user/internal/config"
	"github.com/BitofferHub/user/internal/data"
	"github.com/BitofferHub/user/internal/service"
)

type ServiceContext struct {
	Config      config.Config
	Data        *data.Data
	UserService *service.UserService
}

func NewServiceContext(c config.Config) *ServiceContext {
	zaplog.Init(
		zaplog.WithLogPath("./log/"),
		zaplog.WithLogLevel("info"),
		zaplog.WithFileName("bitstorm.log"),
		zaplog.WithMaxBackups(100),
		zaplog.WithMaxSize(1024*1024*10),
	)

	dataLayer, err := data.NewDataFromConfig(c.Data)
	if err != nil {
		panic(err)
	}
	userRepo := data.NewUserRepo(dataLayer)
	userUsecase := biz.NewUserUsecase(userRepo)
	userService := service.NewUserService(userUsecase)

	return &ServiceContext{
		Config:      c,
		Data:        dataLayer,
		UserService: userService,
	}
}
