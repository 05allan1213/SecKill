package handler

import (
	"net/http"

	gwmiddleware "github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{{
		Method:  http.MethodPost,
		Path:    "/login",
		Handler: LoginHandler(serverCtx),
	}})

	auth := gwmiddleware.NewAuthMiddleware(serverCtx.Config.Auth)
	addProtected := func(routeKey string, route rest.Route) {
		server.AddRoutes(rest.WithMiddlewares([]rest.Middleware{
			auth.Handle,
			gwmiddleware.NewRouteLimitMiddleware(serverCtx, routeKey).Handle,
		}, route))
	}

	addProtected("/get_user_info", rest.Route{
		Method:  http.MethodGet,
		Path:    "/get_user_info",
		Handler: GetUserInfoHandler(serverCtx),
	})
	addProtected("/bitstorm/get_user_info", rest.Route{
		Method:  http.MethodGet,
		Path:    "/bitstorm/get_user_info",
		Handler: BitstormGetUserInfoHandler(serverCtx),
	})
	addProtected("/bitstorm/get_user_info_by_name", rest.Route{
		Method:  http.MethodGet,
		Path:    "/bitstorm/get_user_info_by_name",
		Handler: BitstormGetUserInfoByNameHandler(serverCtx),
	})
	addProtected("/bitstorm/v1/sec_kill", rest.Route{
		Method:  http.MethodPost,
		Path:    "/bitstorm/v1/sec_kill",
		Handler: BitstormSecKillV1Handler(serverCtx),
	})
	addProtected("/bitstorm/v2/sec_kill", rest.Route{
		Method:  http.MethodPost,
		Path:    "/bitstorm/v2/sec_kill",
		Handler: BitstormSecKillV2Handler(serverCtx),
	})
	addProtected("/bitstorm/v3/sec_kill", rest.Route{
		Method:  http.MethodPost,
		Path:    "/bitstorm/v3/sec_kill",
		Handler: BitstormSecKillV3Handler(serverCtx),
	})
	addProtected("/bitstorm/v3/get_sec_kill_info", rest.Route{
		Method:  http.MethodGet,
		Path:    "/bitstorm/v3/get_sec_kill_info",
		Handler: BitstormGetSecKillInfoHandler(serverCtx),
	})
}
