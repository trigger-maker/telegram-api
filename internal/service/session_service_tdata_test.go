package service

import (
	"context"
	"testing"

	"telegram-api/internal/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test 1: ImportTData with invalid api_id - validation error.
func TestSessionService_ImportTData_InvalidApiID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	cfg := &config.Config{
		Encryption: config.EncryptionConfig{Key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
	}

	service := NewSessionService(nil, nil, nil, nil, cfg)

	tdataFiles := map[string][]byte{
		"key_datas": []byte("mock key data"),
	}

	_, err := service.ImportTData(ctx, userID, 0, "test_api_hash", "test_session", tdataFiles)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_id")
}

// Test 2: ImportTData with empty api_hash - validation error.
func TestSessionService_ImportTData_EmptyApiHash(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	cfg := &config.Config{
		Encryption: config.EncryptionConfig{Key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
	}

	service := NewSessionService(nil, nil, nil, nil, cfg)

	tdataFiles := map[string][]byte{
		"key_datas": []byte("mock key data"),
	}

	_, err := service.ImportTData(ctx, userID, 12345, "", "test_session", tdataFiles)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_hash")
}

// Test 3: ImportTData with no files - validation error.
func TestSessionService_ImportTData_NoFiles(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	cfg := &config.Config{
		Encryption: config.EncryptionConfig{Key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
	}

	service := NewSessionService(nil, nil, nil, nil, cfg)

	tdataFiles := map[string][]byte{}

	_, err := service.ImportTData(ctx, userID, 12345, "test_api_hash", "test_session", tdataFiles)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "files")
}

// Test 4: ImportTData with valid input structure - passes validation.
func TestSessionService_ImportTData_ValidInput(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	cfg := &config.Config{
		Encryption: config.EncryptionConfig{Key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
	}

	service := NewSessionService(nil, nil, nil, nil, cfg)

	tdataFiles := map[string][]byte{
		"key_datas": []byte("mock key data"),
	}

	// This will fail because we don't have real tdata files,
	// but it should pass validation
	// Since tgManager is nil, it will panic, so we expect that
	assert.Panics(t, func() {
		_, _ = service.ImportTData(ctx, userID, 12345, "test_api_hash", "test_session", tdataFiles)
	})
}
