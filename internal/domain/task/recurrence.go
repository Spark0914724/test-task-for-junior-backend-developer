package task

import (
	"fmt"
	"time"
)

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthlyDays   RecurrenceType = "monthly_days"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceParity        RecurrenceType = "parity"
)

type ParityType string

const (
	ParityEven ParityType = "even"
	ParityOdd  ParityType = "odd"
)

type Recurrence struct {
	ID            int64
	Type          RecurrenceType
	EveryNDays    int
	MonthDays     []int
	SpecificDates []time.Time
	Parity        ParityType
	StartDate     time.Time
	EndDate       time.Time
}

func (r *Recurrence) Validate() error {
	switch r.Type {
	case RecurrenceDaily:
		if r.EveryNDays <= 0 {
			return fmt.Errorf("every_n_days must be positive for daily recurrence")
		}
	case RecurrenceMonthlyDays:
		if len(r.MonthDays) == 0 {
			return fmt.Errorf("month_days is required for monthly_days recurrence")
		}
		for _, d := range r.MonthDays {
			if d < 1 || d > 30 {
				return fmt.Errorf("month_days values must be between 1 and 30")
			}
		}
	case RecurrenceSpecificDates:
		if len(r.SpecificDates) == 0 {
			return fmt.Errorf("specific_dates is required for specific_dates recurrence")
		}
	case RecurrenceParity:
		if r.Parity != ParityEven && r.Parity != ParityOdd {
			return fmt.Errorf("parity must be 'even' or 'odd'")
		}
	default:
		return fmt.Errorf("unknown recurrence type: %s", r.Type)
	}

	if r.Type != RecurrenceSpecificDates {
		if r.StartDate.IsZero() || r.EndDate.IsZero() {
			return fmt.Errorf("start_date and end_date are required")
		}
		if !r.EndDate.After(r.StartDate) {
			return fmt.Errorf("end_date must be after start_date")
		}
	}

	return nil
}

// Dates returns all dates matching the recurrence rule.
func (r *Recurrence) Dates() []time.Time {
	switch r.Type {
	case RecurrenceDaily:
		return r.dailyDates()
	case RecurrenceMonthlyDays:
		return r.monthlyDaysDates()
	case RecurrenceSpecificDates:
		return r.specificDates()
	case RecurrenceParity:
		return r.parityDates()
	default:
		return nil
	}
}

func (r *Recurrence) dailyDates() []time.Time {
	var dates []time.Time
	for d := r.StartDate; !d.After(r.EndDate); d = d.AddDate(0, 0, r.EveryNDays) {
		dates = append(dates, d)
	}
	return dates
}

func (r *Recurrence) monthlyDaysDates() []time.Time {
	var dates []time.Time
	// iterate month by month
	start := firstOfMonth(r.StartDate)
	end := firstOfMonth(r.EndDate)
	for m := start; !m.After(end); m = m.AddDate(0, 1, 0) {
		for _, day := range r.MonthDays {
			candidate := time.Date(m.Year(), m.Month(), day, 0, 0, 0, 0, time.UTC)
			// skip if the month doesn't have this day (e.g. day=30 in February)
			if candidate.Month() != m.Month() {
				continue
			}
			if (candidate.Equal(r.StartDate) || candidate.After(r.StartDate)) &&
				(candidate.Equal(r.EndDate) || candidate.Before(r.EndDate)) {
				dates = append(dates, candidate)
			}
		}
	}
	return dates
}

func (r *Recurrence) specificDates() []time.Time {
	var dates []time.Time
	for _, d := range r.SpecificDates {
		dates = append(dates, d)
	}
	return dates
}

func (r *Recurrence) parityDates() []time.Time {
	var dates []time.Time
	for d := r.StartDate; !d.After(r.EndDate); d = d.AddDate(0, 0, 1) {
		day := d.Day()
		if r.Parity == ParityEven && day%2 == 0 {
			dates = append(dates, d)
		} else if r.Parity == ParityOdd && day%2 != 0 {
			dates = append(dates, d)
		}
	}
	return dates
}

func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}
