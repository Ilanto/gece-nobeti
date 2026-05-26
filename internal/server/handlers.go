package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/burak/linux-dashboard/internal/ai"
	"github.com/burak/linux-dashboard/internal/collector"
	"github.com/burak/linux-dashboard/internal/controller"
	"github.com/burak/linux-dashboard/internal/storage"
)

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// parsePIDParam extracts the :pid path parameter as uint32.
func parsePIDParam(w http.ResponseWriter, r *http.Request) (uint32, bool) {
	pidStr := chi.URLParam(r, "pid")
	pid, err := strconv.ParseUint(pidStr, 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_pid", "pid must be a positive integer")
		return 0, false
	}
	return uint32(pid), true
}

// ----- new data handlers -----

func handleHost(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, col.LatestHost())
	}
}

func handleCores(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, col.LatestCores())
	}
}

func handleSensors(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, col.LatestSensors())
	}
}

func handleSyslog(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"entries": col.FetchSyslog(200)})
	}
}

func handleConnections(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, col.FetchConnections())
	}
}

// ----- system handlers -----

func handleSystem(col *collector.Manager, store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := col.LatestSnapshot()
		if snap == nil {
			snap = store.Latest() // fallback
		}
		if snap == nil {
			writeError(w, http.StatusServiceUnavailable, "no_data", "no snapshot yet")
			return
		}
		writeJSON(w, http.StatusOK, snap)
	}
}

func handleCPU(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := col.LatestSnapshot()
		if snap == nil {
			writeError(w, http.StatusServiceUnavailable, "no_data", "no snapshot yet")
			return
		}
		writeJSON(w, http.StatusOK, snap.CPU)
	}
}

func handleMemory(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := col.LatestSnapshot()
		if snap == nil {
			writeError(w, http.StatusServiceUnavailable, "no_data", "no snapshot yet")
			return
		}
		writeJSON(w, http.StatusOK, snap.Memory)
	}
}

func handleGPU(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := col.LatestSnapshot()
		if snap == nil {
			writeError(w, http.StatusServiceUnavailable, "no_data", "no snapshot yet")
			return
		}
		writeJSON(w, http.StatusOK, snap.GPU)
	}
}

func handleDisk(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := col.LatestSnapshot()
		if snap == nil {
			writeError(w, http.StatusServiceUnavailable, "no_data", "no snapshot yet")
			return
		}
		writeJSON(w, http.StatusOK, snap.Disk)
	}
}

func handleNetwork(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := col.LatestSnapshot()
		if snap == nil {
			writeError(w, http.StatusServiceUnavailable, "no_data", "no snapshot yet")
			return
		}
		writeJSON(w, http.StatusOK, snap.Network)
	}
}

// ----- process handlers -----

func handleProcesses(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := col.LatestSnapshot()
		if snap == nil {
			writeError(w, http.StatusServiceUnavailable, "no_data", "no snapshot yet")
			return
		}
		writeJSON(w, http.StatusOK, snap.Processes)
	}
}

func handleProcessTree(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tree := col.LatestProcessTree()
		writeJSON(w, http.StatusOK, tree)
	}
}

// ----- ports handler -----

func handlePorts(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ports := col.LatestPortBindings()
		writeJSON(w, http.StatusOK, ports)
	}
}

// ----- alerts handler -----

func handleAlerts(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		alerts := store.ActiveAlerts()
		writeJSON(w, http.StatusOK, alerts)
	}
}

// ----- AI advisor handlers -----

func handleAIStatus(advisor *ai.Advisor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if advisor == nil {
			writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"enabled":  advisor.IsEnabled(),
			"provider": advisor.ProviderName(),
		})
	}
}

func handleAIChat(advisor *ai.Advisor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if advisor == nil || !advisor.IsEnabled() {
			writeError(w, http.StatusServiceUnavailable, "ai_disabled", "AI advisor not configured")
			return
		}
		var body struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
			return
		}
		if body.Message == "" {
			writeError(w, http.StatusBadRequest, "invalid_message", "message required")
			return
		}
		ctx := buildAIContext(body.Message)
		answer, err := advisor.Analyze(r.Context(), ctx)
		if err != nil {
			writeError(w, http.StatusBadGateway, "ai_error", "AI provider request failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"answer":  answer,
			"actions": nil,
		})
	}
}

// buildAIContext creates an AnomalyContext from user message.
func buildAIContext(message string) ai.AnomalyContext {
	return ai.AnomalyContext{
		Type:    "general",
		Details: message,
	}
}

// ----- config handlers -----

func handleConfigGet(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := store.GetConfig()
		writeJSON(w, http.StatusOK, cfg)
	}
}

func handleConfigUpdate(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cfg map[string]any
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
			return
		}
		if err := store.UpdateConfig(cfg); err != nil {
			writeError(w, http.StatusBadRequest, "update_failed", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// ----- process control handlers -----

func handleKill(ctrl *controller.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid, ok := parsePIDParam(w, r)
		if !ok {
			return
		}
		if err := ctrl.ValidateAction(pid, "kill"); err != nil {
			writeError(w, http.StatusForbidden, "protected_process", "Cannot perform this action on protected system process")
			return
		}
		if err := ctrl.Kill(pid); err != nil {
			writeError(w, http.StatusInternalServerError, "kill_failed", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "pid": pid})
	}
}

func handleSuspend(ctrl *controller.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid, ok := parsePIDParam(w, r)
		if !ok {
			return
		}
		if err := ctrl.ValidateAction(pid, "suspend"); err != nil {
			writeError(w, http.StatusForbidden, "protected_process", "Cannot perform this action on protected system process")
			return
		}
		if err := ctrl.Suspend(pid); err != nil {
			writeError(w, http.StatusInternalServerError, "suspend_failed", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "pid": pid})
	}
}

func handleResume(ctrl *controller.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid, ok := parsePIDParam(w, r)
		if !ok {
			return
		}
		if err := ctrl.ValidateAction(pid, "resume"); err != nil {
			writeError(w, http.StatusForbidden, "protected_process", "Cannot perform this action on protected system process")
			return
		}
		if err := ctrl.Resume(pid); err != nil {
			writeError(w, http.StatusInternalServerError, "resume_failed", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "pid": pid})
	}
}

func handlePriority(ctrl *controller.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid, ok := parsePIDParam(w, r)
		if !ok {
			return
		}
		if err := ctrl.ValidateAction(pid, "priority"); err != nil {
			writeError(w, http.StatusForbidden, "protected_process", "Cannot perform this action on protected system process")
			return
		}
		var body struct {
			Nice int `json:"nice"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
			return
		}
		if err := ctrl.SetPriority(pid, body.Nice); err != nil {
			writeError(w, http.StatusInternalServerError, "priority_failed", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "pid": pid, "nice": body.Nice})
	}
}