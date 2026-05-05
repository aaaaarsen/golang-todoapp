package tasks_priority

import (
	"fmt"
	"math"
	"time"

	"github.com/aaaaarsen/golang-todoapp/internal/core/domain"
)

type PriorityResult struct {
	Score  float64
	Level  string
	Reason string
}

type PriorityEngine struct{}

func NewPriorityEngine() *PriorityEngine {
	return &PriorityEngine{}
}

func (e *PriorityEngine) Calculate(task domain.Task) PriorityResult {
	deadlineScore := e.calcDeadlineScore(task.Deadline)
	importanceScore := float64(task.Importance-1) / 4.0
	stagnationScore, stagnationDays := e.calcStagnationScore(task.LastUpdatedAt)
	dependenciesScore := 0.5

	score := 0.35*deadlineScore + 0.30*importanceScore + 0.20*dependenciesScore + 0.15*stagnationScore
	score = math.Round(score*100) / 100

	level := "normal"
	if score >= 0.6 {
		level = "high"
	}

	reason := e.buildReason(deadlineScore, importanceScore, stagnationScore, stagnationDays, task.Deadline)

	return PriorityResult{
		Score:  score,
		Level:  level,
		Reason: reason,
	}
}

func (e *PriorityEngine) calcDeadlineScore(deadline *time.Time) float64 {
	if deadline == nil {
		return 0.0
	}

	now := time.Now().UTC()
	diff := deadline.UTC().Sub(now)

	if diff <= 0 {
		return 1.0
	}

	days := diff.Hours() / 24

	switch {
	case days <= 1:
		return 0.9
	case days <= 3:
		return 0.7
	case days <= 7:
		return 0.5
	case days <= 14:
		return 0.3
	default:
		return 0.1
	}
}

func (e *PriorityEngine) calcStagnationScore(lastUpdatedAt time.Time) (float64, int) {
	days := int(time.Since(lastUpdatedAt).Hours() / 24)

	switch {
	case days >= 7:
		return 1.0, days
	case days >= 3:
		return 0.5, days
	default:
		return 0.0, days
	}
}

func (e *PriorityEngine) buildReason(
	deadlineScore float64,
	importanceScore float64,
	stagnationScore float64,
	stagnationDays int,
	deadline *time.Time,
) string {
	if deadlineScore >= 0.7 {
		if deadline != nil {
			daysLeft := int(time.Until(*deadline).Hours() / 24)
			if daysLeft <= 0 {
				return "Высокий приоритет: дедлайн просрочен"
			}
			return fmt.Sprintf("Высокий приоритет: дедлайн через %d дней", daysLeft)
		}
	}

	if importanceScore >= 0.75 {
		return "Высокий приоритет: важность задачи"
	}

	if stagnationScore >= 0.5 {
		return fmt.Sprintf("Требует внимания: задача не обновлялась %d дней", stagnationDays)
	}

	return "Нормальный приоритет"
}