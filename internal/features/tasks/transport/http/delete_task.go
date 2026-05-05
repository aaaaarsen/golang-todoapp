package tasks_transport_http

import (
	"errors"
	"net/http"
	"strconv"

	core_errors "github.com/aaaaarsen/golang-todoapp/internal/core/errors"
	core_logger "github.com/aaaaarsen/golang-todoapp/internal/core/logger"
	"go.uber.org/zap"
)

func (h *TasksHTTPHandler) DeleteTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)

	log.Debug("invoke DeleteTask handler")

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(rw, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.tasksService.DeleteTask(ctx, id); err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			http.Error(rw, "task not found", http.StatusNotFound)
			return
		}
		log.Error("delete task error", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}