package telegram

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// EventDispatcher sends events to configured webhooks.
type EventDispatcher struct {
	webhookRepo domain.WebhookRepository
	httpClient  *http.Client
	eventChan   chan *dispatchJob
}

type dispatchJob struct {
	SessionID uuid.UUID
	Event     domain.WebhookEvent
}

// NewEventDispatcher creates the dispatcher.
func NewEventDispatcher(webhookRepo domain.WebhookRepository) *EventDispatcher {
	d := &EventDispatcher{
		webhookRepo: webhookRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		eventChan: make(chan *dispatchJob, 1000), // Buffer for 1000 events.
	}

	// Workers to send events.
	for i := 0; i < 10; i++ {
		go d.worker()
	}

	return d
}

// Dispatch envía un evento.
func (d *EventDispatcher) Dispatch(sessionID uuid.UUID, eventType domain.EventType, data interface{}) {
	event := domain.WebhookEvent{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	select {
	case d.eventChan <- &dispatchJob{SessionID: sessionID, Event: event}:
	default:
		logger.Warn().
			Str("session_id", sessionID.String()).
			Str("event_type", string(eventType)).
			Msg("⚠️ Event buffer full, event discarded.")
	}
}

func (d *EventDispatcher) worker() {
	for job := range d.eventChan {
		d.sendToWebhook(job)
	}
}

// prepareWebhookRequest prepares the HTTP request for webhook.
func (d *EventDispatcher) prepareWebhookRequest(
	ctx context.Context,
	job *dispatchJob,
	webhook *domain.WebhookConfig,
) (*http.Request, error) {
	payload, err := json.Marshal(job.Event)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-Event", string(job.Event.Type))
	req.Header.Set("X-Telegram-Session", job.SessionID.String())
	req.Header.Set("X-Telegram-Delivery", job.Event.ID)

	if webhook.Secret != "" {
		signature := d.signPayload(payload, webhook.Secret)
		req.Header.Set("X-Telegram-Signature", signature)
	}

	return req, nil
}

// sendWebhookWithRetry sends webhook request with retry logic.
func (d *EventDispatcher) sendWebhookWithRetry(req *http.Request, maxRetries int, job *dispatchJob) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// #nosec G704 -- Webhook URL is validated and user-configured
		resp, err := d.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		if err := resp.Body.Close(); err != nil {
			logger.Error().Err(err).Msg("Error closing response body")
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			logger.Debug().
				Str("session_id", job.SessionID.String()).
				Str("event_type", string(job.Event.Type)).
				Int("status", resp.StatusCode).
				Msg("✅ Event sent to webhook")
			return nil
		}

		lastErr = fmt.Errorf("webhook returned %d", resp.StatusCode)
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return lastErr
}

func (d *EventDispatcher) sendToWebhook(job *dispatchJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	webhook, err := d.webhookRepo.GetBySessionID(ctx, job.SessionID)
	if err != nil || webhook == nil || !webhook.IsActive {
		return
	}

	if !d.shouldSendEvent(webhook.Events, job.Event.Type) {
		return
	}

	req, err := d.prepareWebhookRequest(ctx, job, webhook)
	if err != nil {
		logger.Error().Err(err).Msg("Error preparing webhook request")
		return
	}

	maxRetries := webhook.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	lastErr := d.sendWebhookWithRetry(req, maxRetries, job)
	if lastErr != nil {
		logger.Error().
			Err(lastErr).
			Str("session_id", job.SessionID.String()).
			Str("url", webhook.URL).
			Msg("❌ Webhook failed after retries")
		go d.updateWebhookError(job.SessionID, lastErr.Error())
	}
}

func (d *EventDispatcher) shouldSendEvent(events []string, eventType domain.EventType) bool {
	if len(events) == 0 {
		return true // If no filter, send all
	}
	for _, e := range events {
		if e == string(eventType) || e == "*" {
			return true
		}
	}
	return false
}

func (d *EventDispatcher) signPayload(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

func (d *EventDispatcher) updateWebhookError(sessionID uuid.UUID, errMsg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	webhook, err := d.webhookRepo.GetBySessionID(ctx, sessionID)
	if err != nil || webhook == nil {
		return
	}

	now := time.Now()
	webhook.LastErrorAt = &now
	webhook.LastError = errMsg
	_ = d.webhookRepo.Update(ctx, webhook)
}
