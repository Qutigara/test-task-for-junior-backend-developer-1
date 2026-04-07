package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	var recurrenceType string
	var recurrenceInterval int
	var recurrenceDayOfMonth sql.NullInt32
	var recurrenceSpecificDates []byte

	if task.Recurrence != nil {
		recurrenceType = string(task.Recurrence.Type)
		recurrenceInterval = task.Recurrence.Interval
		if task.Recurrence.DayOfMonth != nil {
			recurrenceDayOfMonth = sql.NullInt32{Int32: int32(*task.Recurrence.DayOfMonth), Valid: true}
		}
		if len(task.Recurrence.SpecificDates) > 0 {
			specificDatesJSON, _ := json.Marshal(task.Recurrence.SpecificDates)
			recurrenceSpecificDates = specificDatesJSON
		}
	}

	const query = `
		INSERT INTO tasks (title, description, status, recurrence_type, recurrence_interval, recurrence_day_of_month, recurrence_specific_dates, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, title, description, status, recurrence_type, recurrence_interval, recurrence_day_of_month, recurrence_specific_dates, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, recurrenceType, recurrenceInterval, recurrenceDayOfMonth, recurrenceSpecificDates, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence_type, recurrence_interval, recurrence_day_of_month, recurrence_specific_dates, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	var recurrenceType string
	var recurrenceInterval int
	var recurrenceDayOfMonth sql.NullInt32
	var recurrenceSpecificDates []byte

	if task.Recurrence != nil {
		recurrenceType = string(task.Recurrence.Type)
		recurrenceInterval = task.Recurrence.Interval
		if task.Recurrence.DayOfMonth != nil {
			recurrenceDayOfMonth = sql.NullInt32{Int32: int32(*task.Recurrence.DayOfMonth), Valid: true}
		}
		if len(task.Recurrence.SpecificDates) > 0 {
			specificDatesJSON, _ := json.Marshal(task.Recurrence.SpecificDates)
			recurrenceSpecificDates = specificDatesJSON
		}
	}

	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			recurrence_type = $4,
			recurrence_interval = $5,
			recurrence_day_of_month = $6,
			recurrence_specific_dates = $7,
			updated_at = $8
		WHERE id = $9
		RETURNING id, title, description, status, recurrence_type, recurrence_interval, recurrence_day_of_month, recurrence_specific_dates, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, recurrenceType, recurrenceInterval, recurrenceDayOfMonth, recurrenceSpecificDates, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence_type, recurrence_interval, recurrence_day_of_month, recurrence_specific_dates, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task                    taskdomain.Task
		status                  string
		recurrenceType          string
		recurrenceInterval      int
		recurrenceDayOfMonth    sql.NullInt32
		recurrenceSpecificDates []byte
	)

	err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&recurrenceType,
		&recurrenceInterval,
		&recurrenceDayOfMonth,
		&recurrenceSpecificDates,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	if recurrenceType != "" && recurrenceType != string(taskdomain.RecurrenceTypeNone) {
		task.Recurrence = &taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceType(recurrenceType),
			Interval:   recurrenceInterval,
		}
		if recurrenceDayOfMonth.Valid {
			dayOfMonth := int(recurrenceDayOfMonth.Int32)
			task.Recurrence.DayOfMonth = &dayOfMonth
		}
		if len(recurrenceSpecificDates) > 0 {
			json.Unmarshal(recurrenceSpecificDates, &task.Recurrence.SpecificDates)
		}
	}

	return &task, nil
}
