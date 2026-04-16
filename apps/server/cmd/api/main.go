// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/dieter8322-arch/convey-engine/apps/server/internal/api"
	"github.com/dieter8322-arch/convey-engine/apps/server/internal/app"
	"github.com/dieter8322-arch/convey-engine/apps/server/internal/config"
	"github.com/dieter8322-arch/convey-engine/apps/server/internal/repository"
	engineworkflow "github.com/dieter8322-arch/convey-engine/apps/server/internal/workflow"
	"go.temporal.io/sdk/client"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := repository.OpenPostgres(cfg.Database.DSN)
	if err != nil {
		logger.Error("open database failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	temporalClient, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.Address,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		logger.Error("open temporal client failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer temporalClient.Close()

	store := repository.NewStore(db)
	runService := app.NewRunService(
		store,
		app.NewYAMLPipelineResolver(),
		engineworkflow.NewTemporalStarter(temporalClient, cfg.Temporal.TaskQueue),
	)

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           api.NewRouter(logger, api.NewHandler(runService)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("http server exited", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
