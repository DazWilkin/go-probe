# Go Probe

The channel(s) used by `Updater` aren't part of the `Probe` struct.

`Updater` must be running for the probe to work.

The implementations (unfortunately) permits:

1. Forgetting to run `Updater`
1. Manually sharing the channels between subscriber and publisher

## Usage

### Subscriber

```golang
p := probe.New("liveness", logger)
healthz := p.Handler(logger)

// Channel is shared by the Updater (subscriber) and the InstananeousValuesCollector (publisher)
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

ch := make(chan probe.Status)
signal := make(chan struct{})
go p.Updater(ctx, ch, signal)
```

### Publisher

```golang
if err != nil {
    status := probe.Status{
        Healthy: false,
        Message: "...",
    }
    ch <- status

    return
}

status := probe.Status{
    Healthy: true,
    Message: "ok",
}
ch <- status
```
