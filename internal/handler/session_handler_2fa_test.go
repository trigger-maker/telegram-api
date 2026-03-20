package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"telegram-api/internal/domain"
	"telegram-api/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSessionService for 2FA tests.
type MockSessionService2FA struct {
	mock.Mock
}

func (m *MockSessionService2FA) CreateSession(
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

func (m *MockSessionService2FA) VerifyCode(
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

// submitPasswordTestHelper executes a submit password test with given parameters.
func submitPasswordTestHelper(
	t *testing.T,
	mockService *MockSessionService2FA,
	handler *SessionHandler,
	sessionID uuid.UUID,
	password string,
	expectedStatusCode int,
) {
	app := fiber.New()
	userID := uuid.New()

	app.Post("/test/:id", func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return handler.SubmitPassword(c)
	})

	req := map[string]string{"password": password}

	body, err := json.Marshal(req)
	assert.NoError(t, err)
	httpReq := httptest.NewRequest("POST", "/test/"+sessionID.String(), bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedStatusCode, resp.StatusCode)

	var result Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result.Success)
	if result.Error.Code != 0 {
		assert.Equal(t, expectedStatusCode, result.Error.Code)
	}
	mockService.AssertExpectations(t)
}

func (m *MockSessionService2FA) SubmitPassword(
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

func (m *MockSessionService2FA) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.TelegramSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TelegramSession), args.Error(1)
}

func (m *MockSessionService2FA) ListSessions(ctx context.Context, userID uuid.UUID) ([]domain.TelegramSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TelegramSession), args.Error(1)
}

func (m *MockSessionService2FA) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionService2FA) RegenerateQR(ctx context.Context, sessionID uuid.UUID) (string, error) {
	args := m.Called(ctx, sessionID)
	return args.String(0), args.Error(1)
}

func (m *MockSessionService2FA) ImportTData(
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

var _ service.SessionServiceInterface = (*MockSessionService2FA)(nil)

// Test 1: POST /sessions without 2FA - no changes.
func TestSessionHandler_CreateSession_No2FA(t *testing.T) {
	app := fiber.New()
	mockService := new(MockSessionService2FA)
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
	mockService.AssertExpectations(t)
}

// Test 2: POST /sessions/:id/verify with 2FA - auth_state: password_required.
func TestSessionHandler_VerifyCode_With2FA(t *testing.T) {
	app := fiber.New()
	mockService := new(MockSessionService2FA)
	handler := NewSessionHandler(mockService)

	userID := uuid.New()
	sessionID := uuid.New()

	app.Post("/test/:id", func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return handler.VerifyCode(c)
	})

	req := domain.VerifyCodeRequest{Code: "123456"}

	session := &domain.TelegramSession{
		ID:          sessionID,
		UserID:      userID,
		PhoneNumber: "+1234567890",
		AuthState:   domain.SessionPasswordRequired,
	}

	mockService.On("VerifyCode", mock.Anything, sessionID, "123456").
		Return(session, "", nil)

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

	sessionData := result.Data.(map[string]interface{})
	assert.Equal(t, "password_required", sessionData["auth_state"])
	mockService.AssertExpectations(t)
}

// Test 3: POST /sessions/:id/submit-password correct - authenticated.
func TestSessionHandler_SubmitPassword_Correct(t *testing.T) {
	app := fiber.New()
	mockService := new(MockSessionService2FA)
	handler := NewSessionHandler(mockService)

	userID := uuid.New()
	sessionID := uuid.New()

	app.Post("/test/:id", func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return handler.SubmitPassword(c)
	})

	req := map[string]string{"password": "correct_password"}

	session := &domain.TelegramSession{
		ID:               sessionID,
		UserID:           userID,
		PhoneNumber:      "+1234567890",
		AuthState:        domain.SessionAuthenticated,
		IsActive:         true,
		TelegramUserID:   12345,
		TelegramUsername: "testuser",
	}

	mockService.On("SubmitPassword", mock.Anything, sessionID, "correct_password").
		Return(session, nil)

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

	sessionData := result.Data.(map[string]interface{})
	assert.Equal(t, "authenticated", sessionData["auth_state"])
	assert.True(t, sessionData["is_active"].(bool))
	mockService.AssertExpectations(t)
}

// Test 4: POST /sessions/:id/submit-password incorrect - INVALID_PASSWORD.
func TestSessionHandler_SubmitPassword_Incorrect(t *testing.T) {
	mockService := new(MockSessionService2FA)
	handler := NewSessionHandler(mockService)
	sessionID := uuid.New()

	mockService.On("SubmitPassword", mock.Anything, sessionID, "wrong_password").
		Return(nil, domain.ErrInvalidPassword)

	submitPasswordTestHelper(t, mockService, handler, sessionID, "wrong_password", 400)
}

// Test 5: POST /sessions/:id/submit-password non-existent - 404.
func TestSessionHandler_SubmitPassword_NotFound(t *testing.T) {
	mockService := new(MockSessionService2FA)
	handler := NewSessionHandler(mockService)
	sessionID := uuid.New()

	mockService.On("SubmitPassword", mock.Anything, sessionID, "password").
		Return(nil, domain.ErrSessionNotFound)

	submitPasswordTestHelper(t, mockService, handler, sessionID, "password", 404)
}

// Test 6: POST /sessions/:id/submit-password without password_required - 400/409.
func TestSessionHandler_SubmitPassword_WrongState(t *testing.T) {
	mockService := new(MockSessionService2FA)
	handler := NewSessionHandler(mockService)
	sessionID := uuid.New()

	mockService.On("SubmitPassword", mock.Anything, sessionID, "password").
		Return(nil, domain.NewAppError(nil, "Wrong state", 409))

	submitPasswordTestHelper(t, mockService, handler, sessionID, "password", 409)
}

// Test 7: Repeated submit-password after success - 409.
func TestSessionHandler_SubmitPassword_AlreadyAuthenticated(t *testing.T) {
	mockService := new(MockSessionService2FA)
	handler := NewSessionHandler(mockService)
	sessionID := uuid.New()

	mockService.On("SubmitPassword", mock.Anything, sessionID, "password").
		Return(nil, domain.NewAppError(nil, "Already authenticated", 409))

	submitPasswordTestHelper(t, mockService, handler, sessionID, "password", 409)
}
