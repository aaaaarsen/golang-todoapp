package statistics_postgres_repository

import (
	"context"
	"fmt"
	"time"

	core_postgres_pool "github.com/aaaaarsen/golang-todoapp/internal/core/repository/postgres/conn"
)

type StatisticsRepository struct {
	pool core_postgres_pool.Pool
}

func NewStatisticsRepository(pool core_postgres_pool.Pool) *StatisticsRepository {
	return &StatisticsRepository{pool: pool}
}

type StatisticsResult struct {
	TotalTasks             int
	CompletedTasks         int
	OverdueTasks           int
	HighPriorityTasks      int
	AvgCompletionTimeHours float64
}

func (r *StatisticsRepository) GetStatistics(
	ctx context.Context,
	userID *int,
	from *time.Time,
	to *time.Time,
	category *string,
) (StatisticsResult, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	args := []any{}
	argIdx := 1
	conditions := []string{}

	if userID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *userID)
		argIdx++
	}
	if from != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *from)
		argIdx++
	}
	if to != nil {
		conditions = append(conditions, fmt.Sprintf("created_at < $%d", argIdx))
		args = append(args, *to)
		argIdx++
	}
	if category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *category)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + joinConditions(conditions)
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(*)                                                        AS total_tasks,
			COUNT(*) FILTER (WHERE completed = true)                        AS completed_tasks,
			COUNT(*) FILTER (WHERE deadline < NOW() AND completed = false)  AS overdue_tasks,
			COUNT(*) FILTER (WHERE priority_level = 'high')                 AS high_priority_tasks,
			COALESCE(
				AVG(
					EXTRACT(EPOCH FROM (completed_at - created_at)) / 3600.0
				) FILTER (WHERE completed = true AND completed_at IS NOT NULL),
				0
			)                                                               AS avg_completion_time_hours
		FROM todoapp.tasks
		%s
	`, where)

	var result StatisticsResult
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&result.TotalTasks,
		&result.CompletedTasks,
		&result.OverdueTasks,
		&result.HighPriorityTasks,
		&result.AvgCompletionTimeHours,
	)
	if err != nil {
		return StatisticsResult{}, fmt.Errorf("get statistics: %w", err)
	}

	return result, nil
}

func joinConditions(conditions []string) string {
	result := ""
	for i, c := range conditions {
		if i > 0 {
			result += " AND "
		}
		result += c
	}
	return result
}