package handler

import (
	"io"
	"strconv"

	"telegram-api/internal/middleware"
	"telegram-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// ImportTData imports Telegram Desktop session
// @Summary Import Telegram Desktop session
// @Description Import Telegram Desktop session from tdata files
// @Tags Sessions
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param api_id formData int true "API ID"
// @Param api_hash formData string true "API Hash"
// @Param session_name formData string false "Session Name"
// @Param files formData file true "TData files"
// @Success 201 {object} handler.Response{data=fiber.Map}
// @Failure 400 {object} handler.Response
// @Failure 422 {object} handler.Response
// @Router /sessions/import-tdata [post]
func (h *SessionHandler) ImportTData(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(NewErrorResponse(401, "Not authenticated"))
	}

	form, err := c.MultipartForm()
	if err != nil {
		logger.Error().Err(err).Msg("Error parsing multipart form")
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid multipart form"))
	}

	apiID, err := strconv.Atoi(c.FormValue("api_id", "0"))
	if err != nil || apiID <= 0 {
		return c.Status(400).JSON(NewErrorResponse(400, "api_id is required and must be positive"))
	}

	apiHash := c.FormValue("api_hash", "")
	if apiHash == "" {
		return c.Status(400).JSON(NewErrorResponse(400, "api_hash is required"))
	}

	sessionName := c.FormValue("session_name", "")

	tdataFiles := make(map[string][]byte)
	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(400).JSON(NewErrorResponse(400, "tdata files are required"))
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			logger.Error().Err(err).Str("filename", fileHeader.Filename).Msg("Error opening tdata file")
			return c.Status(400).JSON(NewErrorResponse(400, "Error reading tdata file"))
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			logger.Error().Err(err).Str("filename", fileHeader.Filename).Msg("Error reading tdata file content")
			return c.Status(400).JSON(NewErrorResponse(400, "Error reading tdata file content"))
		}

		tdataFiles[fileHeader.Filename] = content
	}

	logger.Debug().
		Str("user_id", userID.String()).
		Int("api_id", apiID).
		Str("session_name", sessionName).
		Int("files_count", len(tdataFiles)).
		Msg("Processing tdata import request...")

	session, err := h.service.ImportTData(c.Context(), userID, apiID, apiHash, sessionName, tdataFiles)
	if err != nil {
		return handleSessionError(c, err)
	}

	response := fiber.Map{
		"session": fiber.Map{
			"session_id":       session.ID,
			"is_active":        session.IsActive,
			"telegram_user_id": session.TelegramUserID,
			"username":         session.TelegramUsername,
			"auth_state":       session.AuthState,
			"auth_method":      "tdata",
		},
	}

	return c.Status(201).JSON(NewSuccessResponse(response))
}
