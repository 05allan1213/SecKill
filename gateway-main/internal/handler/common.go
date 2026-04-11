package handler

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func execute[Resp any](w http.ResponseWriter, r *http.Request, call func() (*Resp, error)) {
	resp, err := call()
	if err != nil {
		writeError(r, w, err)
		return
	}
	middleware.RecordAccessResponse(r.Context(), summarizeGatewayResponse(resp))
	middleware.RecordAccessCode(r.Context(), extractGatewayResponseCode(resp))
	httpx.OkJsonCtx(r.Context(), w, resp)
}

func parseAndExecute[Req any, Resp any](w http.ResponseWriter, r *http.Request, call func(*Req) (*Resp, error)) {
	var req Req
	if err := httpx.Parse(r, &req); err != nil {
		middleware.RecordAccessError(r.Context(), err)
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	middleware.RecordAccessRequest(r.Context(), summarizeGatewayRequest(&req))
	execute(w, r, func() (*Resp, error) {
		return call(&req)
	})
}

func summarizeGatewayRequest(req any) any {
	switch v := req.(type) {
	case *types.LoginRequest:
		if v == nil {
			return nil
		}
		return map[string]any{"username": v.Username}
	case *types.SecKillRequest:
		if v == nil {
			return nil
		}
		return map[string]any{"goodsNum": v.GoodsNum, "num": v.Num}
	case *types.GetSecKillInfoRequest:
		if v == nil {
			return nil
		}
		return map[string]any{"secNum": v.SecNum}
	case *types.GetUserByNameRequest:
		if v == nil {
			return nil
		}
		userName := v.UserName
		if userName == "" {
			userName = v.UserNameAlt
		}
		return map[string]any{"userName": userName}
	default:
		return nil
	}
}

func summarizeGatewayResponse(resp any) any {
	switch v := resp.(type) {
	case *types.LoginResponse:
		if v == nil {
			return nil
		}
		return map[string]any{"code": v.Code, "expire": v.Expire}
	case *types.GetUserReply:
		if v == nil {
			return nil
		}
		return map[string]any{"code": v.Code, "message": v.Message}
	case *types.GetSecKillInfoReply:
		if v == nil {
			return nil
		}
		return map[string]any{
			"code":     v.Code,
			"message":  v.Message,
			"status":   v.Data.Status,
			"orderNum": v.Data.OrderNum,
			"secNum":   v.Data.SecNum,
		}
	case *types.SecKillV1Reply:
		if v == nil {
			return nil
		}
		return map[string]any{"code": v.Code, "message": v.Message, "orderNum": v.Data.OrderNum}
	case *types.SecKillV2Reply:
		if v == nil {
			return nil
		}
		return map[string]any{"code": v.Code, "message": v.Message, "orderNum": v.Data.OrderNum}
	case *types.SecKillV3Reply:
		if v == nil {
			return nil
		}
		return map[string]any{"code": v.Code, "message": v.Message, "secNum": v.Data.SecNum}
	case *types.WelcomeResponse:
		if v == nil {
			return nil
		}
		return map[string]any{"welcome": v.Welcome}
	default:
		return nil
	}
}

func extractGatewayResponseCode(resp any) int {
	switch v := resp.(type) {
	case *types.LoginResponse:
		if v != nil {
			return int(v.Code)
		}
	case *types.GetUserReply:
		if v != nil {
			return int(v.Code)
		}
	case *types.GetSecKillInfoReply:
		if v != nil {
			return int(v.Code)
		}
	case *types.SecKillV1Reply:
		if v != nil {
			return int(v.Code)
		}
	case *types.SecKillV2Reply:
		if v != nil {
			return int(v.Code)
		}
	case *types.SecKillV3Reply:
		if v != nil {
			return int(v.Code)
		}
	}
	return http.StatusOK
}
