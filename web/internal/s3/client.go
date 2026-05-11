package s3

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Client struct {
	client *s3.Client
	bucket string
}

func NewClient(ctx context.Context, endpoint, region, accessKey, secretKey, bucket string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("config.LoadDefaultConfig: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationUnset
	})

	c := &Client{client: client, bucket: bucket}

	if err := c.ensureBucket(ctx); err != nil {
		return nil, fmt.Errorf("ensureBucket: %w", err)
	}

	return c, nil
}

func (c *Client) ensureBucket(ctx context.Context) error {
	_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	})
	if err == nil {
		return nil
	}

	_, err = c.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(c.bucket),
	})
	if err != nil {
		var alreadyExists *types.BucketAlreadyExists
		if errors.As(err, &alreadyExists) {
			return nil
		}
		return fmt.Errorf("CreateBucket: %w", err)
	}

	slog.Info("created bucket", "bucket", c.bucket)
	return nil
}

func (c *Client) Upload(ctx context.Context, filename string, content io.Reader) (string, error) {
	ext := filepath.Ext(filename)
	key := generateKey(ext)

	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   content,
	})
	if err != nil {
		return "", fmt.Errorf("PutObject: %w", err)
	}

	return key, nil
}

func (c *Client) Download(ctx context.Context, key string) (*s3.GetObjectOutput, error) {
	out, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("GetObject: %w", err)
	}
	return out, nil
}

func generateKey(ext string) string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b) + ext
}