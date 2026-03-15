package telegram

import (
	"context"
	"errors"
	"testing"
	"time"

	"telegram-api/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandleMTProtoError tests MTProto error handling
func TestHandleMTProtoError(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()

	tests := []struct {
		name                string
		err                 error
		expectedAction       ErrorAction
		expectedState       domain.SessionStatus
		expectedPause        time.Duration
		expectedError       error
	}{
		{
			name:          "FLOOD_WAIT_30",
			err:           errors.New("FLOOD_WAIT_30"),
			expectedAction: ActionPause,
			expectedPause:  30 * time.Second,
			expectedError:  domain.ErrTelegramFloodWait,
		},
		{
			name:          "FLOOD_WAIT_120",
			err:           errors.New("FLOOD_WAIT_120"),
			expectedAction: ActionPause,
			expectedPause:  120 * time.Second,
			expectedError:  domain.ErrTelegramFloodWait,
		},
		{
			name:          "SESSION_REVOKED",
			err:           errors.New("SESSION_REVOKED"),
			expectedAction: ActionStop,
			expectedState:  domain.SessionBanned,
			expectedError:  domain.ErrSessionNotActive,
		},
		{
			name:          "AUTH_KEY_UNREGISTERED",
			err:           errors.New("AUTH_KEY_UNREGISTERED"),
			expectedAction: ActionStop,
			expectedState:  domain.SessionBanned,
			expectedError:  domain.ErrSessionNotActive,
		},
		{
			name:          "USER_DEACTIVATED",
			err:           errors.New("USER_DEACTIVATED"),
			expectedAction: ActionStop,
			expectedState:  domain.SessionFrozen,
			expectedError:  domain.ErrSessionNotActive,
		},
		{
			name:          "USER_DEACTIVATED_BAN",
			err:           errors.New("USER_DEACTIVATED_BAN"),
			expectedAction: ActionStop,
			expectedState:  domain.SessionBanned,
			expectedError:  domain.ErrSessionNotActive,
		},
		{
			name:          "PHONE_NUMBER_BANNED",
			err:           errors.New("PHONE_NUMBER_BANNED"),
			expectedAction: ActionStop,
			expectedState:  domain.SessionBanned,
			expectedError:  domain.ErrSessionNotActive,
		},
		{
			name:          "PEER_ID_INVALID",
			err:           errors.New("PEER_ID_INVALID"),
			expectedAction: ActionFailTask,
			expectedError:  domain.ErrPeerNotFound,
		},
		{
			name:          "USERNAME_NOT_OCCUPIED",
			err:           errors.New("USERNAME_NOT_OCCUPIED"),
			expectedAction: ActionFailTask,
			expectedError:  domain.ErrPeerNotFound,
		},
		{
			name:          "INPUT_USER_DEACTIVATED",
			err:           errors.New("INPUT_USER_DEACTIVATED"),
			expectedAction: ActionFailTask,
			expectedError:  domain.ErrPeerNotFound,
		},
		{
			name:          "Unknown error",
			err:           errors.New("UNKNOWN_ERROR"),
			expectedAction: ActionContinue,
			expectedError:  domain.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSessionRepository)

			if tt.expectedAction == ActionStop {
				mockRepo.On("GetByID", ctx, sessionID).Return(&domain.TelegramSession{
					ID:        sessionID,
					AuthState: domain.SessionAuthenticated,
					IsActive:  true,
				}, nil)
				mockRepo.On("Update", ctx, mock.Anything).Return(nil)
			}

			action, pause, err := HandleMTProtoError(ctx, sessionID, tt.err, mockRepo)

			assert.Equal(t, tt.expectedAction, action)
			assert.Equal(t, tt.expectedPause, pause)
			assert.Equal(t, tt.expectedError, err)

			if tt.expectedAction == ActionStop {
				mockRepo.AssertCalled(t, "Update", ctx, mock.MatchedBy(func(s *domain.TelegramSession) bool {
					return s.AuthState == tt.expectedState
				}))
			}
		})
	}
}

// TestExtractFloodWaitSeconds tests flood wait extraction
func TestExtractFloodWaitSeconds(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "FLOOD_WAIT_30",
			err:      errors.New("FLOOD_WAIT_30"),
			expected: 30,
		},
		{
			name:     "FLOOD_WAIT_120",
			err:      errors.New("FLOOD_WAIT_120"),
			expected: 120,
		},
		{
			name:     "No flood wait",
			err:      errors.New("SOME_OTHER_ERROR"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFloodWaitSeconds(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsBlockingError tests blocking error detection
func TestIsBlockingError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "SESSION_REVOKED",
			err:      errors.New("SESSION_REVOKED"),
			expected: true,
		},
		{
			name:     "AUTH_KEY_UNREGISTERED",
			err:      errors.New("AUTH_KEY_UNREGISTERED"),
			expected: true,
		},
		{
			name:     "USER_DEACTIVATED_BAN",
			err:      errors.New("USER_DEACTIVATED_BAN"),
			expected: true,
		},
		{
			name:     "PHONE_NUMBER_BANNED",
			err:      errors.New("PHONE_NUMBER_BANNED"),
			expected: true,
		},
		{
			name:     "USER_DEACTIVATED",
			err:      errors.New("USER_DEACTIVATED"),
			expected: true,
		},
		{
			name:     "FLOOD_WAIT_30",
			err:      errors.New("FLOOD_WAIT_30"),
			expected: false,
		},
		{
			name:     "Random error",
			err:      errors.New("RANDOM_ERROR"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBlockingError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsTaskError tests task error detection
func TestIsTaskError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "PEER_ID_INVALID",
			err:      errors.New("PEER_ID_INVALID"),
			expected: true,
		},
		{
			name:     "USERNAME_NOT_OCCUPIED",
			err:      errors.New("USERNAME_NOT_OCCUPIED"),
			expected: true,
		},
		{
			name:     "INPUT_USER_DEACTIVATED",
			err:      errors.New("INPUT_USER_DEACTIVATED"),
			expected: true,
		},
		{
			name:     "FLOOD_WAIT_30",
			err:      errors.New("FLOOD_WAIT_30"),
			expected: false,
		},
		{
			name:     "Random error",
			err:      errors.New("RANDOM_ERROR"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTaskError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetBannedState tests banned state determination
func TestGetBannedState(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected domain.SessionStatus
	}{
		{
			name:     "USER_DEACTIVATED",
			err:      errors.New("USER_DEACTIVATED"),
			expected: domain.SessionFrozen,
		},
		{
			name:     "SESSION_REVOKED",
			err:      errors.New("SESSION_REVOKED"),
			expected: domain.SessionBanned,
		},
		{
			name:     "AUTH_KEY_UNREGISTERED",
			err:      errors.New("AUTH_KEY_UNREGISTERED"),
			expected: domain.SessionBanned,
		},
		{
			name:     "USER_DEACTIVATED_BAN",
			err:      errors.New("USER_DEACTIVATED_BAN"),
			expected: domain.SessionBanned,
		},
		{
			name:     "PHONE_NUMBER_BANNED",
			err:      errors.New("PHONE_NUMBER_BANNED"),
			expected: domain.SessionBanned,
		},
		{
			name:     "Random error",
			err:      errors.New("RANDOM_ERROR"),
			expected: domain.SessionAuthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBannedState(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
