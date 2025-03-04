package probe

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestProbeStatus(t *testing.T) {
	name := "test"
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	probe := New(name, logger)

	ch := make(chan Status)

	// Run Go routine
	// It will terminate upon completion of the tests
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go probe.Updater(ctx, ch, nil)

	// Initially true (healthy)
	{
		status := probe.Status()

		if !status.Healthy {
			t.Error("expect Probe to be healthy")
		}
	}

	// Update with false (unhealthy)
	// Updates aren't immediate so let's wait a second for it to propagate
	{
		ch <- Status{
			Healthy: false,
			Message: "",
		}

		time.Sleep(1 * time.Second)

		status := probe.Status()

		if status.Healthy {
			t.Error("expect Probe to be unhealthy")
		}
	}

	// Update with true (healthy)
	// Updates aren't immediate so let's wait a second for it to propagate
	{
		ch <- Status{
			Healthy: true,
			Message: "",
		}

		time.Sleep(1 * time.Second)

		status := probe.Status()

		if !status.Healthy {
			t.Error("expect Probe to be healthy")
		}
	}
}
func TestProbeHandler(t *testing.T) {
	name := "test"
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	probe := New(name, logger)

	handler := probe.Handler(logger)

	ch := make(chan Status)
	signal := make(chan struct{})

	// Run Go routine
	// It will terminate upon completion of the tests
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go probe.Updater(ctx, ch, signal)

	rqst, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Initially true (healthy)
	{
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, rqst)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	}

	// Update with false (unhealthy)
	// Updates aren't immediate so let's wait a second for it to propagate
	{
		ch <- (Status{
			Healthy: false,
			Message: "",
		})

		// Await signal
		<-signal

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, rqst)

		if status := rr.Code; status == http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want not %v",
				status, http.StatusOK)
		}

	}

	// Update with true (healthy)
	// Updates aren't immediate so let's wait a second for it to propagate
	{
		ch <- (Status{
			Healthy: true,
			Message: "",
		})

		// Await signal
		<-signal

		want := ""

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, rqst)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		if rr.Body.String() != want {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), want)
		}
	}
}
