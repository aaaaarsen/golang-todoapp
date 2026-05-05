package tasks_transport_http

import (
	"net/http"
	"strconv"

	core_logger "github.com/aaaaarsen/golang-todoapp/internal/core/logger"
	core_http_response "github.com/aaaaarsen/golang-todoapp/internal/core/transport/http/response"
	"go.uber.org/zap"
)

type GetTasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
	Total int            `json:"total"`
}

func (h *TasksHTTPHandler) GetTasks(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	log.Debug("invoke GetTasks handler")

	q := r.URL.Query()

	var userID *int
	if v := q.Get("user_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			http.Error(rw, "invalid user_id", http.StatusBadRequest)
			return
		}
		userID = &id
	}

	var category *string
	if v := q.Get("category"); v != "" {
		category = &v
	}

	var completed *bool
	if v := q.Get("completed"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			http.Error(rw, "invalid completed value", http.StatusBadRequest)
			return
		}
		completed = &b
	}

	page := 1
	if v := q.Get("page"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 {
			http.Error(rw, "invalid page", http.StatusBadRequest)
			return
		}
		page = p
	}

	pageSize := 10
	if v := q.Get("page_size"); v != "" {
		ps, err := strconv.Atoi(v)
		if err != nil || ps < 1 {
			http.Error(rw, "invalid page_size", http.StatusBadRequest)
			return
		}
		pageSize = ps
	}

	tasks, total, err := h.tasksService.GetTasks(ctx, userID, category, completed, page, pageSize)
	if err != nil {
		log.Error("get tasks error", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := GetTasksResponse{
		Tasks: make([]TaskResponse, 0, len(tasks)),
		Total: total,
	}
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, taskToResponse(t))
	}

	responseHandler.JSONResponse(resp, http.StatusOK)
}