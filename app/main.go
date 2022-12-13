// The simple service for authenticate and access management with the original Docker registry https://github.com/distribution/distribution
// Some parts of code in this project borrow from Umputun projects https://github.com/umputun

package main

import (
	log "github.com/go-pkgz/lgr"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	version = "unknown"
	opts    *Options
)

func main() {
	log.Printf("REGISTRY ADMIN PORTAL: %s\n", version)

	var err error
	opts, err = parseArgs()

	if err != nil {
		log.Fatalf("[ERROR] failed to parse config parameters: %v", err)
	}
	setupLog(opts.Debug)

	log.Print("[INFO] server starting...")

	if err = run(); err != nil && err != http.ErrServerClosed {
		log.Printf("[ERROR] failed to run server: %v", err)
		os.Exit(1)
	}
}

func setupLog(dbg bool) {
	if dbg {
		log.Setup(log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)
		return
	}

	log.Setup(log.Msec, log.LevelBraces)
}

// getDump reads runtime stack and returns as a string
func getDump() string {
	maxSize := 5 * 1024 * 1024
	stacktrace := make([]byte, maxSize)
	length := runtime.Stack(stacktrace, true)
	if length > maxSize {
		length = maxSize
	}
	return string(stacktrace[:length])
}

// nolint:gochecknoinits // can't avoid it in this place
func init() {
	// catch SIGQUIT and print stack traces
	sigChan := make(chan os.Signal, 1)
	go func() {
		for range sigChan {
			log.Printf("[INFO] SIGQUIT detected, dump:\n%s", getDump())
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}
