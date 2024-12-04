package service

import (
	"context"
	"github.com/minio/minio-go/v7"
	"net/url"
	"time"
)

type Service interface {
	GetPresignedUrl(ctx context.Context, bucketName string, objectName string) (*url.URL, error)
}

type service struct {
	minioClient *minio.Client
	expires     time.Duration
}

func NewService(minioClient *minio.Client) Service {
	return &service{
		minioClient: minioClient,
		expires:     time.Second * 10,
	}
}

func (s *service) GetPresignedUrl(ctx context.Context, bucketName string, objectName string) (*url.URL, error) {
	return s.minioClient.PresignedPutObject(ctx, bucketName, objectName, s.expires)
}