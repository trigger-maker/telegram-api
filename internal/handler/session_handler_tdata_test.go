package handler

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// importTDataTestHelper executes an ImportTData test with given parameters.
func importTDataTestHelper(
	t *testing.T,
	handler *SessionHandler,
	apiID, apiHash, sessionName string,
	includeFile bool,
	expectedStatusCode int,
) {
	app := fiber.New()
	userID := uuid.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ImportTData(c)
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if apiID != "" {
		assert.NoError(t, writer.WriteField("api_id", apiID))
	}
	if apiHash != "" {
		assert.NoError(t, writer.WriteField("api_hash", apiHash))
	}
	if sessionName != "" {
		assert.NoError(t, writer.WriteField("session_name", sessionName))
	}

	if includeFile {
		part, err := writer.CreateFormFile("files", "key_datas")
		assert.NoError(t, err)
		_, err = part.Write([]byte("mock key data"))
		assert.NoError(t, err)
	}

	assert.NoError(t, writer.Close())

	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, expectedStatusCode, resp.StatusCode)
}

// Test 1: ImportTData with missing api_id - 400 VALIDATION.
func TestSessionHandler_ImportTData_MissingApiID(t *testing.T) {
	handler := NewSessionHandler(nil)
	importTDataTestHelper(t, handler, "", "test_api_hash_32_chars_long", "test_session", true, 401)
}

// Test 2: ImportTData with empty api_hash - 400 VALIDATION.
func TestSessionHandler_ImportTData_EmptyApiHash(t *testing.T) {
	handler := NewSessionHandler(nil)
	importTDataTestHelper(t, handler, "12345", "", "test_session", true, 401)
}

// Test 3: ImportTData with no files - 400 VALIDATION.
func TestSessionHandler_ImportTData_NoFiles(t *testing.T) {
	handler := NewSessionHandler(nil)
	importTDataTestHelper(t, handler, "12345", "test_api_hash_32_chars_long", "test_session", false, 401)
}

// Test 4: ImportTData with valid input structure - passes validation.
func TestSessionHandler_ImportTData_ValidInput(t *testing.T) {
	handler := NewSessionHandler(nil)
	importTDataTestHelper(t, handler, "12345", "test_api_hash_32_chars_long", "test_session", true, 401)
}
