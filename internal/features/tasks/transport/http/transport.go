package tasks_transport_http

import (
	"context"
	"net/http"

	"github.com/aaaaarsen/golang-todoapp/internal/core/domain"
	core_http_server "github.com/aaaaarsen/golang-todoapp/internal/core/transport/http/server"
)

type TasksService interface {
	CreateTask(ctx context.Context, task domain.Task) (domain.Task, error)
	GetTasks(ctx context.Context, userID *int, category *string, completed *bool, page int, pageSize int) ([]domain.Task, int, error)
	GetTask(ctx context.Context, id int) (domain.Task, error)
	DeleteTask(ctx context.Context, id int) error
	PatchTask(ctx context.Context, id int, version int, patch domain.TaskPatch) (domain.Task, error)
	GetStagnantTasks(ctx context.Context, userID *int) ([]domain.Task, error)
}

type TasksHTTPHandler struct {
	tasksService TasksService
}

func NewTasksHTTPHandler(tasksService TasksService) *TasksHTTPHandler {
	return &TasksHTTPHandler{
		tasksService: tasksService,
	}
}

func (h *TasksHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{Method: http.MethodPost, Path: "/tasks", Handler: h.CreateTask},
		{Method: http.MethodGet, Path: "/tasks", Handler: h.GetTasks},
		{Method: http.MethodGet, Path: "/tasks/stagnant", Handler: h.GetStagnantTasks},
		{Method: http.MethodGet, Path: "/tasks/{id}", Handler: h.GetTask},
		{Method: http.MethodDelete, Path: "/tasks/{id}", Handler: h.DeleteTask},
		{Method: http.MethodPatch, Path: "/tasks/{id}", Handler: h.PatchTask},
	}
}