package tasks_transport_http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/aaaaarsen/golang-todoapp/internal/core/domain"
	core_errors "github.com/aaaaarsen/golang-todoapp/internal/core/errors"
	core_logger "github.com/aaaaarsen/golang-todoapp/internal/core/logger"
	core_http_response "github.com/aaaaarsen/golang-todoapp/internal/core/transport/http/response"
	"go.uber.org/zap"
)

type PatchTaskRequest struct {
	Title       domain.Nullable[string]   `json:"title"`
	Description domain.Nullable[*string]  `json:"description"`
	Deadline    domain.Nullable[*string]  `json:"deadline"`
	Importance  domain.Nullable[int]      `json:"importance"`
	Category    domain.Nullable[string]   `json:"category"`
	Completed   domain.Nullable[bool]     `json:"completed"`
	Version     int                       `json:"version"`
}

func (h *TasksHTTPHandler) PatchTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	log.Debug("invoke PatchTask handler")

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(rw, "invalid id", http.StatusBadRequest)
		return
	}

	var req PatchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "invalid request body", http.StatusBadRequest)
		return
	}

	patch := domain.TaskPatch{
		Title:      req.Title,
		Importance: req.Importance,
		Category:   req.Category,
		Completed:  req.Completed,
	}

	if req.Description.Valid {
		patch.Description = req.Description
	}

	if req.Deadline.Valid {
		if req.Deadline.Value == nil {
			patch.Deadline = domain.Nullable[*time.Time]{Valid: true, Value: nil}
		} else {
			t, err := time.Parse(time.RFC3339, *req.Deadline.Value)
			if err != nil {
				http.Error(rw, "invalid deadline format, use RFC3339", http.StatusBadRequest)
				return
			}
			utc := t.UTC()
			patch.Deadline = domain.Nullable[*time.Time]{Valid: true, Value: &utc}
		}
	}

	updated, err := h.tasksService.PatchTask(ctx, id, req.Version, patch)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			http.Error(rw, "task not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, core_errors.ErrConflict) {
			http.Error(rw, "version conflict, please refresh and retry", http.StatusConflict)
			return
		}
		if errors.Is(err, core_errors.ErrInvalidArgument) {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		log.Error("patch task error", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	responseHandler.JSONResponse(taskToResponse(updated), http.StatusOK)
}