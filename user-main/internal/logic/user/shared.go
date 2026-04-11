package userlogic

import (
	"fmt"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/model"
)

func newModelUserFromCreateRequest(req *v1.CreateUserRequest) *model.User {
	return &model.User{
		UserName: req.UserName,
		Pwd:      req.Pwd,
		Sex:      int(req.Sex),
		Age:      int(req.Age),
		Email:    req.Email,
		Contact:  req.Contact,
		Mobile:   req.Mobile,
		IdCard:   req.IdCard,
	}
}

func newUserReplyData(user *model.User, includeUserID bool) *v1.GetUserReplyData {
	if user == nil {
		return nil
	}

	reply := &v1.GetUserReplyData{
		UserName: user.UserName,
		Pwd:      user.Pwd,
		Sex:      int32(user.Sex),
		Age:      int32(user.Age),
		Email:    user.Email,
		Contact:  user.Contact,
		Mobile:   user.Mobile,
		IdCard:   user.IdCard,
	}
	if includeUserID {
		reply.UserID = fmt.Sprintf("%d", user.UserID)
	}
	return reply
}
