package gracefulshutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WithShutdownSignal waits for a CTRL+C signal and then executes the provided callback functions.
// It wraps the provided context and returns a new context that will be canceled when the shutdown signal is received.
func WithShutdownSignal(
	ctx context.Context,
	callbacks ...func(),
) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		for _, callback := range callbacks {
			callback()
		}
		cancel()
	}()

	return ctx
}
