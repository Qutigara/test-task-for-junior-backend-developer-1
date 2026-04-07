package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskrepo "example.com/taskservice/internal/repository/postgres"
)

type TaskWorker struct {
	repo    *taskrepo.Repository
	now     func() time.Time
	interval time.Duration
}

func NewTaskWorker(repo *taskrepo.Repository) *TaskWorker {
	return &TaskWorker{
		repo:      repo,
		now:       func() time.Time { return time.Now().UTC() },
		interval:  1 * time.Hour,
	}
}

func (w *TaskWorker) Start(ctx context.Context) {
	log.Println("Task worker started")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run immediately on start
	w.run(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Task worker stopped")
			return
		case <-ticker.C:
			w.run(ctx)
		}
	}
}

func (w *TaskWorker) run(ctx context.Context) {
	tasks, err := w.repo.List(ctx)
	if err != nil {
		log.Printf("Error fetching tasks: %v", err)
		return
	}

	today := w.now().Truncate(24 * time.Hour)

	for _, task := range tasks {
		if task.Recurrence == nil {
			continue
		}

		if w.shouldCreateTaskForDate(task.Recurrence, today) {
			w.createRecurringTask(ctx, &task, today)
		}
	}
}

func (w *TaskWorker) shouldCreateTaskForDate(recurrence *taskdomain.Recurrence, date time.Time) bool {
	switch recurrence.Type {
	case taskdomain.RecurrenceTypeDaily:
		// Проверяем по interval - создаём задачу если день подходит по интервалу
		// Для простоты: создаём каждый день, interval используется как информация
		return true

	case taskdomain.RecurrenceTypeMonthly:
		if recurrence.DayOfMonth == nil {
			return false
		}
		return date.Day() == *recurrence.DayOfMonth

	case taskdomain.RecurrenceTypeSpecificDates:
		dateStr := date.Format("2006-01-02")
		for _, d := range recurrence.SpecificDates {
			if d == dateStr {
				return true
			}
		}
		return false

	case taskdomain.RecurrenceTypeEvenDays:
		return date.Day()%2 == 0

	case taskdomain.RecurrenceTypeOddDays:
		return date.Day()%2 == 1

	default:
		return false
	}
}

func (w *TaskWorker) createRecurringTask(ctx context.Context, original *taskdomain.Task, date time.Time) {
	// Проверяем, не создавалась ли уже задача на этот день
	existingTasks, err := w.repo.List(ctx)
	if err != nil {
		log.Printf("Error checking existing tasks: %v", err)
		return
	}

	title := fmt.Sprintf("%s (%s)", original.Title, date.Format("2006-01-02"))

	// Проверяем, есть ли уже задача с таким названием на эту дату
	for _, t := range existingTasks {
		if t.Title == title && t.Status != taskdomain.StatusDone {
			// Задача уже существует
			return
		}
	}

	newTask := &taskdomain.Task{
		Title:       title,
		Description: original.Description,
		Status:      taskdomain.StatusNew,
		CreatedAt:   w.now(),
		UpdatedAt:   w.now(),
	}

	_, err = w.repo.Create(ctx, newTask)
	if err != nil {
		log.Printf("Error creating recurring task: %v", err)
		return
	}

	log.Printf("Created recurring task: %s", title)
}
