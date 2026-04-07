package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type RecurrenceType string

const (
	RecurrenceTypeNone          RecurrenceType = "none"
	RecurrenceTypeDaily         RecurrenceType = "daily"
	RecurrenceTypeMonthly       RecurrenceType = "monthly"
	RecurrenceTypeSpecificDates RecurrenceType = "specific_dates"
	RecurrenceTypeEvenDays      RecurrenceType = "even_days"
	RecurrenceTypeOddDays       RecurrenceType = "odd_days"
)

func (rt RecurrenceType) Valid() bool {
	switch rt {
	case RecurrenceTypeNone, RecurrenceTypeDaily, RecurrenceTypeMonthly, RecurrenceTypeSpecificDates, RecurrenceTypeEvenDays, RecurrenceTypeOddDays:
		return true
	default:
		return false
	}
}

type Recurrence struct {
	Type          RecurrenceType `json:"type"`
	Interval      int            `json:"interval"`       // для daily: каждый n-й день
	DayOfMonth    *int           `json:"day_of_month,omitempty"`    // для monthly: число месяца (1-30)
	SpecificDates []string       `json:"specific_dates,omitempty"`  // для specific_dates: даты в формате "YYYY-MM-DD"
}

type Task struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      Status       `json:"status"`
	Recurrence  *Recurrence  `json:"recurrence,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}