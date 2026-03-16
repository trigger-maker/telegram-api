package handler

import (
	"github.com/gofiber/fiber/v2"
)

// GetStatus retrieves message job status
// @Summary Message status
// @Description Checks status of a sent message
// @Tags Messages
// @Produce json
// @Security BearerAuth
// @Param jobId path string true "Job ID"
// @Success 200 {object} Response{data=domain.MessageJob}
// @Failure 404 {object} Response
// @Router /messages/{jobId}/status [get].
func (h *MessageHandler) GetStatus(c *fiber.Ctx) error {
	jobID := c.Params("jobId")

	job, err := h.service.GetJobStatus(c.Context(), jobID)
	if err != nil {
		return c.Status(404).JSON(NewErrorResponse(404, "Job not found"))
	}

	return c.JSON(NewSuccessResponse(job))
}
