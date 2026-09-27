package clients

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/lania-smp/backend/internal/config"
)

// ObjectStorage writes image objects to the project's S3 bucket. Pages read them through imgproxy,
// which has its own credentials; GetObject is only for server-side exports like the glyth pack.
type ObjectStorage interface {
	Bucket() string
	PutObject(ctx context.Context, key string, body []byte, contentType string) error
	GetObject(ctx context.Context, key string) ([]byte, error)
}

type objectStorage struct {
	client *s3.Client
	bucket string
}

// NewObjectStorage loads the AWS config once at startup. LoadDefaultConfig makes no network call
// and tolerates missing credentials, so local development still boots and only a real upload fails.
func NewObjectStorage(ctx context.Context) (ObjectStorage, error) {
	opts := []func(*awsConfig.LoadOptions) error{}
	if region := config.GetS3Region(); region != "" {
		opts = append(opts, awsConfig.WithRegion(region))
	}
	cfg, err := awsConfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint := config.GetS3Endpoint(); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = config.IsS3PathStyleForced()
	})
	return &objectStorage{client: client, bucket: config.GetS3Bucket()}, nil
}

func (s *objectStorage) Bucket() string {
	return s.bucket
}

func (s *objectStorage) PutObject(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	return err
}

func (s *objectStorage) GetObject(ctx context.Context, key string) ([]byte, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}
