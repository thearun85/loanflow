package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

type creditCheckRequest struct {
	ApplicationID  string `json:"application_id"`
	ApplicantName  string `json:"applicant_name"`
	ApplicantEmail string `json:"applicant_email"`
}

type creditCheckResponse struct {
	BureauRef     string    `json:"bureau_ref"`
	ApplicationID string    `json:"application_id"`
	Score         int       `json:"score"`
	CheckedAt     time.Time `json:"checked_at"`
}

func handleCreditCheck(chaos *Chaos) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !chaos.Apply(w, r) {
			return
		}

		var req creditCheckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.ApplicationID == "" || req.ApplicantEmail == "" {
			http.Error(w, "application_id and applicant_email are required", http.StatusBadRequest)
			return
		}

		resp := creditCheckResponse{
			BureauRef:     fmt.Sprintf("BR-%06d", rand.IntN(1000000)),
			ApplicationID: req.ApplicationID,
			Score:         scoreFor(req.ApplicantEmail),
			CheckedAt:     time.Now().UTC(),
		}

		slog.Info("credit check completed",
			"application_id", resp.ApplicationID,
			"bureau_ref", resp.BureauRef,
			"score", resp.Score,
		)
		writeJSON(w, http.StatusOK, resp)
	}
}

// scoreFor returns a score between 300 and 999 that is always the same
// for the same email, so test runs are repeatable.
func scoreFor(email string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(email)))
	return 300 + int(h.Sum32()%700)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write response failed", "error", err)
	}
}
