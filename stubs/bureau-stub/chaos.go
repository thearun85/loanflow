package main

import (
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
)

type ChaosConfig struct {
	LatencyMs int     `json:"latency_ms"`
	ErrorRate float64 `json:"error_rate"`
	Down      bool    `json:"down"`
}

type Chaos struct {
	mu  sync.RWMutex
	cfg ChaosConfig
}

func (c *Chaos) Get() ChaosConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cfg
}

func (c *Chaos) Set(cfg ChaosConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg = cfg
}

// Apply injects the configured faults. It returns false if the request
// was failed here and a response (if any) has already been written.
func (c *Chaos) Apply(w http.ResponseWriter, r *http.Request) bool {
	cfg := c.Get()

	if cfg.Down {
		http.Error(w, "service unavailable (chaos: down)", http.StatusServiceUnavailable)
		return false
	}

	if cfg.LatencyMs > 0 {
		select {
		case <-time.After(time.Duration(cfg.LatencyMs) * time.Millisecond):
		case <-r.Context().Done():
			slog.Info("client gave up during injected latency")
			return false
		}
	}

	if cfg.ErrorRate > 0 && rand.Float64() < cfg.ErrorRate {
		http.Error(w, "service unavailable (chaos: error_rate)", http.StatusServiceUnavailable)
		return false
	}

	return true
}

func (c *Chaos) handleGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, c.Get())
}

func (c *Chaos) handleSet(w http.ResponseWriter, r *http.Request) {
	var cfg ChaosConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if cfg.LatencyMs < 0 || cfg.ErrorRate < 0 || cfg.ErrorRate > 1 {
		http.Error(w, "latency_ms must be >= 0 and error_rate between 0 and 1", http.StatusBadRequest)
		return
	}

	c.Set(cfg)
	slog.Info("chaos updated", "latency_ms", cfg.LatencyMs, "error_rate", cfg.ErrorRate, "down", cfg.Down)
	writeJSON(w, http.StatusOK, cfg)
}

func (c *Chaos) handleReset(w http.ResponseWriter, r *http.Request) {
	c.Set(ChaosConfig{})
	slog.Info("chaos reset")
	writeJSON(w, http.StatusOK, c.Get())
}
