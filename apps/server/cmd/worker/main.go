// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package main

import (
	"log/slog"
	"os"

	"github.com/dieter8322-arch/convey-engine/apps/server/internal/config"
	"github.com/dieter8322-arch/convey-engine/apps/server/internal/repository"
	engineworkflow "github.com/dieter8322-arch/convey-engine/apps/server/internal/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
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

	workerInstance := worker.New(temporalClient, cfg.Temporal.TaskQueue, worker.Options{})
	engineworkflow.Register(workerInstance, engineworkflow.NewRunLifecycleActivities(repository.NewStore(db)))

	if err := workerInstance.Run(worker.InterruptCh()); err != nil {
		logger.Error("temporal worker exited", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
