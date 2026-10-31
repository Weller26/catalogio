package httpapi

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var (
	chaosDelayMs atomic.Int64
	chaosErrorPercent atomic.Int64
)

func ChaosMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" || strings.HasPrefix(r.URL.Path, "/api/v1/chaos") {
			next.ServeHTTP(w, r)
			return
		}

		if delay := chaosDelayMs.Load(); delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		if errPercent := chaosErrorPercent.Load(); errPercent > 0 {
			if rand.Intn(100) < int(errPercent) {
				http.Error(w, "Chaos: Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func handleChaosSlow(w http.ResponseWriter, r *http.Request) {
	msStr := r.URL.Query().Get("ms")
	ms, err := strconv.ParseInt(msStr, 10, 64)
	if err != nil || ms < 0 {
		http.Error(w, "invalid 'ms' parameter, expected non-negative integer", http.StatusBadRequest)
		return
	}

	chaosDelayMs.Store(ms)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Chaos delay set to %d ms\n", ms)
}

func handleChaosFail(w http.ResponseWriter, r *http.Request) {
	pctStr := r.URL.Query().Get("percent")
	pct, err := strconv.ParseInt(pctStr, 10, 64)
	if err != nil || pct < 0 || pct > 100 {
		http.Error(w, "invalid 'percent' parameter, expected 0-100", http.StatusBadRequest)
		return
	}

	chaosErrorPercent.Store(pct)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Chaos error rate set to %d%%\n", pct)
}

func handleChaosReset(w http.ResponseWriter, r *http.Request) {
	chaosDelayMs.Store(0)
	chaosErrorPercent.Store(0)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Chaos settings reset to normal")
}

func handleChaosStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{
		"delay_ms":      chaosDelayMs.Load(),
		"error_percent": chaosErrorPercent.Load(),
	})
}