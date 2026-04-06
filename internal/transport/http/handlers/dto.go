package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID            int64             `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	RecurrenceID  *int64            `json:"recurrence_id,omitempty"`
	ScheduledDate *time.Time        `json:"scheduled_date,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:            task.ID,
		Title:         task.Title,
		Description:   task.Description,
		Status:        task.Status,
		RecurrenceID:  task.RecurrenceID,
		ScheduledDate: task.ScheduledDate,
		CreatedAt:     task.CreatedAt,
		UpdatedAt:     task.UpdatedAt,
	}
}

// recurrenceDTO holds the recurrence rule from the request
type recurrenceDTO struct {
	Type          taskdomain.RecurrenceType `json:"type"`
	EveryNDays    int                       `json:"every_n_days,omitempty"`
	MonthDays     []int                     `json:"month_days,omitempty"`
	SpecificDates []time.Time               `json:"specific_dates,omitempty"`
	Parity        taskdomain.ParityType     `json:"parity,omitempty"`
}

type recurringTaskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	StartDate   time.Time         `json:"start_date"`
	EndDate     time.Time         `json:"end_date"`
	Recurrence  recurrenceDTO     `json:"recurrence"`
}
