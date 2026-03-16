package handler

import (
	"strconv"

	"telegram-api/internal/domain"
	"telegram-api/internal/middleware"
	"telegram-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetChats retrieves chats for a session
// @Summary List chats
// @Description Retrieves the list of chats/dialogs for the session (with Redis cache)
// @Tags Chats
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param limit query int false "Result limit (default 50, max 100)"
// @Param offset query int false "Pagination offset"
// @Param archived query bool false "Include archived chats"
// @Param refresh query bool false "Force cache refresh"
// @Success 200 {object} Response{data=domain.ChatsResponse}
// @Router /sessions/{id}/chats [get].
func (h *ChatHandler) GetChats(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid session ID"))
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(NewErrorResponse(401, "Unauthorized"))
	}

	req := domain.GetChatsRequest{
		Limit:    c.QueryInt("limit", 50),
		Offset:   c.QueryInt("offset", 0),
		Archived: c.QueryBool("archived", false),
		Refresh:  c.QueryBool("refresh", false),
	}

	logger.Info().
		Str("session_id", sessionID.String()).
		Int("limit", req.Limit).
		Int("offset", req.Offset).
		Bool("refresh", req.Refresh).
		Msg("GET chats")

	result, err := h.chatService.GetDialogs(c.Context(), userID, sessionID, req)
	if err != nil {
		logger.Error().Err(err).Msg("error getting chats")
		return h.handleError(c, err)
	}

	logger.Info().
		Int("returned", len(result.Chats)).
		Int("total", result.TotalCount).
		Bool("from_cache", result.FromCache).
		Msg("chats retrieved")

	return c.JSON(NewSuccessResponse(result))
}

// GetChatInfo retrieves chat information
// @Summary Get chat information
// @Description Retrieves detailed information about a specific chat (with cache)
// @Tags Chats
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param chatId path int true "Chat ID"
// @Success 200 {object} Response{data=domain.Chat}
// @Router /sessions/{id}/chats/{chatId} [get].
func (h *ChatHandler) GetChatInfo(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid session ID"))
	}

	chatID, err := strconv.ParseInt(c.Params("chatId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid chat ID"))
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(NewErrorResponse(401, "Unauthorized"))
	}

	logger.Info().
		Str("session_id", sessionID.String()).
		Int64("chat_id", chatID).
		Msg("GET chat info")

	result, err := h.chatService.GetChatInfo(c.Context(), userID, sessionID, chatID)
	if err != nil {
		logger.Error().Err(err).Int64("chat_id", chatID).Msg("error getting chat info")
		return h.handleError(c, err)
	}

	return c.JSON(NewSuccessResponse(result))
}

// GetChatHistory retrieves chat message history
// @Summary Get message history
// @Description Retrieves the message history for a chat (no cache - real time)
// @Tags Chats
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param chatId path int true "Chat ID"
// @Param limit query int false "Message limit (default 50, max 100)"
// @Param offset_id query int false "Message ID to start from"
// @Param offset_date query int false "Unix timestamp to start from"
// @Success 200 {object} Response{data=domain.HistoryResponse}
// @Router /sessions/{id}/chats/{chatId}/history [get].
func (h *ChatHandler) GetChatHistory(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid session ID"))
	}

	chatID, err := strconv.ParseInt(c.Params("chatId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid chat ID"))
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(NewErrorResponse(401, "Unauthorized"))
	}

	req := domain.GetHistoryRequest{
		Limit:      c.QueryInt("limit", 50),
		OffsetID:   c.QueryInt("offset_id", 0),
		OffsetDate: c.QueryInt("offset_date", 0),
	}

	logger.Info().
		Str("session_id", sessionID.String()).
		Int64("chat_id", chatID).
		Int("limit", req.Limit).
		Msg("GET chat history")

	result, err := h.chatService.GetChatHistory(c.Context(), userID, sessionID, chatID, req)
	if err != nil {
		logger.Error().Err(err).Int64("chat_id", chatID).Msg("error getting history")
		return h.handleError(c, err)
	}

	logger.Info().Int("messages", result.TotalCount).Msg("history retrieved")
	return c.JSON(NewSuccessResponse(result))
}
