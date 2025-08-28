package s3

import (
	"bytes"
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	s3Client *s3.Client
	bucket   string
)

func init() {
	// Load AWS SDK config (from env, shared credentials, or role)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load AWS SDK config: " + err.Error())
	}

	// Initialize S3 client globally
	s3Client = s3.NewFromConfig(cfg)

	// Read bucket name from environment variable
	bucket = os.Getenv("S3_BUCKET")
	if bucket == "" {
		panic("S3_BUCKET environment variable is not set")
	}
}

// Upload uploads a file to the configured bucket
func UploadBytes(ctx context.Context, key string, content []byte) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               &bucket,
		Key:                  &key,
		Body:                 bytes.NewReader(content), // convert byte slice to io.Reader
		ServerSideEncryption: types.ServerSideEncryptionAes256,
	})
	return err
}
