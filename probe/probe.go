package probe

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
)

// Probe is a type that represents a mutex-protected Status messsage
type Probe struct {
	name   string
	status Status
	logger *slog.Logger
	mu     sync.RWMutex
}

// New is a function that returns a new Probe
func New(name string, l *slog.Logger) *Probe {
	logger := l.With("probe", name)
	logger.Info("creating",
		"name", name,
	)
	return &Probe{
		name: name,
		status: Status{
			// Assume healthy
			// Otherwise, won't be updated until first time metrics are scraped
			Healthy: true,
		},
		logger: logger,
		mu:     sync.RWMutex{},
	}
}

// Handler is a method that returns a new Probe handler
// Handler accepts a possibly different logger than Probe itself
func (h *Probe) Handler(l *slog.Logger) http.HandlerFunc {
	logger := l.With("probe", h.name)
	logger.Info("Creating handler")

	return func(w http.ResponseWriter, _ *http.Request) {
		logger.Info("Invoking handler")

		w.Header().Set("Content-Type", "text/plain")

		status := h.Status()

		if status.Healthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		if _, err := w.Write([]byte(status.Message)); err != nil {
			logger.Error("unable to write response")
		}
	}
}

// Status is a method that returns the current status
func (h *Probe) Status() Status {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.logger.Info("Getting status",
		"status", h.status,
	)

	return h.status
}

// Updater is a method that awaits messages and updates the shared status
// Updater should be run as a Go routine so that it updates the status on new messages
// ch is required as the channel to subscribe to Status updates
// signal is optional (may be nil) as the channel to publish notifications of updates (used by testing)
func (h *Probe) Updater(ctx context.Context, ch <-chan Status, signal chan<- struct{}) {
	h.logger.Info("Starting updater")
	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Done")
			return
		case status := <-ch:
			h.logger.Info("Updating status",
				"status", status,
			)

			h.mu.Lock()
			h.status = status
			h.mu.Unlock()

			// (Optional) signal updated
			if signal != nil {
				signal <- struct{}{}
			}
		}
	}
}
