package telegram

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// downloadFile downloads a file from URL to a temporary file
func (m *ClientManager) downloadFile(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return m.downloadFileWithTimeout(ctx, url)
}

// downloadFileWithTimeout downloads a file with context timeout
func (m *ClientManager) downloadFileWithTimeout(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ext := filepath.Ext(url)
	if ext == "" {
		ext = ".tmp"
	}

	tmp, err := os.CreateTemp("", "tg-media-*"+ext)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(tmp, resp.Body)
	tmp.Close()

	if err != nil {
		os.Remove(tmp.Name())
		return "", err
	}

	return tmp.Name(), nil
}
