package tasks_transport_http

import (
	"errors"
	"net/http"
	"strconv"

	core_errors "github.com/aaaaarsen/golang-todoapp/internal/core/errors"
	core_logger "github.com/aaaaarsen/golang-todoapp/internal/core/logger"
	core_http_response "github.com/aaaaarsen/golang-todoapp/internal/core/transport/http/response"
	"go.uber.org/zap"
)

func (h *TasksHTTPHandler) GetTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	log.Debug("invoke GetTask handler")

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(rw, "invalid id", http.StatusBadRequest)
		return
	}

	task, err := h.tasksService.GetTask(ctx, id)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			http.Error(rw, "task not found", http.StatusNotFound)
			return
		}
		log.Error("get task error", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	responseHandler.JSONResponse(taskToResponse(task), http.StatusOK)
}