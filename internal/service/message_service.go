package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/internal/telegram"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
	"github.com/gotd/td/tg"
)

// MessageServiceInterface defines the interface for MessageService.
type MessageServiceInterface interface {
	SendMessage(ctx context.Context, sessionID uuid.UUID, req *domain.SendMessageRequest) (*domain.MessageResponse, error)
	SendBulk(ctx context.Context, sessionID uuid.UUID, req *domain.BulkMessageRequest) ([]domain.MessageResponse, error)
	GetJobStatus(ctx context.Context, jobID string) (*domain.MessageJob, error)
}

// MessageService handles message sending operations.
type MessageService struct {
	sessionRepo domain.SessionRepository
	cache       domain.CacheRepository
	tgManager   *telegram.ClientManager
	pool        *telegram.SessionPool
}

// NewMessageService creates a new MessageService instance.
func NewMessageService(
	sRepo domain.SessionRepository,
	cache domain.CacheRepository,
	tgMgr *telegram.ClientManager,
	pool *telegram.SessionPool,
) *MessageService {
	return &MessageService{
		sessionRepo: sRepo,
		cache:       cache,
		tgManager:   tgMgr,
		pool:        pool,
	}
}

const (
	queueKey  = "tg:msg:queue"
	jobPrefix = "tg:msg:job:"
	jobTTL    = 86400 // 24 hours
)

// SendMessage sends a single message via Telegram.
func (s *MessageService) SendMessage(
	ctx context.Context,
	sessionID uuid.UUID,
	req *domain.SendMessageRequest,
) (*domain.MessageResponse, error) {
	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if !sess.IsActive || sess.AuthState != domain.SessionAuthenticated {
		return nil, domain.ErrSessionNotActive
	}

	if req.Type == "" {
		req.Type = domain.MessageTypeText
	}

	job := &domain.MessageJob{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		To:        req.To,
		Text:      req.Text,
		Type:      req.Type,
		MediaURL:  req.MediaURL,
		Caption:   req.Caption,
		Status:    domain.MessageStatusPending,
		CreatedAt: time.Now(),
	}

	if req.DelayMs > 0 {
		job.SendAt = time.Now().Add(time.Duration(req.DelayMs) * time.Millisecond)
		job.Status = domain.MessageStatusScheduled
	} else {
		job.SendAt = time.Now()
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		logger.Error().Err(err).Msg("Error marshaling job")
	}
	_ = s.cache.Set(ctx, jobPrefix+job.ID, string(jobData), jobTTL)

	if req.DelayMs > 0 {
		go s.scheduleJob(ctx, job)
	} else {
		go s.processJob(ctx, job)
	}

	return &domain.MessageResponse{
		JobID:   job.ID,
		Status:  job.Status,
		SendAt:  job.SendAt,
		Message: "Message queued",
	}, nil
}

// SendBulk sends messages to multiple recipients.
func (s *MessageService) SendBulk(
	ctx context.Context,
	sessionID uuid.UUID,
	req *domain.BulkMessageRequest,
) ([]domain.MessageResponse, error) {
	if s.pool == nil {
		return nil, domain.ErrSessionNotActive
	}

	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if !sess.IsActive {
		return nil, domain.ErrSessionNotActive
	}

	active, ok := s.pool.GetActiveSession(sessionID)
	if !ok {
		return nil, domain.ErrSessionNotActive
	}

	var responses []domain.MessageResponse

	for _, recipient := range req.Recipients {
		singleReq := &domain.SendMessageRequest{
			To:       recipient,
			Text:     req.Text,
			Type:     req.Type,
			MediaURL: req.MediaURL,
			Caption:  req.Caption,
		}

		if err := s.tgManager.SendMessageWithAPIClient(ctx, active.API, singleReq); err != nil {
			responses = append(responses, domain.MessageResponse{
				Status:  domain.MessageStatusFailed,
				Message: err.Error(),
			})
		} else {
			responses = append(responses, domain.MessageResponse{
				Status:  domain.MessageStatusSent,
				Message: "Message sent",
			})
		}
	}

	return responses, nil
}

// GetJobStatus retrieves the status of a message job.
func (s *MessageService) GetJobStatus(ctx context.Context, jobID string) (*domain.MessageJob, error) {
	data, err := s.cache.Get(ctx, jobPrefix+jobID)
	if err != nil || data == "" {
		return nil, fmt.Errorf("job not found")
	}

	var job domain.MessageJob
	if err := json.Unmarshal([]byte(data), &job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *MessageService) scheduleJob(ctx context.Context, job *domain.MessageJob) {
	delay := time.Until(job.SendAt)
	if delay > 0 {
		time.Sleep(delay)
	}
	s.processJob(ctx, job)
}

// SendMessageWithClient sends a message using a specific Telegram client.
func (s *MessageService) SendMessageWithClient(
	ctx context.Context,
	sessionID uuid.UUID,
	api interface{},
	req *domain.SendMessageRequest,
) (*domain.MessageResponse, error) {
	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if !sess.IsActive || sess.AuthState != domain.SessionAuthenticated {
		return nil, domain.ErrSessionNotActive
	}

	if req.Type == "" {
		req.Type = domain.MessageTypeText
	}

	job := &domain.MessageJob{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		To:        req.To,
		Text:      req.Text,
		Type:      req.Type,
		MediaURL:  req.MediaURL,
		Caption:   req.Caption,
		Status:    domain.MessageStatusPending,
		CreatedAt: time.Now(),
	}

	if req.DelayMs > 0 {
		job.SendAt = time.Now().Add(time.Duration(req.DelayMs) * time.Millisecond)
		job.Status = domain.MessageStatusScheduled
	} else {
		job.SendAt = time.Now()
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		logger.Error().Err(err).Msg("Error marshaling job")
	}
	_ = s.cache.Set(ctx, jobPrefix+job.ID, string(jobData), jobTTL)

	if req.DelayMs > 0 {
		go s.scheduleJob(ctx, job)
	} else {
		go s.processJobWithClient(ctx, job, api)
	}

	return &domain.MessageResponse{
		JobID:   job.ID,
		Status:  job.Status,
		SendAt:  job.SendAt,
		Message: "Message queued",
	}, nil
}

func (s *MessageService) processJob(ctx context.Context, job *domain.MessageJob) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	job.Status = domain.MessageStatusSending
	s.updateJob(ctx, job)

	// Check if session is active in pool
	if s.pool != nil {
		active, ok := s.pool.GetActiveSession(job.SessionID)
		if !ok {
			job.Status = domain.MessageStatusFailed
			job.Error = domain.ErrSessionNotActive.Error()
			s.updateJob(ctx, job)
			return
		}

		req := &domain.SendMessageRequest{
			To:       job.To,
			Text:     job.Text,
			Type:     job.Type,
			MediaURL: job.MediaURL,
			Caption:  job.Caption,
		}

		if err := s.tgManager.SendMessageWithAPIClient(ctx, active.API, req); err != nil {
			job.Status = domain.MessageStatusFailed
			job.Error = err.Error()
			logger.Error().Err(err).Str("job", job.ID).Msg("message failed")
		} else {
			job.Status = domain.MessageStatusSent
			now := time.Now()
			job.SentAt = &now
			logger.Info().Str("job", job.ID).Str("to", job.To).Msg("message sent")
		}

		s.updateJob(ctx, job)
	}
}

func (s *MessageService) processJobWithClient(ctx context.Context, job *domain.MessageJob, api interface{}) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	job.Status = domain.MessageStatusSending
	s.updateJob(ctx, job)

	req := &domain.SendMessageRequest{
		To:       job.To,
		Text:     job.Text,
		Type:     job.Type,
		MediaURL: job.MediaURL,
		Caption:  job.Caption,
	}

	if err := s.tgManager.SendMessageWithAPIClient(ctx, api.(*tg.Client), req); err != nil {
		job.Status = domain.MessageStatusFailed
		job.Error = err.Error()
		logger.Error().Err(err).Str("job", job.ID).Msg("message failed")
	} else {
		job.Status = domain.MessageStatusSent
		now := time.Now()
		job.SentAt = &now
		logger.Info().Str("job", job.ID).Str("to", job.To).Msg("message sent")
	}

	s.updateJob(ctx, job)
}

func (s *MessageService) updateJob(ctx context.Context, job *domain.MessageJob) {
	jobData, err := json.Marshal(job)
	if err != nil {
		logger.Error().Err(err).Msg("Error marshaling job")
	}
	_ = s.cache.Set(ctx, jobPrefix+job.ID, string(jobData), jobTTL)
}
