package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zuhrulumam/pm-tool/business/usecase"
	"github.com/zuhrulumam/pm-tool/config"
	"github.com/zuhrulumam/pm-tool/handler/api/response"
	"github.com/zuhrulumam/pm-tool/pkg/middleware"
	"go.opentelemetry.io/otel/trace"
)

// Handlers holds all HTTP handler instances
type Handlers struct {
	User *UserHandler
	Project *ProjectHandler
	Note *NoteHandler

Config *config.Config
}

// NewHandlers initializes all handlers from usecases
func NewHandlers(usecases *usecase.Usecases, conf *config.Config, tracer trace.Tracer) *Handlers {
	return &Handlers{
		User: NewUserHandler(usecases.User, tracer),
		Project: NewProjectHandler(usecases.Project, tracer),
		Note: NewNoteHandler(usecases.Note, tracer),

	Config: conf,	}
}

// SetupRoutes configures all HTTP routes with middleware
func (h *Handlers) SetupRoutes(engine *gin.Engine) {

	// Health check endpoint (no auth required)
	engine.GET("/health", h.healthCheck)

	// API v1 group with auth
	v1 := engine.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(h.Config.JWT.Secret)) // Apply auth to all API routes
	

	// User routes
	userRoutes := v1.Group("/users")
	{
		userRoutes.POST("", h.User.CreateUser)
		userRoutes.GET("/:id", h.User.GetUser)
		userRoutes.PUT("/:id", h.User.UpdateUser)
		userRoutes.DELETE("/:id", h.User.DeleteUser)
		userRoutes.GET("", h.User.ListUsers)
	}

	// Project routes
	projectRoutes := v1.Group("/projects")
	{
		projectRoutes.POST("", h.Project.CreateProject)
		projectRoutes.GET("/:id", h.Project.GetProject)
		projectRoutes.PUT("/:id", h.Project.UpdateProject)
		projectRoutes.DELETE("/:id", h.Project.DeleteProject)
		projectRoutes.GET("", h.Project.ListProjects)
	}

	// Note routes
	noteRoutes := v1.Group("/notes")
	{
		noteRoutes.POST("", h.Note.CreateNote)
		noteRoutes.GET("/:id", h.Note.GetNote)
		noteRoutes.PUT("/:id", h.Note.UpdateNote)
		noteRoutes.DELETE("/:id", h.Note.DeleteNote)
		noteRoutes.GET("", h.Note.ListNotes)
	}

}

// healthCheck handles health check requests
func (h *Handlers) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, response.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0", // TODO: Get from config or build info
	})
}
