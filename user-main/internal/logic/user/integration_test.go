//go:build integration

package userlogic

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/config"
	"github.com/BitofferHub/user/internal/svc"
)

func TestUserLogicIntegration(t *testing.T) {
	cfg, err := config.Load("../../../etc/user.yaml")
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	svcCtx := svc.NewServiceContext(cfg)
	t.Cleanup(svcCtx.Close)

	suffix := time.Now().UnixNano()
	userName := fmt.Sprintf("itest_user_%d", suffix)

	createReply, err := NewCreateUserLogic(context.Background(), svcCtx).CreateUser(&v1.CreateUserRequest{
		UserName: userName,
		Pwd:      "123456",
		Sex:      1,
		Age:      18,
		Email:    fmt.Sprintf("%s@example.com", userName),
		Contact:  "shanghai",
		Mobile:   fmt.Sprintf("188%08d", suffix%100000000),
		IdCard:   fmt.Sprintf("510681%012d", suffix%1000000000000),
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	if createReply.Message != "trytest" {
		t.Fatalf("unexpected create reply: %#v", createReply)
	}

	byNameReply, err := NewGetUserByNameLogic(context.Background(), svcCtx).GetUserByName(&v1.GetUserByNameRequest{
		UserName: userName,
	})
	if err != nil {
		t.Fatalf("get user by name failed: %v", err)
	}
	if byNameReply.Data == nil || byNameReply.Data.UserName != userName {
		t.Fatalf("unexpected get by name reply: %#v", byNameReply)
	}

	userID, err := strconv.ParseInt(byNameReply.Data.UserID, 10, 64)
	if err != nil {
		t.Fatalf("parse user id failed: %v", err)
	}

	getReply, err := NewGetUserLogic(context.Background(), svcCtx).GetUser(&v1.GetUserRequest{
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	if getReply.Data == nil || getReply.Data.UserName != userName {
		t.Fatalf("unexpected get user reply: %#v", getReply)
	}
}
