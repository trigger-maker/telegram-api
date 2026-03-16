package telegram

import (
	"context"
	"os"
	"path/filepath"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

// cleanupTempFile removes temporary file and logs any errors.
func cleanupTempFile(filePath string) {
	if err := os.Remove(filePath); err != nil {
		logger.Error().Err(err).Str("path", filePath).Msg("Error removing temp file")
	}
}

// uploadMediaFile downloads and uploads a media file, returns upload and cleanup function.
func (m *ClientManager) uploadMediaFile(
	ctx context.Context,
	api *tg.Client,
	mediaURL string,
) (tg.InputFileClass, string, error) {
	filePath, err := m.downloadFile(mediaURL)
	if err != nil {
		return nil, "", err
	}

	up := uploader.NewUploader(api)
	upload, err := up.FromPath(ctx, filePath)
	if err != nil {
		cleanupTempFile(filePath)
		return nil, "", err
	}

	return upload, filePath, nil
}

// getMessageText returns the caption if available, otherwise returns the text.
func getMessageText(caption, text string) string {
	if caption != "" {
		return caption
	}
	return text
}

// sendPhoto sends a photo message.
func (m *ClientManager) sendPhoto(
	ctx context.Context,
	api *tg.Client,
	builder *message.RequestBuilder,
	req *domain.SendMessageRequest,
) error {
	upload, filePath, err := m.uploadMediaFile(ctx, api, req.MediaURL)
	if err != nil {
		return err
	}
	defer cleanupTempFile(filePath)

	text := getMessageText(req.Caption, req.Text)
	photo := message.UploadedPhoto(upload, styling.Plain(text))
	_, err = builder.Media(ctx, photo)
	return err
}

// sendVideo sends a video message.
func (m *ClientManager) sendVideo(
	ctx context.Context,
	api *tg.Client,
	builder *message.RequestBuilder,
	req *domain.SendMessageRequest,
) error {
	upload, filePath, err := m.uploadMediaFile(ctx, api, req.MediaURL)
	if err != nil {
		return err
	}
	defer cleanupTempFile(filePath)

	text := getMessageText(req.Caption, req.Text)
	doc := message.UploadedDocument(upload, styling.Plain(text)).
		MIME("video/mp4").
		Filename(filepath.Base(filePath)).
		Video()

	_, err = builder.Media(ctx, doc)
	return err
}

// sendAudio sends an audio message.
func (m *ClientManager) sendAudio(
	ctx context.Context,
	api *tg.Client,
	builder *message.RequestBuilder,
	req *domain.SendMessageRequest,
) error {
	upload, filePath, err := m.uploadMediaFile(ctx, api, req.MediaURL)
	if err != nil {
		return err
	}
	defer cleanupTempFile(filePath)

	text := getMessageText(req.Caption, req.Text)
	doc := message.UploadedDocument(upload, styling.Plain(text)).
		MIME("audio/mpeg").
		Filename(filepath.Base(filePath)).
		Audio()

	_, err = builder.Media(ctx, doc)
	return err
}

// sendFile sends a file message.
func (m *ClientManager) sendFile(
	ctx context.Context,
	api *tg.Client,
	builder *message.RequestBuilder,
	req *domain.SendMessageRequest,
) error {
	upload, filePath, err := m.uploadMediaFile(ctx, api, req.MediaURL)
	if err != nil {
		return err
	}
	defer cleanupTempFile(filePath)

	text := getMessageText(req.Caption, req.Text)
	doc := message.UploadedDocument(upload, styling.Plain(text)).
		Filename(filepath.Base(filePath))

	_, err = builder.Media(ctx, doc)
	return err
}
