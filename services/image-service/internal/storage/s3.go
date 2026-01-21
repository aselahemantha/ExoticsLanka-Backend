package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client     *s3.Client
	bucketName string
}

func NewS3Client(region, bucketName, endpoint string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	// For custom endpoints (e.g., LocalStack, MinIO)
	if endpoint != "" {
		cfg.BaseEndpoint = aws.String(endpoint)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Client{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *S3Client) UploadFile(key string, file io.Reader, contentType string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		// ACL:         types.ObjectCannedACLPublicRead, // Depending on bucket policy
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Construct URL (This depends on if it's AWS or LocalStack/MinIO)
	// For now, returning Key or a constructed URL
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, key)
	if s.client.Options().BaseEndpoint != nil {
		url = fmt.Sprintf("%s/%s/%s", *s.client.Options().BaseEndpoint, s.bucketName, key)
	}

	return url, nil
}

func (s *S3Client) DeleteFile(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}
