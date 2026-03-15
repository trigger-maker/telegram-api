package telegram

import (
	"context"
	"errors"
	"testing"

	"telegram-api/internal/domain"
	"telegram-api/pkg/crypto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSessionRepository is a mock for SessionRepository
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

func (m *MockSessionRepository) GetByUserAndPhone(ctx context.Context, userID uuid.UUID, phone string) (*domain.TelegramSession, error) {
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

// Test 1: StoreSession with valid bytes - encryption and save
func TestPersistentSessionStorage_StoreSession_ValidBytes(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New().String()
	crypter, err := crypto.NewCrypter("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	mockRepo := new(MockSessionRepository)
	mockRepo.On("UpdateSessionData", sessionID, mock.AnythingOfType("[]uint8")).Return(nil)

	storage := NewPersistentSessionStorage(crypter, mockRepo, sessionID)

	testData := []byte("test session data")
	err = storage.StoreSession(ctx, testData)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "UpdateSessionData", sessionID, mock.AnythingOfType("[]uint8"))
}

// Test 2: LoadSession when record doesn't exist - empty slice, not error
func TestPersistentSessionStorage_LoadSession_NotFound(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New().String()
	crypter, err := crypto.NewCrypter("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	mockRepo := new(MockSessionRepository)
	mockRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(nil, domain.ErrSessionNotFound)

	storage := NewPersistentSessionStorage(crypter, mockRepo, sessionID)

	data, err := storage.LoadSession(ctx)

	assert.NoError(t, err)
	assert.Equal(t, []byte{}, data)
	mockRepo.AssertCalled(t, "GetByID", ctx, mock.AnythingOfType("uuid.UUID"))
}

// Test 3: StoreSession when DB unavailable - error, not panic
func TestPersistentSessionStorage_StoreSession_DBError(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New().String()
	crypter, err := crypto.NewCrypter("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	mockRepo := new(MockSessionRepository)
	mockRepo.On("UpdateSessionData", sessionID, mock.AnythingOfType("[]uint8")).Return(errors.New("500"))

	storage := NewPersistentSessionStorage(crypter, mockRepo, sessionID)

	testData := []byte("test session data")
	err = storage.StoreSession(ctx, testData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// Test 4: Service restart - automatic restore of active sessions
func TestPersistentSessionStorage_RestoreActiveSessions(t *testing.T) {
	ctx := context.Background()
	crypter, err := crypto.NewCrypter("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	sessionID := uuid.New()
	testData := []byte("restored session data")

	encryptedData, err := crypter.Encrypt(testData)
	require.NoError(t, err)

	session := &domain.TelegramSession{
		ID:          sessionID,
		SessionData: encryptedData,
	}

	mockRepo := new(MockSessionRepository)
	mockRepo.On("GetByID", ctx, sessionID).Return(session, nil)

	storage := NewPersistentSessionStorage(crypter, mockRepo, sessionID.String())

	data, err := storage.LoadSession(ctx)

	assert.NoError(t, err)
	assert.Equal(t, testData, data)
}

// Test 5: DC switch - save new bytes, continue work
func TestPersistentSessionStorage_DCSwitch(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New().String()
	crypter, err := crypto.NewCrypter("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	mockRepo := new(MockSessionRepository)

	storage := NewPersistentSessionStorage(crypter, mockRepo, sessionID)

	initialData := []byte("initial data")
	mockRepo.On("UpdateSessionData", sessionID, mock.AnythingOfType("[]uint8")).Return(nil).Once()

	err = storage.StoreSession(ctx, initialData)
	require.NoError(t, err)

	updatedData := []byte("updated data after DC switch")
	mockRepo.On("UpdateSessionData", sessionID, mock.AnythingOfType("[]uint8")).Return(nil).Once()

	err = storage.StoreSession(ctx, updatedData)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test 6: UpdateSessionData with non-existent session_id - repository error
func TestPersistentSessionStorage_UpdateSessionData_NonExistent(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New().String()
	crypter, err := crypto.NewCrypter("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	mockRepo := new(MockSessionRepository)
	mockRepo.On("UpdateSessionData", sessionID, mock.AnythingOfType("[]uint8")).Return(domain.ErrSessionNotFound)

	storage := NewPersistentSessionStorage(crypter, mockRepo, sessionID)

	testData := []byte("test session data")
	err = storage.StoreSession(ctx, testData)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)
}

// Test LoadSession with valid encrypted data
func TestPersistentSessionStorage_LoadSession_ValidData(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	crypter, err := crypto.NewCrypter("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	testData := []byte("test session data")
	encryptedData, err := crypter.Encrypt(testData)
	require.NoError(t, err)

	session := &domain.TelegramSession{
		ID:          sessionID,
		SessionData: encryptedData,
	}

	mockRepo := new(MockSessionRepository)
	mockRepo.On("GetByID", ctx, sessionID).Return(session, nil)

	storage := NewPersistentSessionStorage(crypter, mockRepo, sessionID.String())

	data, err := storage.LoadSession(ctx)

	assert.NoError(t, err)
	assert.Equal(t, testData, data)
}

// Test LoadSession with corrupted data
func TestPersistentSessionStorage_LoadSession_CorruptedData(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	crypter, err := crypto.NewCrypter("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	corruptedData := []byte("corrupted data")

	session := &domain.TelegramSession{
		ID:          sessionID,
		SessionData: corruptedData,
	}

	mockRepo := new(MockSessionRepository)
	mockRepo.On("GetByID", ctx, sessionID).Return(session, nil)

	storage := NewPersistentSessionStorage(crypter, mockRepo, sessionID.String())

	data, err := storage.LoadSession(ctx)

	assert.Error(t, err)
	assert.Equal(t, []byte{}, data)
}
