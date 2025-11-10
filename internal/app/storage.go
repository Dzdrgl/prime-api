package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStorage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (publicURL string, signedURL string, err error)
	PresignedPut(ctx context.Context, key, contentType string, ttl time.Duration) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
	PublicURL(key string) *string
	Delete(ctx context.Context, key string) error
}

type MinioStorage struct {
	client     *minio.Client
	bucket     string
	publicBase string
}

func InitMinio(cfg *Config) (*MinioStorage, error) {
	u, err := url.Parse(cfg.S3Endpoint)
	if err != nil {
		return nil, err
	}

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: u.Scheme == "https",
		Region: cfg.S3Region,
	}
	if cfg.S3UsePathStyle {
		opts.BucketLookup = minio.BucketLookupPath
	}

	cl, err := minio.New(u.Host, opts)
	if err != nil {
		return nil, err
	}

	// bucket yoksa oluşturmayı deneyebiliriz (idempotent)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	exists, err := cl.BucketExists(ctx, cfg.S3Bucket)
	if err == nil && !exists {
		_ = cl.MakeBucket(ctx, cfg.S3Bucket, minio.MakeBucketOptions{Region: cfg.S3Region})
	}

	return &MinioStorage{
		client:     cl,
		bucket:     cfg.S3Bucket,
		publicBase: cfg.S3PublicBaseURL, // boşsa sadece signed dönecek
	}, nil
}

func (m *MinioStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", err
	}

	var publicURL string
	if m.publicBase != "" {
		publicURL = fmt.Sprintf("%s/%s", m.publicBase, path.Clean(key))
	}

	// imzalı (fallback veya kısa ömürlü erişim için)
	reqParams := make(url.Values)
	signed, err := m.client.PresignedGetObject(ctx, m.bucket, key, 24*time.Hour, reqParams)
	if err != nil {
		return publicURL, "", err
	}

	return publicURL, signed.String(), nil
}

func (m *MinioStorage) Delete(ctx context.Context, key string) error {
	return m.client.RemoveObject(ctx, m.bucket, path.Clean(key), minio.RemoveObjectOptions{})
}

func (m *MinioStorage) PresignedPut(ctx context.Context, key, contentType string, ttl time.Duration) (string, error) {
	u, err := m.client.PresignedPutObject(ctx, m.bucket, key, ttl)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (m *MinioStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *MinioStorage) PublicURL(key string) *string {
	if m.publicBase == "" {
		return nil
	}
	u := fmt.Sprintf("%s/%s", m.publicBase, path.Clean(key))
	return &u
}

// Küçük yardımcı (isteğe bağlı)
func BytesReader(b []byte) (io.Reader, int64) { return bytes.NewReader(b), int64(len(b)) }
