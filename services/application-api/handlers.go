package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

type API struct {
	store *Store
}

type createApplicationRequest struct {
	ApplicantName  string `json:"applicant_name"`
	ApplicantEmail string `json:"applicant_email"`
	AmountMinor    int64  `json:"amount_minor"`
	Currency       string `json:"currency"`
	TermMonths     int    `json:"term_months"`
}

func (req createApplicationRequest) validate() []string {
	var problems []string
	if strings.TrimSpace(req.ApplicantName) == "" {
		problems = append(problems, "applicant_name is required")
	}
	if !strings.Contains(req.ApplicantEmail, "@") {
		problems = append(problems, "applicant_email must be a valid email address")
	}
	if req.AmountMinor < 100000 || req.AmountMinor > 5000000 {
		problems = append(problems, "amount_minor must be between 100000 and 5000000 (£1,000 to £50,000)")
	}
	if req.Currency != "GBP" {
		problems = append(problems, "currency must be GBP")
	}
	if req.TermMonths < 6 || req.TermMonths > 84 {
		problems = append(problems, "term_months must be between 6 and 84")
	}
	return problems
}

func (api *API) createApplication(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createApplicationRequest
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body", []string{err.Error()})
		return
	}
	if problems := req.validate(); len(problems) > 0 {
		writeError(w, http.StatusUnprocessableEntity, "validation failed", problems)
		return
	}

	app, err := api.store.CreateApplication(r.Context(), NewApplication{
		ApplicantName:  strings.TrimSpace(req.ApplicantName),
		ApplicantEmail: strings.TrimSpace(req.ApplicantEmail),
		AmountMinor:    req.AmountMinor,
		Currency:       req.Currency,
		TermMonths:     req.TermMonths,
	})
	if err != nil {
		slog.Error("create application failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	slog.Info("application created", "application_id", app.ID, "amount_minor", app.AmountMinor)
	w.Header().Set("Location", "/applications/"+app.ID)
	writeJSON(w, http.StatusCreated, app)
}

func (api *API) getApplication(w http.ResponseWriter, r *http.Request) {
	var id pgtype.UUID
	if err := id.Scan(r.PathValue("id")); err != nil {
		writeError(w, http.StatusNotFound, "application not found", nil)
		return
	}

	app, err := api.store.GetApplication(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "application not found", nil)
		return
	}
	if err != nil {
		slog.Error("get application failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	writeJSON(w, http.StatusOK, app)
}

type errorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write response failed", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string, details []string) {
	writeJSON(w, status, errorResponse{Error: msg, Details: details})
}
