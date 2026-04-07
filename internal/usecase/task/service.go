package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:  normalized.Recurrence,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:  normalized.Recurrence,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := validateRecurrence(input.Recurrence); err != nil {
		return CreateInput{}, err
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := validateRecurrence(input.Recurrence); err != nil {
		return UpdateInput{}, err
	}

	return input, nil
}

func validateRecurrence(r *taskdomain.Recurrence) error {
	if r == nil {
		return nil
	}

	if !r.Type.Valid() {
		return fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	switch r.Type {
	case taskdomain.RecurrenceTypeDaily:
		if r.Interval <= 0 {
			return fmt.Errorf("%w: interval must be positive for daily recurrence", ErrInvalidInput)
		}
	case taskdomain.RecurrenceTypeMonthly:
		if r.DayOfMonth == nil || *r.DayOfMonth < 1 || *r.DayOfMonth > 30 {
			return fmt.Errorf("%w: day_of_month must be between 1 and 30 for monthly recurrence", ErrInvalidInput)
		}
	case taskdomain.RecurrenceTypeSpecificDates:
		if len(r.SpecificDates) == 0 {
			return fmt.Errorf("%w: specific_dates required for specific_dates recurrence", ErrInvalidInput)
		}
		if err := validateSpecificDates(r.SpecificDates); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	case taskdomain.RecurrenceTypeEvenDays, taskdomain.RecurrenceTypeOddDays:
		// no additional validation needed
	case taskdomain.RecurrenceTypeNone:
		// no additional validation needed
	}

	return nil
}

func validateSpecificDates(dates []string) error {
	layout := "2006-01-02"
	today := time.Now().UTC().Truncate(24 * time.Hour)

	for _, dateStr := range dates {
		dateStr = strings.TrimSpace(dateStr)
		if dateStr == "" {
			return fmt.Errorf("specific_dates contains empty date")
		}

		parsedDate, err := time.Parse(layout, dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format '%s', expected YYYY-MM-DD", dateStr)
		}

		if parsedDate.Before(today) {
			return fmt.Errorf("date '%s' is in the past", dateStr)
		}
	}

	return nil
}
