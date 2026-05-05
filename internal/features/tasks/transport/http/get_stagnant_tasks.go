package tasks_transport_http

import (
	"net/http"
	"strconv"

	core_logger "github.com/aaaaarsen/golang-todoapp/internal/core/logger"
	core_http_response "github.com/aaaaarsen/golang-todoapp/internal/core/transport/http/response"
	"go.uber.org/zap"
)

type GetStagnantTasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
	Total int            `json:"total"`
}

func (h *TasksHTTPHandler) GetStagnantTasks(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	log.Debug("invoke GetStagnantTasks handler")

	var userID *int
	if v := r.URL.Query().Get("user_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			http.Error(rw, "invalid user_id", http.StatusBadRequest)
			return
		}
		userID = &id
	}

	tasks, err := h.tasksService.GetStagnantTasks(ctx, userID)
	if err != nil {
		log.Error("get stagnant tasks error", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := GetStagnantTasksResponse{
		Tasks: make([]TaskResponse, 0, len(tasks)),
		Total: len(tasks),
	}
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, taskToResponse(t))
	}

	responseHandler.JSONResponse(resp, http.StatusOK)
}