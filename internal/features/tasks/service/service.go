package tasks_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/aaaaarsen/golang-todoapp/internal/core/domain"
	core_errors "github.com/aaaaarsen/golang-todoapp/internal/core/errors"
	tasks_priority "github.com/aaaaarsen/golang-todoapp/internal/features/tasks/priority"
)

type TasksRepository interface {
	CreateTask(ctx context.Context, task domain.Task, priorityScore float64, priorityLevel string, priorityReason string) (domain.Task, error)
	GetTasks(ctx context.Context, userID *int, category *string, completed *bool, page int, pageSize int) ([]domain.Task, int, error)
	GetTask(ctx context.Context, id int) (domain.Task, error)
	DeleteTask(ctx context.Context, id int) error
	PatchTask(ctx context.Context, id int, version int, patch domain.TaskPatch, priorityScore float64, priorityLevel string, priorityReason string) (domain.Task, error)
	GetStagnantTasks(ctx context.Context, userID *int) ([]domain.Task, error)
}

type TaskService struct {
	tasksRepository TasksRepository
	priorityEngine  *tasks_priority.PriorityEngine
}

func NewTaskService(
	tasksRepository TasksRepository,
	priorityEngine *tasks_priority.PriorityEngine,
) *TaskService {
	return &TaskService{
		tasksRepository: tasksRepository,
		priorityEngine:  priorityEngine,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, task domain.Task) (domain.Task, error) {
	if err := task.Validate(); err != nil {
		return domain.Task{}, fmt.Errorf("validate task: %w", err)
	}

	result := s.priorityEngine.Calculate(task)

	created, err := s.tasksRepository.CreateTask(ctx, task, result.Score, result.Level, result.Reason)
	if err != nil {
		return domain.Task{}, fmt.Errorf("create task: %w", err)
	}

	return created, nil
}

func (s *TaskService) GetTasks(ctx context.Context, userID *int, category *string, completed *bool, page int, pageSize int) ([]domain.Task, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	tasks, total, err := s.tasksRepository.GetTasks(ctx, userID, category, completed, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("get tasks: %w", err)
	}

	return tasks, total, nil
}

func (s *TaskService) GetTask(ctx context.Context, id int) (domain.Task, error) {
	task, err := s.tasksRepository.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}

	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int) error {
	if err := s.tasksRepository.DeleteTask(ctx, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (s *TaskService) PatchTask(ctx context.Context, id int, version int, patch domain.TaskPatch) (domain.Task, error) {
	existing, err := s.tasksRepository.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task for patch: %w", err)
	}

	if patch.Title.Valid {
		existing.Title = patch.Title.Value
	}
	if patch.Importance.Valid {
		existing.Importance = patch.Importance.Value
	}
	if patch.Category.Valid {
		existing.Category = patch.Category.Value
	}
	if patch.Deadline.Valid {
		existing.Deadline = patch.Deadline.Value
	}

	if err := existing.Validate(); err != nil {
		return domain.Task{}, fmt.Errorf("validate task patch: %w", err)
	}

	result := s.priorityEngine.Calculate(existing)

	updated, err := s.tasksRepository.PatchTask(ctx, id, version, patch, result.Score, result.Level, result.Reason)
	if err != nil {
		if errors.Is(err, core_errors.ErrConflict) {
			return domain.Task{}, core_errors.ErrConflict
		}
		return domain.Task{}, fmt.Errorf("patch task: %w", err)
	}

	return updated, nil
}

func (s *TaskService) GetStagnantTasks(ctx context.Context, userID *int) ([]domain.Task, error) {
	tasks, err := s.tasksRepository.GetStagnantTasks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get stagnant tasks: %w", err)
	}

	return tasks, nil
}