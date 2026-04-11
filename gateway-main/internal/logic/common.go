package logic

import (
	"context"
	"strconv"

	"github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
	"github.com/BitofferHub/pkg/constant"
	secproto "github.com/BitofferHub/seckill/api/sec_kill/proto"
	userv1 "github.com/BitofferHub/user/api/user/v1"
)

func rpcContext(ctx context.Context) context.Context {
	traceID := middleware.TraceIDFromContext(ctx)
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, constant.TraceID, traceID)
}

func currentUserID(ctx context.Context) (int64, error) {
	userID := middleware.UserIDFromContext(ctx)
	if userID == "" {
		return 0, &middleware.HTTPError{Status: 401, Message: "no authentication"}
	}
	parsed, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 0, &middleware.HTTPError{Status: 401, Message: "no authentication"}
	}
	return parsed, nil
}

func fetchCurrentUser(ctx context.Context, svcCtx *svc.ServiceContext) (*userv1.GetUserReply, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	return svcCtx.UserClient.GetUser(rpcContext(ctx), &userv1.GetUserRequest{UserID: userID})
}

func runSecKill[PBReq any, PBResp any, Reply any](
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	req *types.SecKillRequest,
	build func(int64, *types.SecKillRequest) *PBReq,
	invoke func(context.Context, *PBReq) (*PBResp, error),
	mapper func(*PBResp) *Reply,
) (*Reply, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := invoke(rpcContext(ctx), build(userID, req))
	if err != nil {
		return nil, err
	}
	return mapper(reply), nil
}

func mapUserReply(resp *userv1.GetUserReply) *types.GetUserReply {
	if resp == nil {
		return &types.GetUserReply{}
	}

	reply := &types.GetUserReply{
		Code:    int64(resp.Code),
		Message: resp.Message,
	}
	if resp.Data != nil {
		reply.Data = types.UserData{
			UserID:   resp.Data.UserID,
			UserName: resp.Data.UserName,
			Pwd:      resp.Data.Pwd,
			Sex:      int64(resp.Data.Sex),
			Age:      int64(resp.Data.Age),
			Email:    resp.Data.Email,
			Contact:  resp.Data.Contact,
			Mobile:   resp.Data.Mobile,
			IdCard:   resp.Data.IdCard,
		}
	}

	return reply
}

func mapUserByNameReply(resp *userv1.GetUserByNameReply) *types.GetUserReply {
	if resp == nil {
		return &types.GetUserReply{}
	}
	return mapUserReply(&userv1.GetUserReply{
		Code:    resp.Code,
		Message: resp.Message,
		Data:    resp.Data,
	})
}

func mapSecKillV1Reply(resp *secproto.SecKillV1Reply) *types.SecKillV1Reply {
	if resp == nil {
		return &types.SecKillV1Reply{}
	}
	reply := &types.SecKillV1Reply{
		Code:    int64(resp.Code),
		Message: resp.Message,
	}
	if resp.Data != nil {
		reply.Data = types.SecKillV1ReplyData{OrderNum: resp.Data.OrderNum}
	}
	return reply
}

func mapSecKillV2Reply(resp *secproto.SecKillV2Reply) *types.SecKillV2Reply {
	if resp == nil {
		return &types.SecKillV2Reply{}
	}
	reply := &types.SecKillV2Reply{
		Code:    int64(resp.Code),
		Message: resp.Message,
	}
	if resp.Data != nil {
		reply.Data = types.SecKillV2ReplyData{OrderNum: resp.Data.OrderNum}
	}
	return reply
}

func mapSecKillV3Reply(resp *secproto.SecKillV3Reply) *types.SecKillV3Reply {
	if resp == nil {
		return &types.SecKillV3Reply{}
	}
	reply := &types.SecKillV3Reply{
		Code:    int64(resp.Code),
		Message: resp.Message,
	}
	if resp.Data != nil {
		reply.Data = types.SecKillV3ReplyData{SecNum: resp.Data.SecNum}
	}
	return reply
}

func mapGetSecKillInfoReply(resp *secproto.GetSecKillInfoReply) *types.GetSecKillInfoReply {
	if resp == nil {
		return &types.GetSecKillInfoReply{}
	}
	reply := &types.GetSecKillInfoReply{
		Code:    int64(resp.Code),
		Message: resp.Message,
	}
	if resp.Data != nil {
		reply.Data = types.GetSecKillInfoReplyData{
			Status:   int64(resp.Data.Status),
			OrderNum: resp.Data.OrderNum,
			SecNum:   resp.Data.SecNum,
			GoodsNum: resp.Data.GoodsNum,
			Reason:   resp.Data.Reason,
		}
	}
	return reply
}
