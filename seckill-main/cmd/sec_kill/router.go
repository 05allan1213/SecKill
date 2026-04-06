package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/BitofferHub/pkg/constant"
	"github.com/BitofferHub/pkg/middlewares/log"
	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/service"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	nethttp "net/http"
	"strconv"
	"time"
)

// SecKill
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: gin的GetUserInfo入口
//	@param ctx
func SecKill(ctx *gin.Context) {
	var reply = new(pb.SecKillV1Reply)
	defer func() {
		//reply.Message =
		reply.Message = service.GetErrMsg(int(reply.Code))
		ctx.JSON(nethttp.StatusOK, reply)
	}()
	userIDStr := ctx.Request.Header.Get(constant.UserID)
	traceID := ctx.Request.Header.Get(constant.TraceID)
	userID, _ := strconv.Atoi(userIDStr)
	//userName := fmt.Sprintf("%s", userID)
	// 解析请求包
	var req pb.SecKillV1Request
	if err := ctx.ShouldBind(&req); err != nil {
		log.ErrorContextf(ctx, " shouldBind err %s\n", err.Error())
		reply.Code = -100
		return
	}
	req.UserID = int64(userID)
	c := context.WithValue(context.Background(), constant.TraceID, traceID)
	log.InfoContextf(c, "req %+v\n", req)
	var err error
	reply, err = us.SecKillV1(c, &req)
	if err != nil {
		log.InfoContextf(c, "SecKillV1 err %s\n", err.Error())
		return
	}
}

// SecKill
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: gin的GetUserInfo入口
//	@param ctx
func SecKillV2(ctx *gin.Context) {
	var reply = new(pb.SecKillV2Reply)
	defer func() {
		//reply.Message =
		ctx.JSON(nethttp.StatusOK, reply)
	}()
	userIDStr := ctx.Request.Header.Get(constant.UserID)
	traceID := ctx.Request.Header.Get(constant.TraceID)
	userID, _ := strconv.Atoi(userIDStr)
	//userName := fmt.Sprintf("%s", userID)
	// 解析请求包
	var req pb.SecKillV2Request
	if err := ctx.ShouldBind(&req); err != nil {
		log.ErrorContextf(ctx, " shouldBind err %s\n", err.Error())
		reply.Code = -100
		return
	}
	req.UserID = int64(userID)
	c := context.WithValue(context.Background(), constant.TraceID, traceID)
	log.InfoContextf(c, "req %+v\n", req)
	var err error
	reply, err = us.SecKillV2(c, &req)
	if err != nil {
		log.InfoContextf(c, "SecKillV2 err %s\n", err.Error())
		return
	}
}

func SecKillV3(ctx *gin.Context) {
	var reply = new(pb.SecKillV3Reply)
	defer func() {
		//reply.Message =
		ctx.JSON(nethttp.StatusOK, reply)
	}()
	userIDStr := ctx.Request.Header.Get(constant.UserID)
	traceID := ctx.Request.Header.Get(constant.TraceID)
	userID, _ := strconv.Atoi(userIDStr)
	//userName := fmt.Sprintf("%s", userID)
	// 解析请求包
	var req pb.SecKillV3Request
	if err := ctx.ShouldBind(&req); err != nil {
		log.ErrorContextf(ctx, " shouldBind err %s\n", err.Error())
		reply.Code = -100
		return
	}
	req.UserID = int64(userID)
	c := context.WithValue(context.Background(), constant.TraceID, traceID)
	log.InfoContextf(c, "req %+v\n", req)
	var err error
	reply, err = us.SecKillV3(c, &req)
	if err != nil {
		log.InfoContextf(c, "SecKillV3 err %s\n", err.Error())
		return
	}
}

func GetSecKillInfo(ctx *gin.Context) {
	var reply = new(pb.GetSecKillInfoReply)
	defer func() {
		//reply.Message =
		ctx.JSON(nethttp.StatusOK, reply)
	}()
	userIDStr := ctx.Request.Header.Get(constant.UserID)
	traceID := ctx.Request.Header.Get(constant.TraceID)
	fmt.Println("userID is ", userIDStr)
	userID, _ := strconv.Atoi(userIDStr)
	//userName := fmt.Sprintf("%s", userID)
	// 解析请求包
	secNum := ctx.Query("sec_num")
	if secNum == "" {
		log.ErrorContextf(ctx, " secNum not exist")
		reply.Code = -100
		return

	}
	var req pb.GetSecKillInfoRequest
	req.SecNum = secNum
	req.UserID = int64(userID)
	c := context.WithValue(context.Background(), constant.TraceID, traceID)
	log.InfoContextf(c, "req %+v\n", req)
	var err error
	reply, err = us.GetSecKillInfo(c, &req)
	if err != nil {
		log.InfoContextf(c, "GetSecKillInfo err %s\n", err.Error())
		return
	}
}

func GetGoodsInfo(ctx *gin.Context) {
	var reply = new(pb.GetGoodsInfoReply)
	defer func() {
		//reply.Message =
		ctx.JSON(nethttp.StatusOK, reply)
	}()
	userIDStr := ctx.Request.Header.Get(constant.UserID)
	traceID := ctx.Request.Header.Get(constant.TraceID)
	fmt.Println("userID is ", userIDStr)
	userID, _ := strconv.Atoi(userIDStr)
	//userName := fmt.Sprintf("%s", userID)
	// 解析请求包
	goodsNum := ctx.Query("goods_num")
	if goodsNum == "" {
		log.ErrorContextf(ctx, " secNum not exist")
		reply.Code = -100
		return

	}
	var req pb.GetGoodsInfoRequest
	req.GoodsNum = goodsNum
	req.UserID = int64(userID)
	c := context.WithValue(context.Background(), constant.TraceID, traceID)
	log.InfoContextf(c, "req %+v\n", req)
	var err error
	reply, err = us.GetGoodsInfo(c, &req)
	if err != nil {
		log.InfoContextf(c, "GetGoodsInfo err %s\n", err.Error())
		return
	}
}

func GetGoodsList(ctx *gin.Context) {
	var reply = new(pb.GetGoodsListReply)
	defer func() {
		//reply.Message =
		ctx.JSON(nethttp.StatusOK, reply)
	}()
	userIDStr := ctx.Request.Header.Get(constant.UserID)
	traceID := ctx.Request.Header.Get(constant.TraceID)
	fmt.Println("userID is ", userIDStr)
	userID, _ := strconv.Atoi(userIDStr)
	limitStr := ctx.Query("limit")
	offsetStr := ctx.Query("offset")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit == 0 {
		limit = 200
	}
	var req pb.GetGoodsListRequest
	req.UserID = int64(userID)
	req.Limit = int32(limit)
	req.Offset = int32(offset)
	c := context.WithValue(context.Background(), constant.TraceID, traceID)
	log.InfoContextf(c, "req %+v\n", req)
	var err error
	reply, err = us.GetGoodsList(c, &req)
	if err != nil {
		log.InfoContextf(c, "GetGoodsList err %s\n", err.Error())
		return
	}
}

// InfoLog
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: gin middleware for log request and reply
//	@return gin.HandlerFunc
func InfoLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		beginTime := time.Now()
		// ***** 1. get request body ****** //
		traceID := c.Request.Header.Get(constant.TraceID)
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close() //  must close
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		// ***** 2. set requestID for goroutine ctx ****** //
		//duration := float64(time.Since(beginTime)) / float64(time.Second)
		ctx := context.WithValue(context.Background(), constant.TraceID, traceID)
		c.Next()
		log.InfoContextf(ctx, "ReqPath[%s]-Cost[%v]\n", c.Request.URL.Path, time.Since(beginTime))
	}
}
