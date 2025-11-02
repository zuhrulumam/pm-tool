package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"github.com/zuhrulumam/pm-tool/business/usecase"
	"github.com/zuhrulumam/pm-tool/handler/api/request"
	"github.com/zuhrulumam/pm-tool/handler/api/response"
)

// NoteHandler handles HTTP requests for Note
type NoteHandler struct {
	usecase usecase.NoteUsecaseItf
	tracer  trace.Tracer
}

// NewNoteHandler creates a new NoteHandler instance
func NewNoteHandler(usecase usecase.NoteUsecaseItf, tracer trace.Tracer) *NoteHandler {
	return &NoteHandler{
		usecase: usecase,
		tracer:  tracer,
	}
}

// parseID parses and validates ID from URL parameter
func (h *NoteHandler) parseID(c *gin.Context) (string, error) {
	idStr := c.Param("id")
	if idStr == "" {
		return "", errors.New("id is required")
	}

	id := idStr

	return id, nil
}

// CreateNote godoc
// @Summary Create new note
// @Description Create a new note
// @Tags notes
// @Accept json
// @Produce json
// @Param request body request.CreateNoteRequest true "Note data"
// @Success 201 {object} response.NoteResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notes [post]
// @Security Bearer
func (h *NoteHandler) CreateNote(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "handler.CreateNote")
	defer span.End()

	var req request.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	note := req.ToEntity()

	if err := h.usecase.CreateNote(ctx, note); err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.NewNoteResponse(note))
}

// GetNote godoc
// @Summary Get note by ID
// @Description Get a specific note by ID
// @Tags notes
// @Accept json
// @Produce json
// @Param id path string true "Note ID"
// @Success 200 {object} response.NoteResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notes/{id} [get]
// @Security Bearer
func (h *NoteHandler) GetNote(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "handler.GetNote")
	defer span.End()

	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_ID",
		})
		return
	}

	note, err := h.usecase.GetNote(ctx, id)
	if err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.NewNoteResponse(note))
}

// UpdateNote godoc
// @Summary Update note
// @Description Update an existing note
// @Tags notes
// @Accept json
// @Produce json
// @Param id path string true "Note ID"
// @Param request body request.UpdateNoteRequest true "Note data"
// @Success 200 {object} response.NoteResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notes/{id} [put]
// @Security Bearer
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "handler.UpdateNote")
	defer span.End()

	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_ID",
		})
		return
	}

	var req request.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	note, err := h.usecase.GetNote(ctx, id)
	if err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	req.ApplyTo(note)

	if err := h.usecase.UpdateNote(ctx, id, note); err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.NewNoteResponse(note))
}

// DeleteNote godoc
// @Summary Delete note
// @Description Delete an existing note
// @Tags notes
// @Accept json
// @Produce json
// @Param id path string true "Note ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notes/{id} [delete]
// @Security Bearer
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "handler.DeleteNote")
	defer span.End()

	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_ID",
		})
		return
	}

	if err := h.usecase.DeleteNote(ctx, id); err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListNotes godoc
// @Summary List notes
// @Description Get a paginated list of notes
// @Tags notes
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} response.NoteListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notes [get]
// @Security Bearer
func (h *NoteHandler) ListNotes(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "handler.ListNotes")
	defer span.End()

	// Parse query parameters into filter request
	var req request.ListNoteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "Invalid request query param",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	// Get pagination
	page, pageSize := req.GetPagination()

	// Convert to filters map
	filters := req.ToFilters()

	items, total, err := h.usecase.ListNotes(ctx, filters, page, pageSize)
	if err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.NewNoteListResponse(items, total, page, pageSize))
}

// CountNotes godoc
// @Summary Count Notes with filters
// @Description Count Notes with optional filters
// @Tags Note
// @Accept json
// @Produce json
// @Param filters query request.ListNoteRequest false "Filters"
// @Success 200 {object} response.SuccessResponse{data=int64}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router //api/v1/notess/count [get]
func (h *NoteHandler) CountNotes(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "handler.Note.Count")
	defer span.End()

	// Parse query parameters into filter request
	var req request.ListNoteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "Invalid request query param",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	// Convert to filters map
	filters := req.ToFilters()

	// Call usecase
	count, err := h.usecase.CountNotes(ctx, filters)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: count,
	})
}
