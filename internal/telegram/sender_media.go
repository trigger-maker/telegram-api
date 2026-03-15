package telegram

import (
	"context"
	"os"
	"path/filepath"

	"telegram-api/internal/domain"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

// sendPhoto sends a photo message
func (m *ClientManager) sendPhoto(ctx context.Context, api *tg.Client, builder *message.RequestBuilder, req *domain.SendMessageRequest) error {
	filePath, err := m.downloadFile(req.MediaURL)
	if err != nil {
		return err
	}
	defer os.Remove(filePath)

	up := uploader.NewUploader(api)
	upload, err := up.FromPath(ctx, filePath)
	if err != nil {
		return err
	}

	text := req.Caption
	if text == "" {
		text = req.Text
	}

	photo := message.UploadedPhoto(upload, styling.Plain(text))
	_, err = builder.Media(ctx, photo)
	return err
}

// sendVideo sends a video message
func (m *ClientManager) sendVideo(ctx context.Context, api *tg.Client, builder *message.RequestBuilder, req *domain.SendMessageRequest) error {
	filePath, err := m.downloadFile(req.MediaURL)
	if err != nil {
		return err
	}
	defer os.Remove(filePath)

	up := uploader.NewUploader(api)
	upload, err := up.FromPath(ctx, filePath)
	if err != nil {
		return err
	}

	text := req.Caption
	if text == "" {
		text = req.Text
	}

	doc := message.UploadedDocument(upload, styling.Plain(text)).
		MIME("video/mp4").
		Filename(filepath.Base(filePath)).
		Video()

	_, err = builder.Media(ctx, doc)
	return err
}

// sendAudio sends an audio message
func (m *ClientManager) sendAudio(ctx context.Context, api *tg.Client, builder *message.RequestBuilder, req *domain.SendMessageRequest) error {
	filePath, err := m.downloadFile(req.MediaURL)
	if err != nil {
		return err
	}
	defer os.Remove(filePath)

	up := uploader.NewUploader(api)
	upload, err := up.FromPath(ctx, filePath)
	if err != nil {
		return err
	}

	text := req.Caption
	if text == "" {
		text = req.Text
	}

	doc := message.UploadedDocument(upload, styling.Plain(text)).
		MIME("audio/mpeg").
		Filename(filepath.Base(filePath)).
		Audio()

	_, err = builder.Media(ctx, doc)
	return err
}

// sendFile sends a file message
func (m *ClientManager) sendFile(ctx context.Context, api *tg.Client, builder *message.RequestBuilder, req *domain.SendMessageRequest) error {
	filePath, err := m.downloadFile(req.MediaURL)
	if err != nil {
		return err
	}
	defer os.Remove(filePath)

	up := uploader.NewUploader(api)
	upload, err := up.FromPath(ctx, filePath)
	if err != nil {
		return err
	}

	text := req.Caption
	if text == "" {
		text = req.Text
	}

	doc := message.UploadedDocument(upload, styling.Plain(text)).
		Filename(filepath.Base(filePath))

	_, err = builder.Media(ctx, doc)
	return err
}
