// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/dieter8322-arch/convey-engine/apps/server/internal/app"
	"github.com/dieter8322-arch/convey-engine/apps/server/internal/domain"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	temporalworkflow "go.temporal.io/sdk/workflow"
)

type PipelineRunWorkflowInput struct {
	RunID             string
	ProjectID         string
	ProjectName       string
	PipelineDefID     string
	PipelineName      string
	PipelineVersionID string
}

type PipelineRunWorkflowResult struct {
	RunID       string `json:"runId"`
	FinalStatus string `json:"finalStatus"`
}

type RunStatusUpdater interface {
	UpdateRunStatus(ctx context.Context, runID string, status domain.RunStatus, finishedAt *time.Time) error
}

type RunLifecycleActivities struct {
	store RunStatusUpdater
}

type TemporalStarter struct {
	client    client.Client
	taskQueue string
}

func NewRunLifecycleActivities(store RunStatusUpdater) *RunLifecycleActivities {
	return &RunLifecycleActivities{store: store}
}

func NewTemporalStarter(client client.Client, taskQueue string) *TemporalStarter {
	return &TemporalStarter{
		client:    client,
		taskQueue: taskQueue,
	}
}

func (s *TemporalStarter) StartPipelineRun(ctx context.Context, input app.StartPipelineRunInput) (app.WorkflowHandle, error) {
	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("pipeline-run-%s", input.RunID),
		TaskQueue: s.taskQueue,
	}

	run, err := s.client.ExecuteWorkflow(ctx, options, PipelineRunWorkflow, PipelineRunWorkflowInput{
		RunID:             input.RunID,
		ProjectID:         input.ProjectID,
		ProjectName:       input.ProjectName,
		PipelineDefID:     input.PipelineDefID,
		PipelineName:      input.PipelineName,
		PipelineVersionID: input.PipelineVersionID,
	})
	if err != nil {
		return app.WorkflowHandle{}, err
	}

	return app.WorkflowHandle{
		WorkflowID: run.GetID(),
		RunID:      run.GetRunID(),
	}, nil
}

func (a *RunLifecycleActivities) MarkRunSucceededActivity(ctx context.Context, runID string) error {
	finishedAt := time.Now().UTC()
	return a.store.UpdateRunStatus(ctx, runID, domain.RunStatusSucceeded, &finishedAt)
}

func (a *RunLifecycleActivities) MarkRunFailedActivity(ctx context.Context, runID string) error {
	finishedAt := time.Now().UTC()
	return a.store.UpdateRunStatus(ctx, runID, domain.RunStatusFailed, &finishedAt)
}

func PipelineRunWorkflow(ctx temporalworkflow.Context, input PipelineRunWorkflowInput) (PipelineRunWorkflowResult, error) {
	options := temporalworkflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = temporalworkflow.WithActivityOptions(ctx, options)

	if err := temporalworkflow.ExecuteActivity(ctx, "MarkRunSucceededActivity", input.RunID).Get(ctx, nil); err != nil {
		_ = temporalworkflow.ExecuteActivity(ctx, "MarkRunFailedActivity", input.RunID).Get(ctx, nil)
		return PipelineRunWorkflowResult{}, err
	}

	return PipelineRunWorkflowResult{
		RunID:       input.RunID,
		FinalStatus: string(domain.RunStatusSucceeded),
	}, nil
}

func Register(workerInstance worker.Worker, activities *RunLifecycleActivities) {
	workerInstance.RegisterWorkflow(PipelineRunWorkflow)
	workerInstance.RegisterActivity(activities)
}
