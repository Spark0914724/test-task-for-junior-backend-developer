package postgres

import (
	"context"
	"errors"
	"time"

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
	const query = `
		INSERT INTO tasks (title, description, status, recurrence_id, scheduled_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, recurrence_id, scheduled_date, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.RecurrenceID, task.ScheduledDate, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence_id, scheduled_date, created_at, updated_at
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
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, status, recurrence_id, scheduled_date, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
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
		SELECT id, title, description, status, recurrence_id, scheduled_date, created_at, updated_at
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
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.RecurrenceID,
		&task.ScheduledDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}

func (r *Repository) CreateRecurrence(ctx context.Context, rec *taskdomain.Recurrence) (*taskdomain.Recurrence, error) {
	const query = `
		INSERT INTO task_recurrences (type, every_n_days, month_days, specific_dates, parity, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, type, every_n_days, month_days, specific_dates, parity, start_date, end_date
	`

	var everyNDays *int
	if rec.EveryNDays > 0 {
		everyNDays = &rec.EveryNDays
	}

	var parity *string
	if rec.Parity != "" {
		p := string(rec.Parity)
		parity = &p
	}

	specificDates := make([]time.Time, len(rec.SpecificDates))
	copy(specificDates, rec.SpecificDates)

	row := r.pool.QueryRow(ctx, query,
		string(rec.Type),
		everyNDays,
		rec.MonthDays,
		specificDates,
		parity,
		rec.StartDate,
		rec.EndDate,
	)

	return scanRecurrence(row)
}

func (r *Repository) BulkCreateTasks(ctx context.Context, tasks []*taskdomain.Task) ([]taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, recurrence_id, scheduled_date, created_at, updated_at)
		SELECT * FROM unnest($1::text[], $2::text[], $3::text[], $4::bigint[], $5::date[], $6::timestamptz[], $7::timestamptz[])
		RETURNING id, title, description, status, recurrence_id, scheduled_date, created_at, updated_at
	`

	titles := make([]string, len(tasks))
	descriptions := make([]string, len(tasks))
	statuses := make([]string, len(tasks))
	recurrenceIDs := make([]*int64, len(tasks))
	scheduledDates := make([]*time.Time, len(tasks))
	createdAts := make([]time.Time, len(tasks))
	updatedAts := make([]time.Time, len(tasks))

	for i, t := range tasks {
		titles[i] = t.Title
		descriptions[i] = t.Description
		statuses[i] = string(t.Status)
		recurrenceIDs[i] = t.RecurrenceID
		scheduledDates[i] = t.ScheduledDate
		createdAts[i] = t.CreatedAt
		updatedAts[i] = t.UpdatedAt
	}

	rows, err := r.pool.Query(ctx, query, titles, descriptions, statuses, recurrenceIDs, scheduledDates, createdAts, updatedAts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]taskdomain.Task, 0, len(tasks))
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *task)
	}

	return result, rows.Err()
}

type recurrenceScanner interface {
	Scan(dest ...any) error
}

func scanRecurrence(scanner recurrenceScanner) (*taskdomain.Recurrence, error) {
	var (
		rec         taskdomain.Recurrence
		recType     string
		everyNDays  *int
		parity      *string
	)

	if err := scanner.Scan(
		&rec.ID,
		&recType,
		&everyNDays,
		&rec.MonthDays,
		&rec.SpecificDates,
		&parity,
		&rec.StartDate,
		&rec.EndDate,
	); err != nil {
		return nil, err
	}

	rec.Type = taskdomain.RecurrenceType(recType)
	if everyNDays != nil {
		rec.EveryNDays = *everyNDays
	}
	if parity != nil {
		rec.Parity = taskdomain.ParityType(*parity)
	}

	return &rec, nil
}
