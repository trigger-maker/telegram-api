package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/internal/service"
	"telegram-api/internal/telegram"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock services for regression tests.
type MockSessionServiceRegression struct {
	mock.Mock
}

func (m *MockSessionServiceRegression) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	req *domain.CreateSessionRequest,
) (*domain.TelegramSession, string, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*domain.TelegramSession), args.String(1), args.Error(2)
}

func (m *MockSessionServiceRegression) VerifyCode(
	ctx context.Context,
	sessionID uuid.UUID,
	code string,
) (*domain.TelegramSession, string, error) {
	args := m.Called(ctx, sessionID, code)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*domain.TelegramSession), args.String(1), args.Error(2)
}

func (m *MockSessionServiceRegression) SubmitPassword(
	ctx context.Context,
	sessionID uuid.UUID,
	password string,
) (*domain.TelegramSession, error) {
	args := m.Called(ctx, sessionID, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TelegramSession), args.Error(1)
}

func (m *MockSessionServiceRegression) GetSession(
	ctx context.Context,
	sessionID uuid.UUID,
) (*domain.TelegramSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TelegramSession), args.Error(1)
}

func (m *MockSessionServiceRegression) ListSessions(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.TelegramSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TelegramSession), args.Error(1)
}

func (m *MockSessionServiceRegression) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionServiceRegression) RegenerateQR(ctx context.Context, sessionID uuid.UUID) (string, error) {
	args := m.Called(ctx, sessionID)
	return args.String(0), args.Error(1)
}

func (m *MockSessionServiceRegression) ImportTData(
	ctx context.Context,
	userID uuid.UUID,
	apiID int,
	apiHash string,
	sessionName string,
	tdataFiles map[string][]byte,
) (*domain.TelegramSession, error) {
	args := m.Called(ctx, userID, apiID, apiHash, sessionName, tdataFiles)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TelegramSession), args.Error(1)
}

var _ service.SessionServiceInterface = (*MockSessionServiceRegression)(nil)

type MockMessageServiceRegression struct {
	mock.Mock
}

func (m *MockMessageServiceRegression) SendMessage(
	ctx context.Context,
	sessionID uuid.UUID,
	req *domain.SendMessageRequest,
) (*domain.MessageResponse, error) {
	args := m.Called(ctx, sessionID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MessageResponse), args.Error(1)
}

func (m *MockMessageServiceRegression) SendBulk(
	ctx context.Context,
	sessionID uuid.UUID,
	req *domain.BulkMessageRequest,
) ([]domain.MessageResponse, error) {
	args := m.Called(ctx, sessionID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MessageResponse), args.Error(1)
}

func (m *MockMessageServiceRegression) GetJobStatus(ctx context.Context, jobID string) (*domain.MessageJob, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MessageJob), args.Error(1)
}

var _ service.MessageServiceInterface = (*MockMessageServiceRegression)(nil)

// R.1. SMS authorization without 2FA - no changes.
func TestRegression_SMSAuthorization(t *testing.T) {
	app := fiber.New()
	mockService := new(MockSessionServiceRegression)
	handler := NewSessionHandler(mockService)

	userID := uuid.New()
	sessionID := uuid.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return handler.Create(c)
	})

	req := domain.CreateSessionRequest{
		APIID:       12345,
		APIHash:     "12345678901234567890123456789012",
		Phone:       "+1234567890",
		SessionName: "test_session",
		AuthMethod:  domain.AuthMethodSMS,
	}

	mockService.On("CreateSession", mock.Anything, userID, &req).
		Return(&domain.TelegramSession{
			ID:          sessionID,
			UserID:      userID,
			PhoneNumber: "+1234567890",
			AuthState:   domain.SessionCodeSent,
		}, "phone_code_hash_123", nil)

	body, err := json.Marshal(req)
	assert.NoError(t, err)
	httpReq := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var result Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.Success)
	assert.Contains(t, result.Data, "phone_code_hash")
	assert.Contains(t, result.Data, "next_step")
	mockService.AssertExpectations(t)
}

// R.2. QR authorization - no changes, QR not printed.
func TestRegression_QRAuthorization(t *testing.T) {
	app := fiber.New()
	mockService := new(MockSessionServiceRegression)
	handler := NewSessionHandler(mockService)

	userID := uuid.New()
	sessionID := uuid.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return handler.Create(c)
	})

	req := domain.CreateSessionRequest{
		APIID:       12345,
		APIHash:     "12345678901234567890123456789012",
		Phone:       "",
		SessionName: "test_qr_session",
		AuthMethod:  domain.AuthMethodQR,
	}

	mockService.On("CreateSession", mock.Anything, userID, &req).
		Return(&domain.TelegramSession{
			ID:        sessionID,
			UserID:    userID,
			AuthState: domain.SessionPending,
		}, "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==", nil)

	body, err := json.Marshal(req)
	assert.NoError(t, err)
	httpReq := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var result Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.Success)
	assert.Contains(t, result.Data, "qr_image_base64")
	assert.Contains(t, result.Data, "message")
	mockService.AssertExpectations(t)
}

// R.3. QR regenerate - no changes.
func TestRegression_QRRegenerate(t *testing.T) {
	app := fiber.New()
	mockService := new(MockSessionServiceRegression)
	handler := NewSessionHandler(mockService)

	userID := uuid.New()
	sessionID := uuid.New()

	app.Post("/test/:id", func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return handler.RegenerateQR(c)
	})

	mockService.On("RegenerateQR", mock.Anything, sessionID).
		Return("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==", nil)

	httpReq := httptest.NewRequest(
		"POST",
		"/test/"+sessionID.String(),
		nil,
	)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.Success)
	assert.Contains(t, result.Data, "qr_image_base64")
	mockService.AssertExpectations(t)
}

// mediaTestCase defines a test case for media message sending.
type mediaTestCase struct {
	name      string
	endpoint  string
	req       interface{}
	mediaType domain.MessageType
	mediaURL  string
}

// getMediaTestCases returns all media test cases.
func getMediaTestCases() []mediaTestCase {
	return []mediaTestCase{
		{
			name:     "photo",
			endpoint: "/photo",
			req: domain.PhotoMessageRequest{
				To:       "@testuser",
				PhotoURL: "https://example.com/photo.jpg",
				Caption:  "Test photo",
			},
			mediaType: domain.MessageTypePhoto,
			mediaURL:  "https://example.com/photo.jpg",
		},
		{
			name:     "video",
			endpoint: "/video",
			req: domain.VideoMessageRequest{
				To:       "@testuser",
				VideoURL: "https://example.com/video.mp4",
				Caption:  "Test video",
			},
			mediaType: domain.MessageTypeVideo,
			mediaURL:  "https://example.com/video.mp4",
		},
		{
			name:     "audio",
			endpoint: "/audio",
			req: domain.AudioMessageRequest{
				To:       "@testuser",
				AudioURL: "https://example.com/audio.mp3",
				Caption:  "Test audio",
			},
			mediaType: domain.MessageTypeAudio,
			mediaURL:  "https://example.com/audio.mp3",
		},
		{
			name:     "file",
			endpoint: "/file",
			req: domain.FileMessageRequest{
				To:      "@testuser",
				FileURL: "https://example.com/document.pdf",
				Caption: "Test file",
			},
			mediaType: domain.MessageTypeFile,
			mediaURL:  "https://example.com/document.pdf",
		},
	}
}

// getMediaHandler returns the appropriate handler function for the media type.
func getMediaHandler(handler *MessageHandler, name string) fiber.Handler {
	switch name {
	case "photo":
		return handler.SendPhoto
	case "video":
		return handler.SendVideo
	case "audio":
		return handler.SendAudio
	case "file":
		return handler.SendFile
	default:
		return nil
	}
}

// testMediaType tests sending a specific media type.
func testMediaType(t *testing.T, tt mediaTestCase, mockService *MockMessageServiceRegression) {
	app := fiber.New()
	handler := NewMessageHandler(mockService)

	sessionID := uuid.New()
	jobID := uuid.New().String()

	mediaHandler := getMediaHandler(handler, tt.name)
	if mediaHandler == nil {
		t.Fatal("Invalid media type")
	}

	app.Post("/test/:id", func(c *fiber.Ctx) error {
		return mediaHandler(c)
	})

	mockService.On("SendMessage", mock.Anything, sessionID, mock.MatchedBy(func(r *domain.SendMessageRequest) bool {
		return r.Type == tt.mediaType && r.MediaURL == tt.mediaURL
	})).Return(&domain.MessageResponse{
		JobID:  jobID,
		Status: domain.MessageStatusPending,
		SendAt: time.Now(),
	}, nil)

	body, err := json.Marshal(tt.req)
	assert.NoError(t, err)
	httpReq := httptest.NewRequest("POST", "/test/"+sessionID.String(), bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	if resp.StatusCode != 202 {
		var errorResp Response
		bodyBytes, err := json.Marshal(resp.Body)
		assert.NoError(t, err)
		t.Logf("Response body: %s", string(bodyBytes))
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Logf("Decode error: %v", err)
		} else {
			t.Logf("Error response: %+v", errorResp)
			if errorResp.Error != nil {
				t.Logf("Error code: %d, message: %s", errorResp.Error.Code, errorResp.Error.Message)
			}
		}
	}
	assert.Equal(t, 202, resp.StatusCode)

	var result Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.Success)
	mockService.AssertExpectations(t)
}

// R.4. Send all media types - no changes.
func TestRegression_SendMediaTypes(t *testing.T) {
	mediaTypes := getMediaTestCases()

	for _, tt := range mediaTypes {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMessageServiceRegression)
			testMediaType(t, tt, mockService)
		})
	}
}

// R.5. Bulk send - no changes.
func TestRegression_BulkSend(t *testing.T) {
	app := fiber.New()
	mockService := new(MockMessageServiceRegression)
	handler := NewMessageHandler(mockService)

	sessionID := uuid.New()
	jobID1 := uuid.New().String()
	jobID2 := uuid.New().String()

	app.Post("/test/:id", func(c *fiber.Ctx) error {
		return handler.SendBulk(c)
	})

	req := domain.BulkTextRequest{
		Recipients: []string{"@user1", "@user2", "+1234567890"},
		Text:       "Bulk test message",
		DelayMs:    1000,
	}

	mockService.On("SendBulk", mock.Anything, sessionID, mock.MatchedBy(func(r *domain.BulkMessageRequest) bool {
		return len(r.Recipients) == 3 && r.Text == "Bulk test message"
	})).Return([]domain.MessageResponse{
		{JobID: jobID1, Status: domain.MessageStatusPending, SendAt: time.Now()},
		{JobID: jobID2, Status: domain.MessageStatusPending, SendAt: time.Now()},
	}, nil)

	body, err := json.Marshal(req)
	assert.NoError(t, err)
	httpReq := httptest.NewRequest("POST", "/test/"+sessionID.String(), bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, 202, resp.StatusCode)

	var result Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.Success)
	mockService.AssertExpectations(t)
}

// R.6. Webhook event delivery - no changes.
func TestRegression_WebhookDelivery(t *testing.T) {
	app := fiber.New()
	mockWebhookRepo := new(MockWebhookRepository)
	mockSessionRepo := new(MockSessionRepository)
	pool := new(MockSessionPool)
	handler := NewWebhookHandler(mockWebhookRepo, mockSessionRepo, pool)

	sessionID := uuid.New()

	app.Post("/test/:id", func(c *fiber.Ctx) error {
		return handler.Configure(c)
	})

	req := domain.WebhookCreateRequest{
		URL:        "https://example.com/webhook",
		Secret:     "test_secret",
		Events:     []string{"message.new", "message.edit"},
		MaxRetries: 3,
		TimeoutMs:  5000,
	}

	mockSessionRepo.On("GetByID", mock.Anything, sessionID).
		Return(&domain.TelegramSession{
			ID:        sessionID,
			IsActive:  true,
			AuthState: domain.SessionAuthenticated,
		}, nil)

	mockWebhookRepo.On("Create", mock.Anything, mock.MatchedBy(func(wh *domain.WebhookConfig) bool {
		return wh.SessionID == sessionID && wh.URL == "https://example.com/webhook"
	})).Return(nil)

	// #nosec G117 -- Test code, not exposing secrets
	body, err := json.Marshal(req)
	assert.NoError(t, err)
	httpReq := httptest.NewRequest("POST", "/test/"+sessionID.String(), bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.Success)
	assert.Contains(t, result.Data, "id")
	assert.Contains(t, result.Data, "url")
	assert.Contains(t, result.Data, "events")
	mockWebhookRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

// R.7. GET /messages/:jobId/status - no changes.
func TestRegression_GetMessageStatus(t *testing.T) {
	app := fiber.New()
	mockService := new(MockMessageServiceRegression)
	handler := NewMessageHandler(mockService)

	jobID := uuid.New().String()
	sessionID := uuid.New()
	sentAt := time.Now()

	app.Get("/test/:jobId", func(c *fiber.Ctx) error {
		return handler.GetStatus(c)
	})

	mockService.On("GetJobStatus", mock.Anything, jobID).
		Return(&domain.MessageJob{
			ID:        jobID,
			SessionID: sessionID,
			To:        "@testuser",
			Text:      "Test message",
			Type:      domain.MessageTypeText,
			Status:    domain.MessageStatusSent,
			SendAt:    sentAt,
			SentAt:    &sentAt,
		}, nil)

	httpReq := httptest.NewRequest("GET", "/test/"+jobID, nil)

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.Success)
	mockService.AssertExpectations(t)
}

// R.8. DELETE /sessions/:id - no changes.
func TestRegression_DeleteSession(t *testing.T) {
	app := fiber.New()
	mockService := new(MockSessionServiceRegression)
	handler := NewSessionHandler(mockService)

	userID := uuid.New()
	sessionID := uuid.New()

	app.Delete("/test/:id", func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return handler.Delete(c)
	})

	mockService.On("DeleteSession", mock.Anything, sessionID).Return(nil)

	httpReq := httptest.NewRequest("DELETE", "/test/"+sessionID.String(), nil)

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.Success)
	mockService.AssertExpectations(t)
}

// R.9. Incoming webhook message.new - no changes.
func TestRegression_IncomingWebhookMessageNew(t *testing.T) {
	app := fiber.New()
	mockWebhookRepo := new(MockWebhookRepository)
	mockSessionRepo := new(MockSessionRepository)
	pool := new(MockSessionPool)
	handler := NewWebhookHandler(mockWebhookRepo, mockSessionRepo, pool)

	sessionID := uuid.New()
	webhookID := uuid.New()

	app.Get("/test/:id", func(c *fiber.Ctx) error {
		return handler.Get(c)
	})

	mockWebhookRepo.On("GetBySessionID", mock.Anything, sessionID).
		Return(&domain.WebhookConfig{
			ID:        webhookID,
			SessionID: sessionID,
			URL:       "https://example.com/webhook",
			Events:    []string{"message.new", "message.edit"},
			IsActive:  true,
		}, nil)

	httpReq := httptest.NewRequest("GET", "/test/"+sessionID.String(), nil)

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.Success)
	webhookData := result.Data.(map[string]interface{})
	assert.Equal(t, "https://example.com/webhook", webhookData["url"])
	assert.Contains(t, webhookData["events"], "message.new")
	mockWebhookRepo.AssertExpectations(t)
}

// Mock repositories for webhook tests.
type MockWebhookRepository struct {
	mock.Mock
}

func (m *MockWebhookRepository) Create(ctx context.Context, wh *domain.WebhookConfig) error {
	args := m.Called(ctx, wh)
	return args.Error(0)
}

func (m *MockWebhookRepository) Update(ctx context.Context, wh *domain.WebhookConfig) error {
	args := m.Called(ctx, wh)
	return args.Error(0)
}

func (m *MockWebhookRepository) GetBySessionID(
	ctx context.Context,
	sessionID uuid.UUID,
) (*domain.WebhookConfig, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WebhookConfig), args.Error(1)
}

func (m *MockWebhookRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockWebhookRepository) ListActive(ctx context.Context) ([]domain.WebhookConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.WebhookConfig), args.Error(1)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.TelegramSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TelegramSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TelegramSession), args.Error(1)
}

func (m *MockSessionRepository) GetByPhone(ctx context.Context, phone string) (*domain.TelegramSession, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TelegramSession), args.Error(1)
}

func (m *MockSessionRepository) GetByUserAndPhone(
	ctx context.Context,
	userID uuid.UUID,
	phone string,
) (*domain.TelegramSession, error) {
	args := m.Called(ctx, userID, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TelegramSession), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *domain.TelegramSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.TelegramSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TelegramSession), args.Error(1)
}

func (m *MockSessionRepository) ListAllActive(ctx context.Context) ([]domain.TelegramSession, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TelegramSession), args.Error(1)
}

func (m *MockSessionRepository) UpdateSessionData(sessionID string, data []byte) error {
	args := m.Called(sessionID, data)
	return args.Error(0)
}

// MockSessionPool for webhook tests.
type MockSessionPool struct {
	mock.Mock
}

func (m *MockSessionPool) StartSession(ctx context.Context, sess *domain.TelegramSession) error {
	args := m.Called(ctx, sess)
	return args.Error(0)
}

func (m *MockSessionPool) StopSession(sessionID uuid.UUID) {
	m.Called(sessionID)
}

func (m *MockSessionPool) GetActiveSession(sessionID uuid.UUID) (*telegram.ActiveSession, bool) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*telegram.ActiveSession), args.Bool(1)
}

func (m *MockSessionPool) ListActive() []uuid.UUID {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]uuid.UUID)
}

var _ telegram.SessionPoolInterface = (*MockSessionPool)(nil)
