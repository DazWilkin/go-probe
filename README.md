# Go Probe

[![Go Reference](https://pkg.go.dev/badge/github.com/DazWilkin/go-probe.svg)](https://pkg.go.dev/github.com/DazWilkin/go-probe)
[![Go Report Card](https://goreportcard.com/badge/github.com/DazWilkin/go-probe)](https://goreportcard.com/report/github.com/DazWilkin/go-probe)

A preliminary (!) implementation for healthcheck (liveness|readiness) checking

The channel(s) used by `Updater` aren't part of the `Probe` struct.

`Updater` must be running for the probe to work.

The implementations (unfortunately) permits:

1. Forgetting to run `Updater`
1. Manually sharing the channels between subscriber and publisher

## Usage

### Subscriber

```golang
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

p := probe.New("example", logger)

// Handler's logger may differ
healthz := p.Handler(logger)

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

ch := make(chan Status)
signal := make(chan struct{})

// Cancellable
go p.Updater(ctx, ch, signal)
```

### Publisher

```golang
if err != nil {
    status := Status{
        Healthy: false,
        Message: "...",
    }
    ch <- status

    return
}

// Otherwise
status := Status{
    Healthy: true,
    Message: "ok",
}
ch <- status
```
