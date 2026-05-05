package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aaaaarsen/golang-todoapp/internal/core/domain"
	core_errors "github.com/aaaaarsen/golang-todoapp/internal/core/errors"
	core_postgres_pool "github.com/aaaaarsen/golang-todoapp/internal/core/repository/postgres/conn"
)

type TasksRepository struct {
	pool core_postgres_pool.Pool
}

func NewTasksRepository(pool core_postgres_pool.Pool) *TasksRepository {
	return &TasksRepository{pool: pool}
}

type taskModel struct {
	ID          int
	Version     int
	UserID      int
	Title       string
	Description *string
	Deadline    *time.Time
	Importance  int
	Category    string
	Completed   bool
	CompletedAt *time.Time

	PriorityScore  *float64
	PriorityLevel  *string
	PriorityReason *string

	LastUpdatedAt time.Time
	CreatedAt     time.Time
}

func (m taskModel) toDomain() domain.Task {
	return domain.Task{
		ID:             m.ID,
		Version:        m.Version,
		UserID:         m.UserID,
		Title:          m.Title,
		Description:    m.Description,
		Deadline:       m.Deadline,
		Importance:     m.Importance,
		Category:       m.Category,
		Completed:      m.Completed,
		CompletedAt:    m.CompletedAt,
		PriorityScore:  m.PriorityScore,
		PriorityLevel:  m.PriorityLevel,
		PriorityReason: m.PriorityReason,
		LastUpdatedAt:  m.LastUpdatedAt,
		CreatedAt:      m.CreatedAt,
	}
}

func (r *TasksRepository) CreateTask(ctx context.Context, task domain.Task, priorityScore float64, priorityLevel string, priorityReason string) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var m taskModel
	err := r.pool.QueryRow(ctx, `
		INSERT INTO todoapp.tasks (
			user_id, title, description, deadline, importance, category,
			priority_score, priority_level, priority_reason
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, version, user_id, title, description, deadline,
			importance, category, completed, completed_at,
			priority_score, priority_level, priority_reason,
			last_updated_at, created_at
	`,
		task.UserID, task.Title, task.Description, task.Deadline,
		task.Importance, task.Category,
		priorityScore, priorityLevel, priorityReason,
	).Scan(
		&m.ID, &m.Version, &m.UserID, &m.Title, &m.Description, &m.Deadline,
		&m.Importance, &m.Category, &m.Completed, &m.CompletedAt,
		&m.PriorityScore, &m.PriorityLevel, &m.PriorityReason,
		&m.LastUpdatedAt, &m.CreatedAt,
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf("insert task: %w", err)
	}

	return m.toDomain(), nil
}

func (r *TasksRepository) GetTasks(ctx context.Context, userID *int, category *string, completed *bool, page int, pageSize int) ([]domain.Task, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	conditions := []string{}
	args := []any{}
	argIdx := 1

	if userID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *userID)
		argIdx++
	}
	if category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *category)
		argIdx++
	}
	if completed != nil {
		conditions = append(conditions, fmt.Sprintf("completed = $%d", argIdx))
		args = append(args, *completed)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countArgs := make([]any, len(args))
	copy(countArgs, args)

	var total int
	err := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM todoapp.tasks %s", where),
		countArgs...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, version, user_id, title, description, deadline,
			importance, category, completed, completed_at,
			priority_score, priority_level, priority_reason,
			last_updated_at, created_at
		FROM todoapp.tasks %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var m taskModel
		if err := rows.Scan(
			&m.ID, &m.Version, &m.UserID, &m.Title, &m.Description, &m.Deadline,
			&m.Importance, &m.Category, &m.Completed, &m.CompletedAt,
			&m.PriorityScore, &m.PriorityLevel, &m.PriorityReason,
			&m.LastUpdatedAt, &m.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, m.toDomain())
	}

	return tasks, total, nil
}

func (r *TasksRepository) GetTask(ctx context.Context, id int) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var m taskModel
	err := r.pool.QueryRow(ctx, `
		SELECT id, version, user_id, title, description, deadline,
			importance, category, completed, completed_at,
			priority_score, priority_level, priority_reason,
			last_updated_at, created_at
		FROM todoapp.tasks WHERE id = $1
	`, id).Scan(
		&m.ID, &m.Version, &m.UserID, &m.Title, &m.Description, &m.Deadline,
		&m.Importance, &m.Category, &m.Completed, &m.CompletedAt,
		&m.PriorityScore, &m.PriorityLevel, &m.PriorityReason,
		&m.LastUpdatedAt, &m.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Task{}, core_errors.ErrNotFound
	}
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}

	return m.toDomain(), nil
}

func (r *TasksRepository) DeleteTask(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tag, err := r.pool.Exec(ctx, `DELETE FROM todoapp.tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return core_errors.ErrNotFound
	}

	return nil
}

func (r *TasksRepository) PatchTask(ctx context.Context, id int, version int, patch domain.TaskPatch, priorityScore float64, priorityLevel string, priorityReason string) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sets := []string{}
	args := []any{}
	argIdx := 1

	if patch.Title.Valid {
		sets = append(sets, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, patch.Title.Value)
		argIdx++
	}
	if patch.Description.Valid {
		sets = append(sets, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, patch.Description.Value)
		argIdx++
	}
	if patch.Deadline.Valid {
		sets = append(sets, fmt.Sprintf("deadline = $%d", argIdx))
		args = append(args, patch.Deadline.Value)
		argIdx++
	}
	if patch.Importance.Valid {
		sets = append(sets, fmt.Sprintf("importance = $%d", argIdx))
		args = append(args, patch.Importance.Value)
		argIdx++
	}
	if patch.Category.Valid {
		sets = append(sets, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, patch.Category.Value)
		argIdx++
	}
	if patch.Completed.Valid {
		sets = append(sets, fmt.Sprintf("completed = $%d", argIdx))
		args = append(args, patch.Completed.Value)
		argIdx++
		if patch.Completed.Value {
			sets = append(sets, fmt.Sprintf("completed_at = $%d", argIdx))
			args = append(args, time.Now().UTC())
			argIdx++
		} else {
			sets = append(sets, "completed_at = NULL")
		}
	}

	sets = append(sets,
		fmt.Sprintf("priority_score = $%d", argIdx),
		fmt.Sprintf("priority_level = $%d", argIdx+1),
		fmt.Sprintf("priority_reason = $%d", argIdx+2),
		fmt.Sprintf("last_updated_at = $%d", argIdx+3),
		"version = version + 1",
	)
	args = append(args, priorityScore, priorityLevel, priorityReason, time.Now().UTC())
	argIdx += 4

	args = append(args, id, version)

	var m taskModel
	err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		UPDATE todoapp.tasks SET %s
		WHERE id = $%d AND version = $%d
		RETURNING id, version, user_id, title, description, deadline,
			importance, category, completed, completed_at,
			priority_score, priority_level, priority_reason,
			last_updated_at, created_at
	`, strings.Join(sets, ", "), argIdx, argIdx+1), args...).Scan(
		&m.ID, &m.Version, &m.UserID, &m.Title, &m.Description, &m.Deadline,
		&m.Importance, &m.Category, &m.Completed, &m.CompletedAt,
		&m.PriorityScore, &m.PriorityLevel, &m.PriorityReason,
		&m.LastUpdatedAt, &m.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Task{}, core_errors.ErrConflict
	}
	if err != nil {
		return domain.Task{}, fmt.Errorf("patch task: %w", err)
	}

	return m.toDomain(), nil
}

func (r *TasksRepository) GetStagnantTasks(ctx context.Context, userID *int) ([]domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	args := []any{}
	userFilter := ""
	if userID != nil {
		userFilter = "AND user_id = $1"
		args = append(args, *userID)
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, version, user_id, title, description, deadline,
			importance, category, completed, completed_at,
			priority_score, priority_level, priority_reason,
			last_updated_at, created_at
		FROM todoapp.tasks
		WHERE importance >= 3
			AND completed = false
			AND last_updated_at < NOW() - INTERVAL '5 days'
			%s
		ORDER BY importance DESC, last_updated_at ASC
	`, userFilter), args...)
	if err != nil {
		return nil, fmt.Errorf("query stagnant tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var m taskModel
		if err := rows.Scan(
			&m.ID, &m.Version, &m.UserID, &m.Title, &m.Description, &m.Deadline,
			&m.Importance, &m.Category, &m.Completed, &m.CompletedAt,
			&m.PriorityScore, &m.PriorityLevel, &m.PriorityReason,
			&m.LastUpdatedAt, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan stagnant task: %w", err)
		}
		tasks = append(tasks, m.toDomain())
	}

	return tasks, nil
}