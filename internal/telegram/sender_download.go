package telegram

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"telegram-api/pkg/logger"
)

// downloadFile downloads a file from URL to a temporary file.
func (m *ClientManager) downloadFile(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return m.downloadFileWithTimeout(ctx, url)
}

// downloadFileWithTimeout downloads a file with context timeout.
func (m *ClientManager) downloadFileWithTimeout(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error().Err(err).Msg("Error closing response body")
		}
	}()

	ext := filepath.Ext(url)
	if ext == "" {
		ext = ".tmp"
	}

	tmp, err := os.CreateTemp("", "tg-media-*"+ext)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(tmp, resp.Body)
	if closeErr := tmp.Close(); closeErr != nil {
		logger.Error().Err(closeErr).Msg("Error closing temp file")
	}

	if err != nil {
		if removeErr := os.Remove(tmp.Name()); removeErr != nil {
			logger.Error().Err(removeErr).Msg("Error removing temp file")
		}
		return "", err
	}

	return tmp.Name(), nil
}
