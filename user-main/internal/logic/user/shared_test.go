package userlogic

import (
	"testing"

	"github.com/BitofferHub/user/internal/model"
)

func TestNewUserReplyData(t *testing.T) {
	user := &model.User{
		UserID:   1,
		UserName: "alice",
		Pwd:      "pwd",
		Sex:      1,
		Age:      18,
		Email:    "alice@example.com",
		Contact:  "Alice",
		Mobile:   "13800000000",
		IdCard:   "1234567890",
	}

	reply := newUserReplyData(user, false)
	if reply.UserID != "" {
		t.Fatalf("expected empty user id when includeUserID is false, got %q", reply.UserID)
	}
	if reply.UserName != user.UserName || reply.Pwd != user.Pwd || reply.Sex != int32(user.Sex) ||
		reply.Age != int32(user.Age) || reply.Email != user.Email || reply.Contact != user.Contact ||
		reply.Mobile != user.Mobile || reply.IdCard != user.IdCard {
		t.Fatalf("unexpected reply data mapping: %+v", reply)
	}

	replyWithID := newUserReplyData(user, true)
	if replyWithID.UserID != "1" {
		t.Fatalf("expected user id %q, got %q", "1", replyWithID.UserID)
	}
}
