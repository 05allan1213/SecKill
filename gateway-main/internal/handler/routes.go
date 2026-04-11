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

	protectedRoutes := []struct {
		routeKey string
		route    rest.Route
	}{
		{
			routeKey: "/get_user_info",
			route: rest.Route{
				Method:  http.MethodGet,
				Path:    "/get_user_info",
				Handler: GetUserInfoHandler(serverCtx),
			},
		},
		{
			routeKey: "/bitstorm/get_user_info",
			route: rest.Route{
				Method:  http.MethodGet,
				Path:    "/bitstorm/get_user_info",
				Handler: BitstormGetUserInfoHandler(serverCtx),
			},
		},
		{
			routeKey: "/bitstorm/get_user_info_by_name",
			route: rest.Route{
				Method:  http.MethodGet,
				Path:    "/bitstorm/get_user_info_by_name",
				Handler: BitstormGetUserInfoByNameHandler(serverCtx),
			},
		},
		{
			routeKey: "/bitstorm/v1/sec_kill",
			route: rest.Route{
				Method:  http.MethodPost,
				Path:    "/bitstorm/v1/sec_kill",
				Handler: BitstormSecKillV1Handler(serverCtx),
			},
		},
		{
			routeKey: "/bitstorm/v2/sec_kill",
			route: rest.Route{
				Method:  http.MethodPost,
				Path:    "/bitstorm/v2/sec_kill",
				Handler: BitstormSecKillV2Handler(serverCtx),
			},
		},
		{
			routeKey: "/bitstorm/v3/sec_kill",
			route: rest.Route{
				Method:  http.MethodPost,
				Path:    "/bitstorm/v3/sec_kill",
				Handler: BitstormSecKillV3Handler(serverCtx),
			},
		},
		{
			routeKey: "/bitstorm/v3/get_sec_kill_info",
			route: rest.Route{
				Method:  http.MethodGet,
				Path:    "/bitstorm/v3/get_sec_kill_info",
				Handler: BitstormGetSecKillInfoHandler(serverCtx),
			},
		},
	}

	for _, item := range protectedRoutes {
		addProtected(item.routeKey, item.route)
	}
}
