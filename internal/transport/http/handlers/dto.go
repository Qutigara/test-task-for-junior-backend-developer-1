package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type recurrenceDTO struct {
	Type          taskdomain.RecurrenceType `json:"type"`
	Interval      int                        `json:"interval"`
	DayOfMonth    *int                       `json:"day_of_month,omitempty"`
	SpecificDates []string                   `json:"specific_dates,omitempty"`
}

type taskMutationDTO struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Status      taskdomain.Status    `json:"status"`
	Recurrence  *recurrenceDTO       `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID          int64                 `json:"id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Status      taskdomain.Status     `json:"status"`
	Recurrence  *taskdomain.Recurrence `json:"recurrence,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Recurrence:  task.Recurrence,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
