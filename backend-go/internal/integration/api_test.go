//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/admins"
)

func TestBookingFlowEndToEnd(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)

	var services []map[string]any
	e.call(t, http.MethodGet, "/api/services", "", "").expect(200, "list services").json(&services)
	if len(services) != 0 {
		t.Fatalf("services = %v, want none", services)
	}
	serviceID := e.createService(t, token, "Cut + beard", 7000)

	at := futureAt(3, 14, 0)
	var created appointment
	e.call(t, http.MethodPost, "/api/appointments", bookingBody(at, serviceID, "1150000001", "Juan", "Perez"), "").
		expect(201, "book").json(&created)
	if created.ServiceName == nil || *created.ServiceName != "Cut + beard" || created.ServicePrice != 7000 {
		t.Fatalf("created = %+v", created)
	}

	// Same slot again → 409 appointment.slot_taken.
	if code := e.call(t, http.MethodPost, "/api/appointments", bookingBody(at, serviceID, "1150000002", "Ana", "Paz"), "").
		expect(409, "double booking").problemCode(); code != "appointment.slot_taken" {
		t.Fatalf("code = %q", code)
	}

	// Booked slots are listed without seconds.
	var slots []string
	e.call(t, http.MethodGet, "/api/appointments/booked-slots", "", "").expect(200, "booked slots").json(&slots)
	if len(slots) != 1 || slots[0] != at.Format("2006-01-02T15:04") {
		t.Fatalf("slots = %v", slots)
	}

	path := fmt.Sprintf("/api/appointments/%d", created.ID)
	// A service object without data returns the in-memory result, unrefreshed.
	var patched appointment
	e.call(t, http.MethodPatch, path, fmt.Sprintf(`{"service":{"id":%d}}`, serviceID), token).expect(200, "patch").json(&patched)
	if patched.ServiceName != nil || patched.ServicePrice != 0 {
		t.Fatalf("patch should not refresh associations: %+v", patched)
	}
	var fetched appointment
	e.call(t, http.MethodGet, path, "", token).expect(200, "get").json(&fetched)
	if fetched.ServiceName == nil || *fetched.ServiceName != "Cut + beard" {
		t.Fatalf("get after patch = %+v", fetched)
	}

	if code := e.call(t, http.MethodPatch, path, `{"service":{"id":999}}`, token).expect(404, "unknown service").problemCode(); code != "service.not_found" {
		t.Fatalf("code = %q", code)
	}
	if code := e.call(t, http.MethodGet, "/api/appointments/99999", "", token).expect(404, "missing").problemCode(); code != "resource.not_found" {
		t.Fatalf("code = %q", code)
	}

	var page struct {
		Items []appointment `json:"items"`
		Total int64         `json:"total"`
	}
	e.call(t, http.MethodGet, "/api/appointments/paged?page=0&size=10", "", token).expect(200, "page").json(&page)
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("page = %+v", page)
	}

	if code := e.call(t, http.MethodPost, "/api/auth/login", `{"email":"admin@admin.com","password":"wrong-password"}`, "").
		expect(401, "bad login").problemCode(); code != "auth.invalid_credentials" {
		t.Fatalf("code = %q", code)
	}
}

func TestBookingRules(t *testing.T) {
	e := newEnv(t)
	serviceID := e.createService(t, e.ownerToken(t), "Cut", 5000)
	day := futureAt(5, 0, 0)

	cases := []struct {
		at, code string
		status   int
	}{
		{"13:00", "", 201},
		{"13:30", "", 201},
		{"19:30", "", 201}, // the last slot ends at closing time
		{"20:00", "appointment.outside_hours", 400},
		{"12:30", "appointment.outside_hours", 400},
		{"20:30", "appointment.outside_hours", 400},
		{"14:15", "appointment.invalid_interval", 400},
	}
	for i, c := range cases {
		var h, m int
		_, _ = fmt.Sscanf(c.at, "%d:%d", &h, &m)
		at := time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, time.Local)
		res := e.call(t, http.MethodPost, "/api/appointments", bookingBody(at, serviceID, fmt.Sprintf("11530000%02d", i), "N", "A"), "").
			expect(c.status, c.at)
		if c.code != "" && res.problemCode() != c.code {
			t.Fatalf("%s: code = %q, want %q", c.at, res.problemCode(), c.code)
		}
	}

	past := time.Now().AddDate(0, 0, -1)
	if code := e.call(t, http.MethodPost, "/api/appointments", bookingBody(past, serviceID, "1150000000", "V", "P"), "").
		expect(400, "past date").problemCode(); code != "appointment.in_the_past" {
		t.Fatalf("code = %q", code)
	}

	if code := e.call(t, http.MethodPost, "/api/appointments", bookingBody(futureAt(6, 14, 0), 9999, "1150000000", "V", "P"), "").
		expect(404, "unknown service").problemCode(); code != "service.not_found" {
		t.Fatalf("code = %q", code)
	}
}

// A booking with a known phone reuses the customer and does not overwrite it.
func TestBookingReusesExistingCustomer(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	serviceID := e.createService(t, token, "Cut", 5000)

	e.call(t, http.MethodPost, "/api/appointments",
		fmt.Sprintf(`{"startsAt":%q,"serviceId":%d,"customerPhone":"1154444444","customerFirstName":"Original","customerLastName":"Name","customerEmail":"original@example.com"}`,
			futureAt(7, 14, 0).Format(dateLayout), serviceID), "").expect(201, "first booking")
	e.call(t, http.MethodPost, "/api/appointments", bookingBody(futureAt(8, 14, 0), serviceID, "1154444444", "Other", "Ignored"), "").
		expect(201, "second booking")

	var c struct {
		FirstName string `json:"firstName"`
		Email     string `json:"email"`
	}
	e.call(t, http.MethodGet, "/api/customers/1154444444", "", token).expect(200, "get customer").json(&c)
	if c.FirstName != "Original" || c.Email != "original@example.com" {
		t.Fatalf("the customer was overwritten: %+v", c)
	}
}

// The unique index also rejects double bookings in the past (ADR-006).
func TestPastDuplicateRejectedByIndex(t *testing.T) {
	e := newEnv(t)
	serviceID := e.createService(t, e.ownerToken(t), "Cut", 5000)
	e.insertCustomer(t, e.tenantID, "1150000001")
	ctx := context.Background()

	past := futureAt(-7, 15, 0)
	insert := func(at time.Time) int64 {
		var id int64
		if err := e.pool.QueryRow(ctx, `INSERT INTO appointments (starts_at, customer_phone, service_id, tenant_id)
			VALUES ($1, '1150000001', $2, $3) RETURNING id`, at, serviceID, e.tenantID).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	insert(past)
	other := insert(past.AddDate(0, 0, 1))
	if code := e.call(t, http.MethodPatch, fmt.Sprintf("/api/appointments/%d", other),
		fmt.Sprintf(`{"startsAt":%q}`, past.Format(dateLayout)), e.ownerToken(t)).
		expect(409, "move onto a past slot").problemCode(); code != "appointment.slot_taken" {
		t.Fatalf("code = %q", code)
	}
}

func TestConcurrentDoubleBooking(t *testing.T) {
	e := newEnv(t)
	serviceID := e.createService(t, e.ownerToken(t), "Cut", 5000)
	at := futureAt(4, 16, 0)
	body := bookingBody(at, serviceID, "1155555555", "Ana", "Gomez")

	const n = 20
	codes := make([]int, n)
	problems := make([]string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res := e.call(t, http.MethodPost, "/api/appointments", body, "")
			codes[i] = res.Code
			if res.Code == 409 {
				problems[i] = res.problemCode()
			}
		}(i)
	}
	wg.Wait()

	created := 0
	for i, c := range codes {
		switch c {
		case 201:
			created++
		case 409:
			if problems[i] != "appointment.slot_taken" {
				t.Fatalf("409 with code %q", problems[i])
			}
		default:
			t.Fatalf("unexpected status %d (all: %v)", c, codes)
		}
	}
	if created != 1 {
		t.Fatalf("created = %d, want exactly 1 (codes %v)", created, codes)
	}
	var rows int
	_ = e.pool.QueryRow(context.Background(), `SELECT count(*) FROM appointments WHERE starts_at = $1`, at).Scan(&rows)
	if rows != 1 {
		t.Fatalf("rows = %d, want 1", rows)
	}
}

func TestCustomerCRUD(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)

	e.call(t, http.MethodPost, "/api/customers", `{"phone":"1151111111","firstName":"Ana","lastName":"Gomez","email":"ana@example.com"}`, token).
		expect(201, "create")
	if code := e.call(t, http.MethodPost, "/api/customers", `{"phone":"1151111111","firstName":"Other","lastName":"Name"}`, token).
		expect(409, "duplicate").problemCode(); code != "customer.already_exists" {
		t.Fatalf("code = %q", code)
	}

	var c struct {
		Phone, FirstName, LastName string
		Email                      *string
	}
	e.call(t, http.MethodPatch, "/api/customers/1151111111", `{"firstName":"Ana Maria"}`, token).expect(200, "patch").json(&c)
	if c.FirstName != "Ana Maria" || c.LastName != "Gomez" || c.Email == nil || *c.Email != "ana@example.com" {
		t.Fatalf("partial patch lost fields: %+v", c)
	}
	// The phone is the key and cannot be changed through PATCH.
	e.call(t, http.MethodPatch, "/api/customers/1151111111", `{"phone":"999999999","firstName":"Ana M."}`, token).expect(200, "patch").json(&c)
	if c.Phone != "1151111111" {
		t.Fatalf("the phone changed: %+v", c)
	}
	e.call(t, http.MethodGet, "/api/customers/999999999", "", token).expect(404, "old key")

	var page struct {
		Items []any `json:"items"`
		Total int64 `json:"total"`
	}
	e.call(t, http.MethodGet, "/api/customers/paged", "", token).expect(200, "page").json(&page)
	if page.Total != 1 {
		t.Fatalf("page = %+v", page)
	}

	e.call(t, http.MethodDelete, "/api/customers/1151111111", "", token).expect(204, "delete")
	e.call(t, http.MethodDelete, "/api/customers/1151111111", "", token).expect(404, "delete again")
}

func TestCustomerWithAppointmentsCannotBeDeleted(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	serviceID := e.createService(t, token, "Cut", 5000)
	e.call(t, http.MethodPost, "/api/appointments", bookingBody(futureAt(3, 14, 0), serviceID, "1152222222", "Luis", "Sosa"), "").expect(201, "book")
	if code := e.call(t, http.MethodDelete, "/api/customers/1152222222", "", token).expect(409, "delete in use").problemCode(); code != "resource.in_use" {
		t.Fatalf("code = %q", code)
	}
}

func TestServiceCRUD(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	id := e.createService(t, token, "Beard", 2500)
	path := fmt.Sprintf("/api/services/%d", id)

	var s struct {
		Name  string  `json:"name"`
		Price float32 `json:"price"`
	}
	e.call(t, http.MethodGet, path, "", "").expect(200, "public get").json(&s)
	if s.Price != 2500 {
		t.Fatalf("service = %+v", s)
	}
	e.call(t, http.MethodPatch, path, `{"price":3000}`, token).expect(200, "patch").json(&s)
	if s.Price != 3000 || s.Name != "Beard" {
		t.Fatalf("patched = %+v", s)
	}
	e.call(t, http.MethodPost, "/api/services", `{"name":"","price":1000}`, token).expect(400, "blank name")
	e.call(t, http.MethodPost, "/api/services", `{"name":"X","price":-1}`, token).expect(400, "negative price")
	e.call(t, http.MethodDelete, path, "", token).expect(204, "delete")
	e.call(t, http.MethodGet, path, "", "").expect(404, "deleted")
}

func TestAdminCRUDAndPasswordRotation(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)

	var created struct {
		ID   int64  `json:"id"`
		Role string `json:"role"`
	}
	res := e.call(t, http.MethodPost, "/api/admins",
		`{"firstName":"Carla","lastName":"Ruiz","email":"carla@example.com","password":"password123"}`, token).expect(201, "create")
	res.json(&created)
	if strings.Contains(res.Body.String(), "argon2") || strings.Contains(res.Body.String(), "password") {
		t.Fatalf("the hash leaked: %s", res.Body.String())
	}
	if created.Role != auth.RoleStaff {
		t.Fatalf("new admins default to staff, got %q", created.Role)
	}
	if code := e.call(t, http.MethodPost, "/api/admins",
		`{"firstName":"Other","lastName":"Person","email":"carla@example.com","password":"password123"}`, token).
		expect(409, "duplicate").problemCode(); code != "admin.already_exists" {
		t.Fatalf("code = %q", code)
	}

	staffToken := e.login(t, "carla@example.com", "password123")
	e.call(t, http.MethodGet, "/api/admins", "", staffToken).expect(403, "staff listing admins")

	path := fmt.Sprintf("/api/admins/%d", created.ID)
	e.call(t, http.MethodPatch, path, `{"password":"newPassword123"}`, token).expect(200, "rotate password")
	e.call(t, http.MethodPost, "/api/auth/login", `{"email":"carla@example.com","password":"password123"}`, "").expect(401, "old password")
	e.login(t, "carla@example.com", "newPassword123")
	e.call(t, http.MethodPost, "/api/auth/login", `{"email":"carla@example.com","password":"short"}`, "").expect(400, "short password")

	e.call(t, http.MethodGet, "/api/admins/paged?size=10", "", token).expect(200, "page")
	e.call(t, http.MethodDelete, path, "", token).expect(204, "delete")
}

func TestDashboardWithData(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	cut := e.createService(t, token, "Cut", 5000)
	beard := e.createService(t, token, "Beard", 2500)
	e.insertCustomer(t, e.tenantID, "1150000001")

	now := time.Now()
	today14 := time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, time.Local)
	for _, a := range []struct {
		at      time.Time
		service int64
	}{{today14, cut}, {today14.Add(2 * time.Hour), cut}, {today14.AddDate(0, 0, -1).Add(time.Hour), beard}} {
		if _, err := e.pool.Exec(context.Background(),
			`INSERT INTO appointments (starts_at, customer_phone, service_id, tenant_id) VALUES ($1, '1150000001', $2, $3)`,
			a.at, a.service, e.tenantID); err != nil {
			t.Fatal(err)
		}
	}
	thisMonth, lastMonth := 10000.0, 0.0
	if now.Day() != 1 {
		thisMonth += 2500
	} else {
		lastMonth = 2500
	}

	var d struct {
		AppointmentsToday     int64   `json:"appointmentsToday"`
		AppointmentsYesterday int64   `json:"appointmentsYesterday"`
		RevenueThisMonth      float64 `json:"revenueThisMonth"`
		RevenueLastMonth      float64 `json:"revenueLastMonth"`
		CustomerCount         int64   `json:"customerCount"`
		DailyRevenue          []struct {
			Date  string  `json:"date"`
			Total float64 `json:"total"`
		} `json:"dailyRevenue"`
		Services []struct {
			Name  string `json:"name"`
			Count int64  `json:"count"`
		} `json:"services"`
		Hours []struct {
			Hour  int32 `json:"hour"`
			Count int64 `json:"count"`
		} `json:"hours"`
	}
	e.call(t, http.MethodGet, "/api/stats/dashboard", "", token).expect(200, "dashboard").json(&d)
	if d.AppointmentsToday != 2 || d.AppointmentsYesterday != 1 || d.RevenueThisMonth != thisMonth ||
		d.RevenueLastMonth != lastMonth || d.CustomerCount != 1 {
		t.Fatalf("dashboard = %+v", d)
	}
	if len(d.DailyRevenue) == 0 || !strings.HasPrefix(d.DailyRevenue[0].Date, today14.Format("2006-01-02")) || d.DailyRevenue[0].Total != 10000 {
		t.Fatalf("daily revenue should start today: %+v", d.DailyRevenue[:1])
	}
	if len(d.Services) == 0 || d.Services[0].Name != "Cut" || d.Services[0].Count != 2 {
		t.Fatalf("services = %+v", d.Services)
	}
	if len(d.Hours) == 0 {
		t.Fatalf("hours empty")
	}
}

func TestTokenWithoutTenantIsRejected(t *testing.T) {
	e := newEnv(t)
	m, _ := auth.NewJWTManager(jwtSecret, time.Hour)
	token, err := m.Issue(1, seedEmail, 0, auth.RoleOwner)
	if err != nil {
		t.Fatal(err)
	}
	e.call(t, http.MethodGet, "/api/admins", "", token).expect(401, "token without tenant")
}

func TestSeedIsIdempotent(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	m, _ := auth.NewJWTManager(jwtSecret, time.Hour)
	seeder := admins.New(e.pool, e.q, m, quietLogger(), jwtSecret)
	if err := seeder.Seed(ctx, e.tenantID, seedEmail, seedPassword, "Admin", "Default"); err != nil {
		t.Fatal(err)
	}
	var total int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM admins`).Scan(&total); err != nil || total != 1 {
		t.Fatalf("seed duplicated: total=%d err=%v", total, err)
	}
	if err := seeder.Seed(ctx, 999, "x@y.com", "short", "X", "Y"); err == nil || !strings.Contains(err.Error(), "8 characters") {
		t.Fatalf("seed with a short password: %v", err)
	}
}

func TestTenantIsolation(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	e.createService(t, token, "Cut default", 5000)
	tenant2 := e.insertTenant(t, "salon2", "salon2.example.com")
	if _, err := e.pool.Exec(context.Background(),
		`INSERT INTO services (name, price, tenant_id) VALUES ('Cut salon2', 9000, $1)`, tenant2); err != nil {
		t.Fatal(err)
	}

	if body := e.call(t, http.MethodGet, "/api/services", "", "").Body.String(); !strings.Contains(body, "Cut default") || strings.Contains(body, "Cut salon2") {
		t.Fatalf("default tenant sees: %s", body)
	}
	if body := e.call(t, http.MethodGet, "/api/services", "", "", "salon2.example.com").Body.String(); !strings.Contains(body, "Cut salon2") || strings.Contains(body, "Cut default") {
		t.Fatalf("salon2 sees: %s", body)
	}
	// A token of the default tenant used against salon2's host → 403.
	e.call(t, http.MethodGet, "/api/services", "", token, "salon2.example.com").expect(403, "cross-tenant token")
}

// A tenant cannot link another tenant's service to its appointment.
func TestCrossTenantWritesAreRejected(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	own := e.createService(t, token, "Own cut", 5000)
	var created appointment
	e.call(t, http.MethodPost, "/api/appointments", bookingBody(futureAt(4, 14, 0), own, "1160000001", "Ana", "Luz"), "").
		expect(201, "book").json(&created)

	tenant3 := e.insertTenant(t, "salon3", "salon3.example.com")
	var foreign int64
	if err := e.pool.QueryRow(context.Background(),
		`INSERT INTO services (name, price, tenant_id) VALUES ('Foreign cut', 9999, $1) RETURNING id`, tenant3).Scan(&foreign); err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/api/appointments/%d", created.ID)
	e.call(t, http.MethodPatch, path, fmt.Sprintf(`{"service":{"id":%d}}`, foreign), token).expect(404, "foreign service")

	var after appointment
	e.call(t, http.MethodGet, path, "", token).expect(200, "get").json(&after)
	if after.ServiceID == nil || *after.ServiceID != own {
		t.Fatalf("the appointment changed service: %+v", after)
	}
}

func TestBookedSlotsAreTenantScoped(t *testing.T) {
	e := newEnv(t)
	ownService := e.createService(t, e.ownerToken(t), "Cut", 5000)
	tenant2 := e.insertTenant(t, "salon2", "salon2.example.com")
	ctx := context.Background()
	var service2 int64
	if err := e.pool.QueryRow(ctx, `INSERT INTO services (name, price, tenant_id) VALUES ('Cut2', 9000, $1) RETURNING id`, tenant2).Scan(&service2); err != nil {
		t.Fatal(err)
	}
	e.insertCustomer(t, tenant2, "1150000001")
	e.insertCustomer(t, e.tenantID, "1150000002")

	// The same time in both tenants is allowed.
	at := futureAt(12, 14, 0)
	for _, row := range []struct {
		phone             string
		service, tenantID int64
	}{{"1150000001", service2, tenant2}, {"1150000002", ownService, e.tenantID}} {
		if _, err := e.pool.Exec(ctx, `INSERT INTO appointments (starts_at, customer_phone, service_id, tenant_id) VALUES ($1, $2, $3, $4)`,
			at, row.phone, row.service, row.tenantID); err != nil {
			t.Fatal(err)
		}
	}
	for _, host := range []string{"", "salon2.example.com"} {
		var slots []string
		e.call(t, http.MethodGet, "/api/appointments/booked-slots", "", "", host).expect(200, "slots "+host).json(&slots)
		if len(slots) != 1 {
			t.Fatalf("host %q sees %v, want exactly its own slot", host, slots)
		}
	}
}

// The owner's business hours drive what can be booked.
func TestBookingFollowsTheBusinessHours(t *testing.T) {
	e := newEnv(t)
	token := e.ownerToken(t)
	serviceID := e.createService(t, token, "Cut", 5000)
	e.call(t, http.MethodPut, "/api/admin/business",
		`{"whatsapp":"","instagram":"","address":"","currency":"ARS","slot":60,"notice":0,"open":"09:00","close":"12:00","tz":"America/Argentina/Buenos_Aires"}`,
		token).expect(200, "save business")

	day := futureAt(6, 0, 0)
	book := func(hour, minute int, phone string) response {
		at := time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, time.Local)
		return e.call(t, http.MethodPost, "/api/appointments", bookingBody(at, serviceID, phone, "N", "A"), "")
	}
	book(9, 0, "1160000001").expect(201, "09:00 is the first slot")
	book(11, 0, "1160000002").expect(201, "11:00 is the last slot")
	if code := book(12, 0, "1160000003").expect(400, "12:00").problemCode(); code != "appointment.outside_hours" {
		t.Fatalf("code = %q", code)
	}
	if code := book(9, 30, "1160000004").expect(400, "09:30").problemCode(); code != "appointment.invalid_interval" {
		t.Fatalf("code = %q", code)
	}
	// The old hardcoded 13:00 hour is now closed.
	if code := book(14, 0, "1160000005").expect(400, "14:00").problemCode(); code != "appointment.outside_hours" {
		t.Fatalf("code = %q", code)
	}
}
