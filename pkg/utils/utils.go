package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

func OpenBrowser(url string) error {
	var err error

	// Try to open the URL in different OS
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = os.ErrNotExist
	}

	return err
}

func Shutdown() <-chan struct{} {
	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine to listen for signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // Wait for a signal
		fmt.Println("Received termination signal, initiating shutdown...")
		cancel() // Cancel the context when a signal is received
	}()

	return ctx.Done()
}
