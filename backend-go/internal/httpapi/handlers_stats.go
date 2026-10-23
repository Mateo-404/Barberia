package httpapi

import (
	"net/http"

	"github.com/Mateo-404/Barberia/backend-go/internal/modules/stats"
	"github.com/Mateo-404/Barberia/backend-go/internal/timeutil"
)

type dashboardDTO struct {
	AppointmentsToday     int64             `json:"appointmentsToday"`
	AppointmentsYesterday int64             `json:"appointmentsYesterday"`
	RevenueThisMonth      float64           `json:"revenueThisMonth"`
	RevenueLastMonth      float64           `json:"revenueLastMonth"`
	CustomerCount         int64             `json:"customerCount"`
	DailyRevenue          []dailyRevenueDTO `json:"dailyRevenue"`
	Services              []serviceCountDTO `json:"services"`
	Hours                 []hourCountDTO    `json:"hours"`
}

type dailyRevenueDTO struct {
	Date  string  `json:"date"`
	Total float64 `json:"total"`
}

type serviceCountDTO struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type hourCountDTO struct {
	Hour  int32 `json:"hour"`
	Count int64 `json:"count"`
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	d, err := s.stats.Dashboard(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dashboardDTO{
		AppointmentsToday:     d.AppointmentsToday,
		AppointmentsYesterday: d.AppointmentsYesterday,
		RevenueThisMonth:      d.RevenueThisMonth,
		RevenueLastMonth:      d.RevenueLastMonth,
		CustomerCount:         d.CustomerCount,
		DailyRevenue: mapSlice(d.DailyRevenue, func(x stats.DailyRevenue) dailyRevenueDTO {
			return dailyRevenueDTO{Date: timeutil.FormatDateTime(x.Date), Total: x.Total}
		}),
		Services: mapSlice(d.Services, func(x stats.ServiceCount) serviceCountDTO {
			return serviceCountDTO(x)
		}),
		Hours: mapSlice(d.Hours, func(x stats.HourCount) hourCountDTO {
			return hourCountDTO(x)
		}),
	})
}
