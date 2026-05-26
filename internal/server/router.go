package server

import (
	"io"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/burak/linux-dashboard/internal/ai"
	"github.com/burak/linux-dashboard/internal/collector"
	"github.com/burak/linux-dashboard/internal/controller"
	"github.com/burak/linux-dashboard/internal/event"
	"github.com/burak/linux-dashboard/internal/storage"
)

// WireRoutes registers all HTTP routes on the chi router.
func WireRoutes(r *chi.Mux, col *collector.Manager, store *storage.Store, advisor *ai.Advisor, emitter *event.Emitter) {
	// Root endpoint — serve embedded frontend index.html
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		subFS, _ := fs.Sub(webAssets, "web")
		if subFS != nil {
			if f, err := subFS.Open("index.html"); err == nil {
				defer f.Close()
				content, _ := io.ReadAll(f)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write(content)
				return
			}
		}
		// Fallback: JSON API info
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"Linux Dashboard","version":"dev","endpoints":["/api/v1/system","/api/v1/cpu","/api/v1/memory","/api/v1/disk","/api/v1/network","/api/v1/processes","/api/v1/ports","/api/v1/alerts","/api/v1/ai/status","/api/v1/stream"]}`))
	})

	// System endpoints.
	r.Get("/api/v1/system", handleSystem(col, store))
	r.Get("/api/v1/cpu", handleCPU(col))
	r.Get("/api/v1/memory", handleMemory(col))
	r.Get("/api/v1/gpu", handleGPU(col))
	r.Get("/api/v1/disk", handleDisk(col))
	r.Get("/api/v1/network", handleNetwork(col))

	// New Gece Nöbeti endpoints.
	r.Get("/api/v1/host", handleHost(col))
	r.Get("/api/v1/cores", handleCores(col))
	r.Get("/api/v1/sensors", handleSensors(col))
	r.Get("/api/v1/syslog", handleSyslog(col))
	r.Get("/api/v1/connections", handleConnections(col))

	// Process endpoints.
	r.Get("/api/v1/processes", handleProcesses(col))
	r.Get("/api/v1/processes/tree", handleProcessTree(col))

	// Port endpoints.
	r.Get("/api/v1/ports", handlePorts(col))

	// Alert endpoints.
	r.Get("/api/v1/alerts", handleAlerts(store))

	// AI advisor endpoints.
	r.Get("/api/v1/ai/status", handleAIStatus(advisor))
	r.Post("/api/v1/ai/chat", handleAIChat(advisor))

	// Config endpoints.
	r.Get("/api/v1/config", handleConfigGet(store))
	r.Post("/api/v1/config", handleConfigUpdate(store))

	// Process control endpoints.
	r.Post("/api/v1/processes/{pid}/kill", handleKill(controller.NewController(nil)))
	r.Post("/api/v1/processes/{pid}/suspend", handleSuspend(controller.NewController(nil)))
	r.Post("/api/v1/processes/{pid}/resume", handleResume(controller.NewController(nil)))
	r.Post("/api/v1/processes/{pid}/priority", handlePriority(controller.NewController(nil)))

	// SSE stream endpoint.
	r.Get("/api/v1/stream", handleSSE(emitter, col))
}