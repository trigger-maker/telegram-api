package telegram

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// ErrorAction represents the action to take on an error
type ErrorAction string

const (
	ActionContinue  ErrorAction = "continue"
	ActionPause    ErrorAction = "pause"
	ActionStop     ErrorAction = "stop"
	ActionFailTask ErrorAction = "fail_task"
)

// HandleMTProtoError handles MTProto errors centrally
func HandleMTProtoError(
	ctx context.Context,
	sessionID uuid.UUID,
	err error,
	repo domain.SessionRepository,
) (ErrorAction, time.Duration, error) {
	if err == nil {
		return ActionContinue, 0, nil
	}

	errStr := err.Error()
	errorCode := extractErrorCode(err)

	logger.Error().
		Str("session_id", sessionID.String()).
		Str("error_code", errorCode).
		Str("error", errStr).
		Msg("MTProto error occurred")

	if isTaskError(err) {
		logger.Warn().
			Str("session_id", sessionID.String()).
			Str("error_code", errorCode).
			Str("error", errStr).
			Msg("Task-specific error, marking task as failed")
		return ActionFailTask, 0, getTaskError(err)
	}

	if isBlockingError(err) {
		newState := getBannedState(err)
		logger.Warn().
			Str("session_id", sessionID.String()).
			Str("error_code", errorCode).
			Str("error", errStr).
			Str("new_state", string(newState)).
			Msg("Blocking error, stopping session")

		if err := updateSessionState(ctx, sessionID, newState, repo); err != nil {
			logger.Error().
				Err(err).
				Str("session_id", sessionID.String()).
				Msg("Failed to update session state")
		}

		return ActionStop, 0, domain.ErrSessionNotActive
	}

	if isFloodWaitError(err) {
		seconds := extractFloodWaitSeconds(err)
		if seconds > 0 {
			logger.Warn().
				Str("session_id", sessionID.String()).
				Str("error_code", errorCode).
				Int("wait_seconds", seconds).
				Msg("Flood wait error, pausing")
			return ActionPause, time.Duration(seconds) * time.Second, domain.ErrTelegramFloodWait
		}
	}

	if isSlowmodeError(err) {
		seconds := extractFloodWaitSeconds(err)
		if seconds > 0 {
			logger.Warn().
				Str("session_id", sessionID.String()).
				Str("error_code", errorCode).
				Int("wait_seconds", seconds).
				Msg("Slowmode error, pausing for peer")
			return ActionPause, time.Duration(seconds) * time.Second, domain.ErrTelegramFloodWait
		}
	}

	logger.Error().
		Str("session_id", sessionID.String()).
		Str("error_code", errorCode).
		Str("error", errStr).
		Msg("Unknown MTProto error")

	return ActionContinue, 0, domain.ErrInternal
}

// isBlockingError checks if error is blocking (stops session)
func isBlockingError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	blockingErrors := []string{
		"SESSION_REVOKED",
		"AUTH_KEY_UNREGISTERED",
		"USER_DEACTIVATED_BAN",
		"PHONE_NUMBER_BANNED",
		"USER_DEACTIVATED",
	}

	for _, be := range blockingErrors {
		if strings.Contains(errStr, be) {
			return true
		}
	}

	return false
}

// isTaskError checks if error is task-specific (fails only current task)
func isTaskError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	taskErrors := []string{
		"PEER_ID_INVALID",
		"USERNAME_NOT_OCCUPIED",
		"INPUT_USER_DEACTIVATED",
	}

	for _, te := range taskErrors {
		if strings.Contains(errStr, te) {
			return true
		}
	}

	return false
}

// isFloodWaitError checks if error is flood wait
func isFloodWaitError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "FLOOD_WAIT_")
}

// isSlowmodeError checks if error is slowmode
func isSlowmodeError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "SLOWMODE_WAIT_")
}

// extractFloodWaitSeconds extracts wait time from error message
func extractFloodWaitSeconds(err error) int {
	if err == nil {
		return 0
	}

	re := regexp.MustCompile(`(?:FLOOD_WAIT_|SLOWMODE_WAIT_)(\d+)`)
	matches := re.FindStringSubmatch(err.Error())

	if len(matches) > 1 {
		var seconds int
		_, err := fmt.Sscanf(matches[1], "%d", &seconds)
		if err == nil && seconds > 0 {
			return seconds
		}
	}

	return 0
}

// extractErrorCode extracts error code from error string
func extractErrorCode(err error) string {
	if err == nil {
		return "UNKNOWN"
	}

	errStr := err.Error()

	// Check for known error codes
	errorCodes := []string{
		"SESSION_REVOKED",
		"AUTH_KEY_UNREGISTERED",
		"USER_DEACTIVATED_BAN",
		"PHONE_NUMBER_BANNED",
		"USER_DEACTIVATED",
		"FLOOD_WAIT_",
		"SLOWMODE_WAIT_",
		"PEER_ID_INVALID",
		"USERNAME_NOT_OCCUPIED",
		"INPUT_USER_DEACTIVATED",
	}

	for _, code := range errorCodes {
		if strings.Contains(errStr, code) {
			// For FLOOD_WAIT_X and SLOWMODE_WAIT_X, extract the full code with number
			if strings.Contains(code, "_WAIT_") {
				re := regexp.MustCompile(`(?:FLOOD_WAIT_|SLOWMODE_WAIT_)(\d+)`)
				matches := re.FindStringSubmatch(errStr)
				if len(matches) > 1 {
					return code + matches[1]
				}
			}
			return code
		}
	}

	return "UNKNOWN"
}

// getBannedState returns the banned state for blocking errors
func getBannedState(err error) domain.SessionStatus {
	if err == nil {
		return domain.SessionAuthenticated
	}

	errStr := err.Error()

	if strings.Contains(errStr, "USER_DEACTIVATED") && !strings.Contains(errStr, "BAN") {
		return domain.SessionFrozen
	}

	blockingErrors := []string{
		"SESSION_REVOKED",
		"AUTH_KEY_UNREGISTERED",
		"USER_DEACTIVATED_BAN",
		"PHONE_NUMBER_BANNED",
	}

	for _, be := range blockingErrors {
		if strings.Contains(errStr, be) {
			return domain.SessionBanned
		}
	}

	return domain.SessionAuthenticated
}

// getTaskError returns the appropriate error for task failures
func getTaskError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	if strings.Contains(errStr, "PEER_ID_INVALID") ||
		strings.Contains(errStr, "USERNAME_NOT_OCCUPIED") ||
		strings.Contains(errStr, "INPUT_USER_DEACTIVATED") {
		return domain.ErrPeerNotFound
	}

	return domain.ErrInternal
}

// updateSessionState updates session state in repository
func updateSessionState(
	ctx context.Context,
	sessionID uuid.UUID,
	state domain.SessionStatus,
	repo domain.SessionRepository,
) error {
	sess, err := repo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	sess.AuthState = state
	sess.IsActive = false

	return repo.Update(ctx, sess)
}
