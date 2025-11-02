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

// UserHandler handles HTTP requests for User
type UserHandler struct {
	usecase usecase.UserUsecaseItf
	tracer trace.Tracer
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(usecase usecase.UserUsecaseItf, tracer trace.Tracer) *UserHandler {
	return &UserHandler{
		usecase: usecase,
		tracer: tracer,
	}
}

// parseID parses and validates ID from URL parameter
func (h *UserHandler) parseID(c *gin.Context) (string, error) {
	idStr := c.Param("id")
	if idStr == "" {
		return "", errors.New("id is required")
	}

	id := idStr
	err := error(nil)
	
	if err != nil {
		return "", errors.New("invalid id format")
	}
	
	return id, nil
}

// CreateUser godoc
// @Summary Create new user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param request body request.CreateUserRequest true "User data"
// @Success 201 {object} response.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users [post]
// @Security Bearer
func (h *UserHandler) CreateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	
	ctx, span := h.tracer.Start(ctx, "handler.CreateUser")
	defer span.End()
	

	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	user := req.ToEntity()

	if err := h.usecase.CreateUser(ctx, user); err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.NewUserResponse(user))
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get a specific user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [get]
// @Security Bearer
func (h *UserHandler) GetUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	ctx, span := h.tracer.Start(ctx, "handler.GetUser")
	defer span.End()
	

	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_ID",
		})
		return
	}

	user, err := h.usecase.GetUser(ctx, id)
	if err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.NewUserResponse(user))
}

// UpdateUser godoc
// @Summary Update user
// @Description Update an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body request.UpdateUserRequest true "User data"
// @Success 200 {object} response.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [put]
// @Security Bearer
func (h *UserHandler) UpdateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	
	ctx, span := h.tracer.Start(ctx, "handler.UpdateUser")
	defer span.End()
	

	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_ID",
		})
		return
	}

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	user, err := h.usecase.GetUser(ctx, id)
	if err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	req.ApplyTo(user)

	if err := h.usecase.UpdateUser(ctx, id, user); err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.NewUserResponse(user))
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [delete]
// @Security Bearer
func (h *UserHandler) DeleteUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	
	ctx, span := h.tracer.Start(ctx, "handler.DeleteUser")
	defer span.End()
	

	id, err := h.parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_ID",
		})
		return
	}

	if err := h.usecase.DeleteUser(ctx, id); err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListUsers godoc
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} response.UserListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users [get]
// @Security Bearer
func (h *UserHandler) ListUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	
	ctx, span := h.tracer.Start(ctx, "handler.ListUsers")
	defer span.End()
	

	// Parse query parameters into filter request
	var req request.ListUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request query param",
			Code:  "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}
	
	// Get pagination
	page, pageSize := req.GetPagination()
	
	// Convert to filters map
	filters := req.ToFilters()

	items, total, err := h.usecase.ListUsers(ctx, filters, page, pageSize)
	if err != nil {
		span.RecordError(err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.NewUserListResponse(items, total, page, pageSize))
}

// CountUsers godoc
// @Summary Count Users with filters
// @Description Count Users with optional filters
// @Tags User
// @Accept json
// @Produce json
// @Param filters query request.ListUserRequest false "Filters"
// @Success 200 {object} response.SuccessResponse{data=int64}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router //api/v1/userss/count [get]
func (h *UserHandler) CountUsers(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "handler.User.Count")
	defer span.End()
	
	
	// Parse query parameters into filter request
	var req request.ListUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request query param",
			Code:  "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}
	
	// Convert to filters map
	filters := req.ToFilters()
	
	// Call usecase
	count, err := h.usecase.CountUsers(ctx, filters)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	
	c.JSON(http.StatusOK, response.Response{
		Data: count,
	})
}
