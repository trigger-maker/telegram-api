package handler

import (
	"github.com/go-playground/validator/v10"
)

// Response represents the standard API response
type Response struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
	Meta    *MetaResponse  `json:"meta,omitempty"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// MetaResponse represents pagination metadata
type MetaResponse struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// ValidationError represents an individual validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Global validator
var validate = validator.New()

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) []ValidationError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errors []ValidationError
	for _, e := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   e.Field(),
			Message: getValidationMessage(e),
		})
	}
	return errors
}

// getValidationMessage returns a readable message for each validation type
func getValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email"
	case "min":
		return "Must be at least " + e.Param() + " characters"
	case "max":
		return "Must not exceed " + e.Param() + " characters"
	case "alphanum":
		return "Must contain only letters and numbers"
	case "uuid":
		return "Must be a valid UUID"
	case "e164":
		return "Must be a valid phone number (E.164 format)"
	case "len":
		return "Must be exactly " + e.Param() + " characters"
	default:
		return "Invalid value"
	}
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(code int, message string) Response {
	return Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
	}
}

// NewPaginatedResponse creates a paginated response
func NewPaginatedResponse(data interface{}, page, perPage int, total int64) Response {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return Response{
		Success: true,
		Data:    data,
		Meta: &MetaResponse{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
