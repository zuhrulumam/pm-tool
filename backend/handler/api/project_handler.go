package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"github.com/zuhrulumam/pm-tool/business/usecase"
	"github.com/zuhrulumam/pm-tool/handler/api/request"
	"github.com/zuhrulumam/pm-tool/handler/api/response"
	contexthelper "github.com/zuhrulumam/pm-tool/pkg/context"
)

// ProjectHandler handles HTTP requests for Project
type ProjectHandler struct {
	usecase usecase.ProjectUsecaseItf
	tracer  trace.Tracer
}

// NewProjectHandler creates a new ProjectHandler instance
func NewProjectHandler(usecase usecase.ProjectUsecaseItf, tracer trace.Tracer) *ProjectHandler {
	return &ProjectHandler{
		usecase: usecase,
		tracer:  tracer,
	}
}

// parseID parses and validates ID from URL parameter
func (h *ProjectHandler) parseID(c *gin.Context) (string, error) {
	idStr := c.Param("id")
	if idStr == "" {
		return "", errors.New("id is required")
	}

	id := idStr

	return id, nil
}

// CreateProject godoc
// @Summary Create new project
// @Description Create a new project
// @Tags projects
// @Accept json
// @Produce json
// @Param request body request.CreateProjectRequest true "Project data"
// @Success 201 {object} response.ProjectResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/projects [post]
// @Security Bearer
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "handler.CreateProject")
	defer span.End()

	var req request.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	project := req.ToEntity()
	project.UserId = contexthelper.MustGetUserID(ctx)

	if err := h.usecase.CreateProject(ctx, project); err != nil {
		fmt.Println("sini", err)
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.NewProjectResponse(project))
}

// GetProject godoc
// @Summary Get project by ID
// @Description Get a specific project by ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} response.ProjectResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/projects/{id} [get]
// @Security Bearer
func (h *ProjectHandler) GetProject(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "handler.GetProject")
	defer span.End()

	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_ID",
		})
		return
	}

	project, err := h.usecase.GetProject(ctx, id)
	if err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.NewProjectResponse(project))
}

// UpdateProject godoc
// @Summary Update project
// @Description Update an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body request.UpdateProjectRequest true "Project data"
// @Success 200 {object} response.ProjectResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/projects/{id} [put]
// @Security Bearer
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "handler.UpdateProject")
	defer span.End()

	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_ID",
		})
		return
	}

	var req request.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	project, err := h.usecase.GetProject(ctx, id)
	if err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	req.ApplyTo(project)

	if err := h.usecase.UpdateProject(ctx, id, project); err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.NewProjectResponse(project))
}

// DeleteProject godoc
// @Summary Delete project
// @Description Delete an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/projects/{id} [delete]
// @Security Bearer
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "handler.DeleteProject")
	defer span.End()

	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_ID",
		})
		return
	}

	if err := h.usecase.DeleteProject(ctx, id); err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListProjects godoc
// @Summary List projects
// @Description Get a paginated list of projects
// @Tags projects
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} response.ProjectListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/projects [get]
// @Security Bearer
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	ctx, span := h.tracer.Start(ctx, "handler.ListProjects")
	defer span.End()

	// Parse query parameters into filter request
	var req request.ListProjectRequest
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

	items, total, err := h.usecase.ListProjects(ctx, filters, page, pageSize)
	if err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.NewProjectListResponse(items, total, page, pageSize))
}

// CountProjects godoc
// @Summary Count Projects with filters
// @Description Count Projects with optional filters
// @Tags Project
// @Accept json
// @Produce json
// @Param filters query request.ListProjectRequest false "Filters"
// @Success 200 {object} response.SuccessResponse{data=int64}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router //api/v1/projectss/count [get]
func (h *ProjectHandler) CountProjects(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "handler.Project.Count")
	defer span.End()

	// Parse query parameters into filter request
	var req request.ListProjectRequest
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
	count, err := h.usecase.CountProjects(ctx, filters)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: count,
	})
}
