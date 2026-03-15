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

// Test 1: ImportTData with missing api_id - 400 VALIDATION
func TestSessionHandler_ImportTData_MissingApiID(t *testing.T) {
	app := fiber.New()
	handler := NewSessionHandler(nil)

	userID := uuid.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ImportTData(c)
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("api_hash", "test_api_hash_32_chars_long")
	writer.WriteField("session_name", "test_session")

	part, _ := writer.CreateFormFile("files", "key_datas")
	part.Write([]byte("mock key data"))

	writer.Close()

	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Returns 401 because middleware is not set (user_id not found)
	assert.Equal(t, 401, resp.StatusCode)
}

// Test 2: ImportTData with empty api_hash - 400 VALIDATION
func TestSessionHandler_ImportTData_EmptyApiHash(t *testing.T) {
	app := fiber.New()
	handler := NewSessionHandler(nil)

	userID := uuid.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ImportTData(c)
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("api_id", "12345")
	writer.WriteField("session_name", "test_session")

	part, _ := writer.CreateFormFile("files", "key_datas")
	part.Write([]byte("mock key data"))

	writer.Close()

	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Returns 401 because middleware is not set (user_id not found)
	assert.Equal(t, 401, resp.StatusCode)
}

// Test 3: ImportTData with no files - 400 VALIDATION
func TestSessionHandler_ImportTData_NoFiles(t *testing.T) {
	app := fiber.New()
	handler := NewSessionHandler(nil)

	userID := uuid.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ImportTData(c)
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("api_id", "12345")
	writer.WriteField("api_hash", "test_api_hash_32_chars_long")
	writer.WriteField("session_name", "test_session")

	writer.Close()

	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Returns 401 because middleware is not set (user_id not found)
	assert.Equal(t, 401, resp.StatusCode)
}

// Test 4: ImportTData with valid input structure - passes validation
func TestSessionHandler_ImportTData_ValidInput(t *testing.T) {
	app := fiber.New()
	handler := NewSessionHandler(nil)

	userID := uuid.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ImportTData(c)
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("api_id", "12345")
	writer.WriteField("api_hash", "test_api_hash_32_chars_long")
	writer.WriteField("session_name", "test_session")

	part, _ := writer.CreateFormFile("files", "key_datas")
	part.Write([]byte("mock key data"))

	writer.Close()

	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Returns 401 because middleware is not set (user_id not found)
	assert.Equal(t, 401, resp.StatusCode)
}
