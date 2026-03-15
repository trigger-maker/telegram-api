package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// InvalidateCache invalidates cache for specified type
func (s *ChatService) InvalidateCache(ctx context.Context, sessionID uuid.UUID, cacheType string) error {
	var keys []string

	switch cacheType {
	case "contacts":
		keys = []string{fmt.Sprintf("tg:contacts:%s", sessionID.String())}
	case "chats":
		keys = []string{
			fmt.Sprintf("tg:chats:%s:archived_true", sessionID.String()),
			fmt.Sprintf("tg:chats:%s:archived_false", sessionID.String()),
		}
	case "all":
		keys = []string{
			fmt.Sprintf("tg:contacts:%s", sessionID.String()),
			fmt.Sprintf("tg:chats:%s:archived_true", sessionID.String()),
			fmt.Sprintf("tg:chats:%s:archived_false", sessionID.String()),
		}
		pattern := fmt.Sprintf("tg:chat:%s:*", sessionID.String())
		if scanned, err := s.cacheRepo.ScanKeys(ctx, pattern, 100); err == nil {
			keys = append(keys, scanned...)
		}
		pattern = fmt.Sprintf("tg:resolve:%s:*", sessionID.String())
		if scanned, err := s.cacheRepo.ScanKeys(ctx, pattern, 100); err == nil {
			keys = append(keys, scanned...)
		}
	default:
		return fmt.Errorf("invalid cache type: %s", cacheType)
	}

	if len(keys) > 0 {
		return s.cacheRepo.Delete(ctx, keys...)
	}
	return nil
}
