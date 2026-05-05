package domain

import (
	"fmt"
	"time"

	core_errors "github.com/aaaaarsen/golang-todoapp/internal/core/errors"
)

type Task struct {
	ID      int
	Version int
	UserID  int

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

type TaskPatch struct {
	Title       Nullable[string]
	Description Nullable[*string]
	Deadline    Nullable[*time.Time]
	Importance  Nullable[int]
	Category    Nullable[string]
	Completed   Nullable[bool]
}

func NewTask(
	userID int,
	title string,
	description *string,
	deadline *time.Time,
	importance int,
	category string,
) Task {
	return Task{
		ID:          UninitializedID,
		Version:     UninitializedVersion,
		UserID:      userID,
		Title:       title,
		Description: description,
		Deadline:    deadline,
		Importance:  importance,
		Category:    category,
		Completed:   false,
	}
}

func (t *Task) Validate() error {
	titleLen := len([]rune(t.Title))
	if titleLen < 1 || titleLen > 255 {
		return fmt.Errorf(
			"invalid `Title` len: %d: %w",
			titleLen,
			core_errors.ErrInvalidArgument,
		)
	}

	if t.Importance < 1 || t.Importance > 5 {
		return fmt.Errorf(
			"invalid `Importance` value: %d: %w",
			t.Importance,
			core_errors.ErrInvalidArgument,
		)
	}

	if t.Category != "work" && t.Category != "study" && t.Category != "personal" {
		return fmt.Errorf(
			"invalid `Category` value: %s: %w",
			t.Category,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}