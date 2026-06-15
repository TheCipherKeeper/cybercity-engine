// Command cybercity-engine — CLI entry point для CyberCity Engine.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheCipherKeeper/cybercity-engine/internal/application"
	"github.com/TheCipherKeeper/cybercity-engine/internal/config"
)

func main() {
	if err := run(); err != nil {
		slog.Error("engine failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		engineZip = flag.String("engine-zip", "", "Path or URL to the engine.zip / engine.json artifact.")
		host      = flag.String("host", "", "API bind host (default 0.0.0.0).")
		port      = flag.Int("port", 0, "API bind port (default 8000).")
		debug     = flag.Bool("debug", false, "Enable debug logging.")
	)
	flag.Parse()

	setupLogging(*debug)

	cfg := config.LoadEnvConfig()
	cfg.ApplyFlags(engineZip, host, port, debug)

	runtime, err := application.NewRuntime(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := runtime.Engine.Start(ctx); err != nil && err != context.Canceled {
			slog.Error("engine tick loop failed", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		slog.Info("shutting down")
		runtime.Engine.Stop()
		cancel()
	}()

	return runtime.Server.Run(ctx)
}

func setupLogging(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
}
