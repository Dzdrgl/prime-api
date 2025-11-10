package user

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"githup.com/ademdizdar/prime-api/internal/app"
)

func FetchAndUploadFromURL(ctx context.Context, storage app.ObjectStorage, srcURL, key string, maxBytes int64) (publicURL string, signedURL string, err error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srcURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("fetch source: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("bad status: %s", resp.Status)
	}

	var reader io.Reader = resp.Body
	if maxBytes > 0 {
		reader = io.LimitReader(resp.Body, maxBytes)
	}

	buf, err := io.ReadAll(reader)
	if err != nil {
		return "", "", fmt.Errorf("read body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(buf)
	}

	r, size := app.BytesReader(buf)

	pub, signed, putErr := storage.Put(ctx, key, r, size, contentType)
	if putErr != nil {
		return "", "", fmt.Errorf("put to storage: %w", putErr)
	}
	return pub, signed, nil
}
