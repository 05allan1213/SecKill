package userlogic

import (
	"errors"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func userRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return grpcstatus.Error(codes.NotFound, "user not found")
	}
	return grpcstatus.Error(codes.Unavailable, "user storage unavailable")
}
