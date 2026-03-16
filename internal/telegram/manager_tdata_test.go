package telegram

import (
	"context"
	"errors"
	"testing"

	"telegram-api/internal/config"
	"telegram-api/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test 1: Invalid api_id - validation error.
func TestClientManager_ImportTData_InvalidApiID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockSessionRepository)
	sessionID := uuid.New().String()

	manager, err := NewManager(
		&config.Config{
			Encryption: config.EncryptionConfig{
				Key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		},
		mockRepo,
	)
	require.NoError(t, err)

	tdataFiles := map[string][]byte{
		"key_datas": []byte("mock key data"),
	}

	_, err = manager.ImportTData(ctx, 0, "test_api_hash", "test_session", sessionID, tdataFiles)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid api_id")
}

// Test 2: Empty api_hash - validation error.
func TestClientManager_ImportTData_EmptyApiHash(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockSessionRepository)
	sessionID := uuid.New().String()

	manager, err := NewManager(
		&config.Config{
			Encryption: config.EncryptionConfig{
				Key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		},
		mockRepo,
	)
	require.NoError(t, err)

	tdataFiles := map[string][]byte{
		"key_datas": []byte("mock key data"),
	}

	_, err = manager.ImportTData(ctx, 12345, "", "test_session", sessionID, tdataFiles)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_hash required")
}

// Test 3: No files - validation error.
func TestClientManager_ImportTData_NoFiles(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockSessionRepository)
	sessionID := uuid.New().String()

	manager, err := NewManager(
		&config.Config{
			Encryption: config.EncryptionConfig{
				Key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		},
		mockRepo,
	)
	require.NoError(t, err)

	tdataFiles := map[string][]byte{}

	_, err = manager.ImportTData(ctx, 12345, "test_api_hash", "test_session", sessionID, tdataFiles)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tdata files required")
}

// Test 4: Corrupted files - ErrTDataInvalid.
func TestClientManager_ImportTData_CorruptedFiles(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockSessionRepository)
	sessionID := uuid.New().String()

	manager, err := NewManager(
		&config.Config{
			Encryption: config.EncryptionConfig{
				Key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		},
		mockRepo,
	)
	require.NoError(t, err)

	tdataFiles := map[string][]byte{
		"key_datas": []byte("corrupted data"),
	}

	_, err = manager.ImportTData(ctx, 12345, "test_api_hash", "test_session", sessionID, tdataFiles)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrTDataInvalid, errors.Unwrap(err))
}

// Test 5: Valid input structure - passes validation.
func TestClientManager_ImportTData_ValidInput(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockSessionRepository)
	sessionID := uuid.New().String()

	manager, err := NewManager(
		&config.Config{
			Encryption: config.EncryptionConfig{
				Key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		},
		mockRepo,
	)
	require.NoError(t, err)

	tdataFiles := map[string][]byte{
		"key_datas": []byte("mock key data"),
	}

	// This will fail because we don't have real tdata files,
	// but it should pass validation
	_, err = manager.ImportTData(ctx, 12345, "test_api_hash", "test_session", sessionID, tdataFiles)

	// Should fail with ErrTDataInvalid (not validation error)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrTDataInvalid, errors.Unwrap(err))
}
