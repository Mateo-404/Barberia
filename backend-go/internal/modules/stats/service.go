// Package stats computes the admin dashboard figures.
package stats

import (
	"context"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database"
	"github.com/Mateo-404/Barberia/backend-go/internal/platform/database/sqlc"
	"github.com/Mateo-404/Barberia/backend-go/internal/timeutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DailyRevenue is one day of the revenue chart.
type DailyRevenue struct {
	Date  time.Time
	Total float64
}

// ServiceCount is the number of appointments of a service.
type ServiceCount struct {
	ID    int64
	Name  string
	Count int64
}

// HourCount is the number of appointments that start at an hour of the day.
type HourCount struct {
	Hour  int32
	Count int64
}

// Dashboard is the full dashboard payload.
type Dashboard struct {
	AppointmentsToday     int64
	AppointmentsYesterday int64
	RevenueThisMonth      float64
	RevenueLastMonth      float64
	CustomerCount         int64
	DailyRevenue          []DailyRevenue
	Services              []ServiceCount
	Hours                 []HourCount
}

func ts(t time.Time) pgtype.Timestamp { return pgtype.Timestamp{Time: t, Valid: true} }

// Service computes the dashboard.
type Service struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
	now  func() time.Time
}

// New builds the service.
func New(pool *pgxpool.Pool, q *sqlc.Queries) *Service {
	return &Service{pool: pool, q: q, now: time.Now}
}

// Dashboard computes today's figures, the last 30 days of revenue, service
// popularity over the last month and appointments per hour this year.
func (s *Service) Dashboard(ctx context.Context, tenantID int64) (*Dashboard, error) {
	var out *Dashboard
	err := database.WithTenantTx(ctx, s.pool, func(q *sqlc.Queries) error {
		now := s.now()

		today, err := q.CountAppointmentsOnDate(ctx, sqlc.CountAppointmentsOnDateParams{TenantID: tenantID, Day: pgtype.Date{Time: now, Valid: true}})
		if err != nil {
			return err
		}
		yesterday, err := q.CountAppointmentsOnDate(ctx, sqlc.CountAppointmentsOnDateParams{TenantID: tenantID, Day: pgtype.Date{Time: now.AddDate(0, 0, -1), Valid: true}})
		if err != nil {
			return err
		}

		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		nextMonthStart := timeutil.PlusMonths(monthStart, 1)
		lastMonthStart := timeutil.MinusMonths(monthStart, 1)

		revenue := func(from, until time.Time) (float64, error) {
			return q.SumRevenueBetween(ctx, sqlc.SumRevenueBetweenParams{TenantID: tenantID, FromTs: ts(from), UntilTs: ts(until)})
		}
		thisMonth, err := revenue(monthStart, nextMonthStart)
		if err != nil {
			return err
		}
		lastMonth, err := revenue(lastMonthStart, monthStart)
		if err != nil {
			return err
		}

		customers, err := q.CountCustomers(ctx, tenantID)
		if err != nil {
			return err
		}

		oneMonthAgo := timeutil.MinusMonths(now, 1)
		daily := make([]DailyRevenue, 0, 31)
		for i := 0; ; i++ {
			day := timeutil.StartOfDay(now.AddDate(0, 0, -i))
			if day.Before(oneMonthAgo) {
				break
			}
			total, err := revenue(day, day.AddDate(0, 0, 1))
			if err != nil {
				return err
			}
			daily = append(daily, DailyRevenue{Date: day, Total: total})
		}

		serviceRows, err := q.CountServicesBetween(ctx, sqlc.CountServicesBetweenParams{
			TenantID: tenantID,
			FromTs:   ts(timeutil.StartOfDay(oneMonthAgo)),
			UntilTs:  ts(timeutil.StartOfDay(now.AddDate(0, 0, 1))),
		})
		if err != nil {
			return err
		}
		services := make([]ServiceCount, 0, len(serviceRows))
		for _, r := range serviceRows {
			name := ""
			if r.Name != nil {
				name = *r.Name
			}
			services = append(services, ServiceCount{ID: r.ID, Name: name, Count: r.Total})
		}

		// From 1 January of this year, keeping the current time of day.
		yearStart := time.Date(now.Year(), time.January, 1,
			now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
		hourRows, err := q.CountAppointmentsByHour(ctx, sqlc.CountAppointmentsByHourParams{TenantID: tenantID, FromTs: ts(yearStart)})
		if err != nil {
			return err
		}
		hours := make([]HourCount, 0, len(hourRows))
		for _, r := range hourRows {
			hours = append(hours, HourCount{Hour: r.Hour, Count: r.Total})
		}

		out = &Dashboard{
			AppointmentsToday: today, AppointmentsYesterday: yesterday,
			RevenueThisMonth: thisMonth, RevenueLastMonth: lastMonth,
			CustomerCount: customers, DailyRevenue: daily,
			Services: services, Hours: hours,
		}
		return nil
	})
	return out, err
}
