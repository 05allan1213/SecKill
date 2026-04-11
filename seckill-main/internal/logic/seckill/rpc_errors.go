package seckill

import (
	"errors"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func goodsLookupError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return grpcstatus.Error(codes.InvalidArgument, "goods not found")
	}
	return grpcstatus.Error(codes.Unavailable, "goods storage unavailable")
}

func dependencyUnavailableError(message string) error {
	return grpcstatus.Error(codes.Unavailable, message)
}
