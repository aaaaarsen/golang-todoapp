package statistics_service

import (
	"context"
	"fmt"
	"time"

	statistics_postgres_repository "github.com/aaaaarsen/golang-todoapp/internal/features/statistics/repository/postgres"
)

type StatisticsResult struct {
	TotalTasks             int     `json:"total_tasks"`
	CompletedTasks         int     `json:"completed_tasks"`
	CompletionRate         float64 `json:"completion_rate"`
	OverdueTasks           int     `json:"overdue_tasks"`
	HighPriorityTasks      int     `json:"high_priority_tasks"`
	AvgCompletionTimeHours float64 `json:"avg_completion_time_hours"`
}

type StatisticsRepository interface {
	GetStatistics(
		ctx context.Context,
		userID *int,
		from *time.Time,
		to *time.Time,
		category *string,
	) (statistics_postgres_repository.StatisticsResult, error)
}

type StatisticsService struct {
	statisticsRepository StatisticsRepository
}

func NewStatisticsService(statisticsRepository StatisticsRepository) *StatisticsService {
	return &StatisticsService{
		statisticsRepository: statisticsRepository,
	}
}

func (s *StatisticsService) GetStatistics(
	ctx context.Context,
	userID *int,
	from *time.Time,
	to *time.Time,
	category *string,
) (StatisticsResult, error) {
	raw, err := s.statisticsRepository.GetStatistics(ctx, userID, from, to, category)
	if err != nil {
		return StatisticsResult{}, fmt.Errorf("get statistics: %w", err)
	}

	var completionRate float64
	if raw.TotalTasks > 0 {
		completionRate = float64(raw.CompletedTasks) / float64(raw.TotalTasks)
	}

	return StatisticsResult{
		TotalTasks:             raw.TotalTasks,
		CompletedTasks:         raw.CompletedTasks,
		CompletionRate:         completionRate,
		OverdueTasks:           raw.OverdueTasks,
		HighPriorityTasks:      raw.HighPriorityTasks,
		AvgCompletionTimeHours: raw.AvgCompletionTimeHours,
	}, nil
}