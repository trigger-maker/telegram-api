package domain

import (
	"time"

	"github.com/google/uuid"
)

// ==================== MESSAGE TYPES ====================

// MessageType represents the type of a message.
type MessageType string

const (
	// MessageTypeText represents a text message.
	MessageTypeText MessageType = "text"
	// MessageTypePhoto represents a photo message.
	MessageTypePhoto MessageType = "photo"
	// MessageTypeVideo represents a video message.
	MessageTypeVideo MessageType = "video"
	// MessageTypeAudio represents an audio message.
	MessageTypeAudio MessageType = "audio"
	// MessageTypeFile represents a file message.
	MessageTypeFile MessageType = "file"
)

// MessageStatus represents the status of a message.
type MessageStatus string

const (
	// MessageStatusPending represents a pending message.
	MessageStatusPending MessageStatus = "pending"
	// MessageStatusScheduled represents a scheduled message.
	MessageStatusScheduled MessageStatus = "scheduled"
	// MessageStatusSending represents a message being sent.
	MessageStatusSending MessageStatus = "sending"
	// MessageStatusSent represents a sent message.
	MessageStatusSent MessageStatus = "sent"
	// MessageStatusFailed represents a failed message.
	MessageStatusFailed MessageStatus = "failed"
)

// ==================== REQUEST DTOs ====================

// TextMessageRequest to send text message.
// @Description Simple text message.
type TextMessageRequest struct {
	To   string `json:"to" validate:"required" example:"@username o +573001234567"`
	Text string `json:"text" validate:"required" example:"Hello from the API!"`
}

// PhotoMessageRequest to send photo.
// @Description Message with photo.
type PhotoMessageRequest struct {
	To       string `json:"to" validate:"required" example:"@username"`
	PhotoURL string `json:"photo_url" validate:"required,url" example:"https://example.com/image.jpg"`
	Caption  string `json:"caption,omitempty" example:"Look at this image"`
}

// VideoMessageRequest to send video.
// @Description Message with video.
type VideoMessageRequest struct {
	To       string `json:"to" validate:"required" example:"@username"`
	VideoURL string `json:"video_url" validate:"required,url" example:"https://example.com/video.mp4"`
	Caption  string `json:"caption,omitempty" example:"Interesting video"`
}

// AudioMessageRequest to send audio.
// @Description Message with audio.
type AudioMessageRequest struct {
	To       string `json:"to" validate:"required" example:"@username"`
	AudioURL string `json:"audio_url" validate:"required,url" example:"https://example.com/audio.mp3"`
	Caption  string `json:"caption,omitempty" example:"Listen to this"`
}

// FileMessageRequest to send document.
// @Description Message with file/document.
type FileMessageRequest struct {
	To      string `json:"to" validate:"required" example:"@username"`
	FileURL string `json:"file_url" validate:"required,url" example:"https://example.com/doc.pdf"`
	Caption string `json:"caption,omitempty" example:"Attached document"`
}

// BulkTextRequest for bulk sending.
// @Description Bulk sending of text to multiple recipients.
type BulkTextRequest struct {
	Recipients []string `json:"recipients" validate:"required,min=1" example:"@user1,@user2,+573001234567"`
	Text       string   `json:"text" validate:"required" example:"Message for everyone"`
	DelayMs    int      `json:"delay_ms,omitempty" example:"3000"`
}

// ==================== INTERNAL REQUEST (para el servicio) ====================

// SendMessageRequest represents a request to send a single message.
type SendMessageRequest struct {
	To       string      `json:"to"`
	Text     string      `json:"text,omitempty"`
	Type     MessageType `json:"type,omitempty"`
	MediaURL string      `json:"media_url,omitempty"`
	Caption  string      `json:"caption,omitempty"`
	DelayMs  int         `json:"delay_ms,omitempty"`
}

// BulkMessageRequest represents a request to send bulk messages.
type BulkMessageRequest struct {
	Recipients []string    `json:"recipients"`
	Text       string      `json:"text"`
	Type       MessageType `json:"type,omitempty"`
	MediaURL   string      `json:"media_url,omitempty"`
	Caption    string      `json:"caption,omitempty"`
	DelayMs    int         `json:"delay_ms,omitempty"`
}

// ==================== RESPONSE DTOs ====================

// MessageResponse response when sending message.
// @Description Message sending response.
type MessageResponse struct {
	JobID   string        `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status  MessageStatus `json:"status" example:"pending"`
	SendAt  time.Time     `json:"send_at,omitempty"`
	Message string        `json:"message,omitempty" example:"Mensaje en cola"`
}

// MessageJob estado completo del job.
// @Description Estado detallado del mensaje.
type MessageJob struct {
	ID        string        `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SessionID uuid.UUID     `json:"session_id"`
	To        string        `json:"to" example:"@username"`
	Text      string        `json:"text,omitempty"`
	Type      MessageType   `json:"type" example:"text"`
	MediaURL  string        `json:"media_url,omitempty"`
	Caption   string        `json:"caption,omitempty"`
	Status    MessageStatus `json:"status" example:"sent"`
	Error     string        `json:"error,omitempty"`
	SendAt    time.Time     `json:"send_at"`
	SentAt    *time.Time    `json:"sent_at,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// Nota: Los errores están en errors.go (ErrSessionNotActive, etc.)
