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
			return fmt.Errorf("every_n_days must be positive")
		}
	case RecurrenceMonthlyDays:
		if len(r.MonthDays) == 0 {
			return fmt.Errorf("month_days can't be empty")
		}
		for _, d := range r.MonthDays {
			if d < 1 || d > 30 {
				return fmt.Errorf("month day %d is out of range (1-30)", d)
			}
		}
	case RecurrenceSpecificDates:
		if len(r.SpecificDates) == 0 {
			return fmt.Errorf("specific_dates can't be empty")
		}
	case RecurrenceParity:
		if r.Parity != ParityEven && r.Parity != ParityOdd {
			return fmt.Errorf("parity must be 'even' or 'odd'")
		}
	default:
		return fmt.Errorf("unknown recurrence type: %s", r.Type)
	}

	// specific_dates doesn't need a range
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

func (r *Recurrence) Dates() []time.Time {
	switch r.Type {
	case RecurrenceDaily:
		return r.dailyDates()
	case RecurrenceMonthlyDays:
		return r.monthlyDaysDates()
	case RecurrenceSpecificDates:
		return r.SpecificDates
	case RecurrenceParity:
		return r.parityDates()
	}
	return nil
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

	// go month by month from start to end
	cur := time.Date(r.StartDate.Year(), r.StartDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(r.EndDate.Year(), r.EndDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !cur.After(endMonth) {
		for _, day := range r.MonthDays {
			candidate := time.Date(cur.Year(), cur.Month(), day, 0, 0, 0, 0, time.UTC)
			// skip if day doesn't exist in this month (e.g. Feb 30)
			if candidate.Month() != cur.Month() {
				continue
			}
			if !candidate.Before(r.StartDate) && !candidate.After(r.EndDate) {
				dates = append(dates, candidate)
			}
		}
		cur = cur.AddDate(0, 1, 0)
	}

	return dates
}

func (r *Recurrence) parityDates() []time.Time {
	var dates []time.Time
	for d := r.StartDate; !d.After(r.EndDate); d = d.AddDate(0, 0, 1) {
		isEven := d.Day()%2 == 0
		if r.Parity == ParityEven && isEven {
			dates = append(dates, d)
		} else if r.Parity == ParityOdd && !isEven {
			dates = append(dates, d)
		}
	}
	return dates
}
