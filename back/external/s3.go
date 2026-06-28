package external

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// realS3Client は AWS SDK for Go v2 を使って学習CSVをS3へPUTする S3Client 実装。
type realS3Client struct {
	client *s3.Client
	bucket string
}

// NewS3Client は S3_BUCKET_NAME が設定されていれば実クライアントを、
// 未設定ならスタブを返す。
// 認証情報（AWS_ACCESS_KEY_ID/SECRET、またはLambda実行ロール）と AWS_REGION は
// AWS SDK のデフォルト設定チェーンから解決する。
func NewS3Client() S3Client {
	bucket := os.Getenv("S3_BUCKET_NAME")
	if !isConfigured(bucket) {
		return &stubS3Client{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("[external] S3 設定の読み込みに失敗したためスタブにフォールバック: %v", err)
		return &stubS3Client{}
	}

	log.Printf("[external] S3 client 有効化（bucket=%s, region=%s）", bucket, cfg.Region)
	return &realS3Client{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}
}

// Upload はオブジェクトをS3へPUTし、s3://bucket/key 形式のパスを返す。
func (c *realS3Client) Upload(key string, body []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String("text/csv"),
	})
	if err != nil {
		return "", fmt.Errorf("S3 Upload: put object failed (bucket=%s key=%s): %w", c.bucket, key, err)
	}
	return fmt.Sprintf("s3://%s/%s", c.bucket, key), nil
}
