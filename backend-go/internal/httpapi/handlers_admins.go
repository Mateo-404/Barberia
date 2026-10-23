package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/apperr"
	"github.com/Mateo-404/Barberia/backend-go/internal/auth"
	"github.com/Mateo-404/Barberia/backend-go/internal/modules/admins"
)

// adminDTO never includes the password hash (ADR-004).
type adminDTO struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

type loginDTO struct {
	adminDTO
	Token string `json:"token"`
}

type adminCreateReq struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email"`
	Password  *string `json:"password"`
	Role      *string `json:"role"`
}

type adminPatchReq struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email"`
	Password  *string `json:"password"`
	Role      *string `json:"role"`
}

type loginReq struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

func toAdminDTO(a admins.Admin) adminDTO {
	return adminDTO{ID: a.ID, FirstName: a.FirstName, LastName: a.LastName, Email: a.Email, Role: a.Role}
}

func validRole(v *validator, role *string) {
	if role != nil && *role != auth.RoleOwner && *role != auth.RoleAdmin && *role != auth.RoleStaff {
		v.add("role", "invalid_value")
	}
}

func (s *Server) handleAdminList(w http.ResponseWriter, r *http.Request) {
	list, err := s.admins.List(r.Context(), tenantID(r))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(list, toAdminDTO))
}

func (s *Server) handleAdminPage(w http.ResponseWriter, r *http.Request) {
	number, size, ok := pageParams(w, r)
	if !ok {
		return
	}
	list, total, err := s.admins.Page(r.Context(), tenantID(r), number, size)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newPage(mapSlice(list, toAdminDTO), total, number, size))
}

func (s *Server) handleAdminCreate(w http.ResponseWriter, r *http.Request) {
	var req adminCreateReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	var v validator
	v.required("firstName", req.FirstName)
	v.required("lastName", req.LastName)
	v.maxLen("firstName", req.FirstName, maxName)
	v.maxLen("lastName", req.LastName, maxName)
	if v.required("email", req.Email) {
		v.email("email", req.Email)
		v.maxLen("email", req.Email, maxEmail)
	}
	if v.required("password", req.Password) {
		v.minLen("password", req.Password, minPassword)
	}
	validRole(&v, req.Role)
	if !v.ok(w, r) {
		return
	}
	in := admins.Input{FirstName: *req.FirstName, LastName: *req.LastName, Email: *req.Email, Password: *req.Password}
	if req.Role != nil {
		in.Role = *req.Role
	}
	a, err := s.admins.Create(r.Context(), tenantID(r), in)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAdminDTO(*a))
}

func (s *Server) handleAdminGet(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	a, err := s.admins.Get(r.Context(), tenantID(r), id)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if a == nil {
		resourceNotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, toAdminDTO(*a))
}

func (s *Server) handleAdminPatch(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var req adminPatchReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	var v validator
	v.maxLen("firstName", req.FirstName, maxName)
	v.maxLen("lastName", req.LastName, maxName)
	v.email("email", req.Email)
	v.maxLen("email", req.Email, maxEmail)
	v.minLen("password", req.Password, minPassword)
	validRole(&v, req.Role)
	if !v.ok(w, r) {
		return
	}
	a, err := s.admins.Update(r.Context(), tenantID(r), id, admins.Patch{
		FirstName: req.FirstName, LastName: req.LastName, Email: req.Email, Password: req.Password, Role: req.Role,
	})
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if a == nil {
		resourceNotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, toAdminDTO(*a))
}

func (s *Server) handleAdminDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	deleted, err := s.admins.Delete(r.Context(), tenantID(r), id)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if !deleted {
		resourceNotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	var v validator
	v.required("email", req.Email)
	if v.required("password", req.Password) {
		v.minLen("password", req.Password, minPassword)
	}
	if !v.ok(w, r) {
		return
	}

	ip := s.clientIP(r)
	if !s.loginLimiter.Allow(ip) {
		rateLimited(w, r)
		return
	}
	admin, token, requires2FA, err := s.admins.Login(r.Context(), tenantID(r), *req.Email, *req.Password)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if requires2FA {
		s.loginLimiter.Reset(ip)
		writeJSON(w, http.StatusOK, map[string]any{"requires2fa": true, "tempToken": token})
		return
	}
	s.loginLimiter.Reset(ip)
	writeJSON(w, http.StatusOK, loginDTO{adminDTO: toAdminDTO(admin), Token: token})
}

type googleStartReq struct {
	Redirect *string `json:"redirect"`
}

type googleCallbackReq struct {
	Code  *string `json:"code"`
	State *string `json:"state"`
}

type totpCodeReq struct {
	Code *string `json:"code"`
}

type totpChallengeReq struct {
	TempToken *string `json:"tempToken"`
	Code      *string `json:"code"`
}

func (s *Server) handleAuthProviders(w http.ResponseWriter, r *http.Request) {
	gc := auth.GoogleConfig{ClientID: s.cfg.GoogleClientID, ClientSecret: s.cfg.GoogleClientSecret, RedirectURL: s.cfg.GoogleRedirectURL}
	writeJSON(w, http.StatusOK, map[string]any{"google": gc.Enabled()})
}

func (s *Server) handleGoogleStart(w http.ResponseWriter, r *http.Request) {
	gc := auth.GoogleConfig{ClientID: s.cfg.GoogleClientID, ClientSecret: s.cfg.GoogleClientSecret, RedirectURL: s.cfg.GoogleRedirectURL}
	if !gc.Enabled() {
		s.writeError(w, r, apperr.New(apperr.KindNotFound, "auth.google_disabled", nil, "google login disabled"))
		return
	}
	var req googleStartReq
	_ = decodeJSON(w, r, &req, s.cfg.MaxRequestBody)
	redirect := "/admin"
	if req.Redirect != nil && *req.Redirect != "" {
		redirect = *req.Redirect
	}
	// state carries the redirect target, signed.
	payload := fmt.Sprintf("%d|%s", timeNowUnix(), redirect)
	state := auth.SignState([]byte(s.cfg.JWTSecret), payload)
	url := gc.AuthURL(state)
	writeJSON(w, http.StatusOK, map[string]any{"url": url, "state": state})
}

func (s *Server) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	var req googleCallbackReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	if req.Code == nil || *req.Code == "" || req.State == nil || *req.State == "" {
		badRequest(w, r)
		return
	}
	payload, ok := auth.VerifyState([]byte(s.cfg.JWTSecret), *req.State)
	if !ok {
		badRequest(w, r)
		return
	}
	// payload format: timestamp|redirect (redirect unused server-side)
	_ = payload

	ip := s.clientIP(r)
	if !s.loginLimiter.Allow(ip) {
		rateLimited(w, r)
		return
	}

	gc := auth.GoogleConfig{ClientID: s.cfg.GoogleClientID, ClientSecret: s.cfg.GoogleClientSecret, RedirectURL: s.cfg.GoogleRedirectURL}
	idToken, err := gc.ExchangeCode(*req.Code)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	gu, err := gc.VerifyIDToken(idToken)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	if !gu.EmailVerified {
		badRequest(w, r)
		return
	}

	admin, token, requires2FA, err := s.admins.LoginWithGoogle(r.Context(), tenantID(r), gu)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	s.loginLimiter.Reset(ip)
	if requires2FA {
		writeJSON(w, http.StatusOK, map[string]any{"requires2fa": true, "tempToken": token})
		return
	}
	writeJSON(w, http.StatusOK, loginDTO{adminDTO: toAdminDTO(admin), Token: token})
}

func (s *Server) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(ctxKeyClaims).(*auth.Claims)
	adminID, _ := strconv.ParseInt(claims.Subject, 10, 64)
	secret, uri, err := s.admins.SetupTOTP(r.Context(), tenantID(r), adminID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"secret": secret, "otpauthUrl": uri})
}

func (s *Server) handleTOTPVerify(w http.ResponseWriter, r *http.Request) {
	var req totpCodeReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	if req.Code == nil || *req.Code == "" {
		badRequest(w, r)
		return
	}
	claims, _ := r.Context().Value(ctxKeyClaims).(*auth.Claims)
	adminID, _ := strconv.ParseInt(claims.Subject, 10, 64)
	if err := s.admins.VerifyTOTP(r.Context(), tenantID(r), adminID, *req.Code); err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"enabled": true})
}

func (s *Server) handleTOTPDisable(w http.ResponseWriter, r *http.Request) {
	var req totpCodeReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	if req.Code == nil || *req.Code == "" {
		badRequest(w, r)
		return
	}
	claims, _ := r.Context().Value(ctxKeyClaims).(*auth.Claims)
	adminID, _ := strconv.ParseInt(claims.Subject, 10, 64)
	if err := s.admins.DisableTOTP(r.Context(), tenantID(r), adminID, *req.Code); err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
}

func (s *Server) handleTOTPChallenge(w http.ResponseWriter, r *http.Request) {
	var req totpChallengeReq
	if !decodeJSON(w, r, &req, s.cfg.MaxRequestBody) {
		return
	}
	if req.TempToken == nil || *req.TempToken == "" || req.Code == nil || *req.Code == "" {
		badRequest(w, r)
		return
	}
	claims, err := s.jwt.Parse(*req.TempToken)
	if err != nil || claims.Scope != "2fa_pending" {
		invalidToken(w, r)
		return
	}
	ip := s.clientIP(r)
	if !s.loginLimiter.Allow(ip) {
		rateLimited(w, r)
		return
	}
	admin, token, err := s.admins.ChallengeTOTP(r.Context(), claims, *req.Code)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	s.loginLimiter.Reset(ip)
	writeJSON(w, http.StatusOK, loginDTO{adminDTO: toAdminDTO(admin), Token: token})
}

func timeNowUnix() int64 { return time.Now().Unix() }
