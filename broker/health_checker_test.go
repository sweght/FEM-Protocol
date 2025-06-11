package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Test checkAgentConnectivity with various server responses
func TestCheckAgentConnectivity(t *testing.T) {
	hc := NewHealthChecker(time.Second, 0.8)

	// Healthy server returning 200 on /health
	healthySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer healthySrv.Close()

	if !hc.checkAgentConnectivity(healthySrv.URL) {
		t.Error("expected connectivity check to succeed")
	}

	// Server returning 503
	badSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	if hc.checkAgentConnectivity(badSrv.URL) {
		t.Error("expected connectivity check to fail with bad status")
	}
	badSrv.Close()

	// Unreachable endpoint
	if hc.checkAgentConnectivity(badSrv.URL) {
		t.Error("expected connectivity check to fail for unreachable server")
	}
}

// Test checkAgentCapabilities scoring logic using mocked servers
func TestCheckAgentCapabilities(t *testing.T) {
	hc := NewHealthChecker(time.Second, 0.8)

	// Server returning valid JSON
	okSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer okSrv.Close()
	if score := hc.checkAgentCapabilities(okSrv.URL); score != 1.0 {
		t.Errorf("expected score 1.0, got %f", score)
	}

	// Server returning invalid JSON
	invalidSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "not-json")
	}))
	if score := hc.checkAgentCapabilities(invalidSrv.URL); score != 0.7 {
		t.Errorf("expected score 0.7, got %f", score)
	}
	invalidSrv.Close()

	// Server returning non-OK status
	statusSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	if score := hc.checkAgentCapabilities(statusSrv.URL); score != 0.5 {
		t.Errorf("expected score 0.5, got %f", score)
	}
	statusSrv.Close()

	// Unreachable server
	if score := hc.checkAgentCapabilities(statusSrv.URL); score != 0.0 {
		t.Errorf("expected score 0.0, got %f", score)
	}
}

func TestCalculateTimeScore(t *testing.T) {
	hc := NewHealthChecker(time.Second, 0.8)

	cases := []struct {
		dur  time.Duration
		want float64
	}{
		{50 * time.Millisecond, 1.0},
		{300 * time.Millisecond, 0.8},
		{800 * time.Millisecond, 0.6},
		{3 * time.Second, 0.4},
		{6 * time.Second, 0.2},
	}

	for _, c := range cases {
		got := hc.calculateTimeScore(c.dur)
		if got != c.want {
			t.Errorf("duration %v: expected %v, got %v", c.dur, c.want, got)
		}
	}
}

func TestDetermineAgentStatus(t *testing.T) {
	hc := NewHealthChecker(time.Second, 0.8)

	tests := []struct {
		score float64
		want  AgentStatus
	}{
		{0.9, AgentStatusHealthy},
		{0.65, AgentStatusDegraded},
		{0.2, AgentStatusUnhealthy},
		{0.0, AgentStatusUnknown},
	}

	for _, tt := range tests {
		got := hc.determineAgentStatus(tt.score)
		if got != tt.want {
			t.Errorf("score %f: expected %s, got %s", tt.score, tt.want, got)
		}
	}
}
