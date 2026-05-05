package tasks_transport_http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aaaaarsen/golang-todoapp/internal/core/domain"
	core_errors "github.com/aaaaarsen/golang-todoapp/internal/core/errors"
	core_logger "github.com/aaaaarsen/golang-todoapp/internal/core/logger"
	core_http_response "github.com/aaaaarsen/golang-todoapp/internal/core/transport/http/response"
	"go.uber.org/zap"
)

type CreateTaskRequest struct {
	UserID      int     `json:"user_id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Deadline    *string `json:"deadline"`
	Importance  int     `json:"importance"`
	Category    string  `json:"category"`
}

type TaskResponse struct {
	ID          int      `json:"id"`
	Version     int      `json:"version"`
	UserID      int      `json:"user_id"`
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	Deadline    *string  `json:"deadline"`
	Importance  int      `json:"importance"`
	Category    string   `json:"category"`
	Completed   bool     `json:"completed"`
	CompletedAt *string  `json:"completed_at"`

	PriorityScore  *float64 `json:"priority_score"`
	PriorityLevel  *string  `json:"priority_level"`
	PriorityReason *string  `json:"priority_reason"`

	LastUpdatedAt string `json:"last_updated_at"`
	CreatedAt     string `json:"created_at"`
}

func taskToResponse(t domain.Task) TaskResponse {
	resp := TaskResponse{
		ID:             t.ID,
		Version:        t.Version,
		UserID:         t.UserID,
		Title:          t.Title,
		Description:    t.Description,
		Importance:     t.Importance,
		Category:       t.Category,
		Completed:      t.Completed,
		PriorityScore:  t.PriorityScore,
		PriorityLevel:  t.PriorityLevel,
		PriorityReason: t.PriorityReason,
		LastUpdatedAt:  t.LastUpdatedAt.UTC().Format(time.RFC3339),
		CreatedAt:      t.CreatedAt.UTC().Format(time.RFC3339),
	}

	if t.Deadline != nil {
		s := t.Deadline.UTC().Format(time.RFC3339)
		resp.Deadline = &s
	}

	if t.CompletedAt != nil {
		s := t.CompletedAt.UTC().Format(time.RFC3339)
		resp.CompletedAt = &s
	}

	return resp
}

func (h *TasksHTTPHandler) CreateTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	log.Debug("invoke CreateTask handler")

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Importance == 0 {
		req.Importance = 1
	}
	if req.Category == "" {
		req.Category = "personal"
	}

	var deadline *time.Time
	if req.Deadline != nil {
		t, err := time.Parse(time.RFC3339, *req.Deadline)
		if err != nil {
			http.Error(rw, "invalid deadline format, use RFC3339", http.StatusBadRequest)
			return
		}
		utc := t.UTC()
		deadline = &utc
	}

	task := domain.NewTask(
		req.UserID,
		req.Title,
		req.Description,
		deadline,
		req.Importance,
		req.Category,
	)

	created, err := h.tasksService.CreateTask(ctx, task)
	if err != nil {
		if errors.Is(err, core_errors.ErrInvalidArgument) {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		log.Error("create task error", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	responseHandler.JSONResponse(taskToResponse(created), http.StatusCreated)
}