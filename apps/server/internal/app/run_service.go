// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dieter8322-arch/convey-engine/apps/server/internal/domain"
)

var ErrInvalidRunRequest = errors.New("invalid run request")

type EnsureProjectParams struct {
	Name          string
	RepoURL       string
	Provider      string
	DefaultBranch string
}

type EnsurePipelineDefParams struct {
	ProjectID string
	Name      string
}

type CreatePipelineVersionParams struct {
	PipelineDefID     string
	ConfigRaw         string
	ConfigHash        string
	ParsedSummaryJSON []byte
}

type CreateRunParams struct {
	ProjectID         string
	PipelineVersionID string
	Status            domain.RunStatus
	Ref               string
	CommitSHA         string
	TriggerType       string
	TriggeredBy       string
}

type UpdateRunWorkflowParams struct {
	RunID              string
	Status             domain.RunStatus
	TemporalWorkflowID string
	TemporalRunID      string
	StartedAt          time.Time
}

type MetadataStore interface {
	EnsureProject(ctx context.Context, params EnsureProjectParams) (domain.Project, error)
	EnsurePipelineDef(ctx context.Context, params EnsurePipelineDefParams) (domain.PipelineDef, error)
	CreatePipelineVersion(ctx context.Context, params CreatePipelineVersionParams) (domain.PipelineVersion, error)
	CreateRun(ctx context.Context, params CreateRunParams) (domain.Run, error)
	UpdateRunWorkflow(ctx context.Context, params UpdateRunWorkflowParams) error
	UpdateRunStatus(ctx context.Context, runID string, status domain.RunStatus, finishedAt *time.Time) error
	GetRun(ctx context.Context, runID string) (domain.Run, error)
}

type StartPipelineRunInput struct {
	RunID             string
	ProjectID         string
	ProjectName       string
	PipelineDefID     string
	PipelineName      string
	PipelineVersionID string
}

type WorkflowHandle struct {
	WorkflowID string
	RunID      string
}

type WorkflowStarter interface {
	StartPipelineRun(ctx context.Context, input StartPipelineRunInput) (WorkflowHandle, error)
}

type CreateManualRunInput struct {
	ProjectName   string
	RepoURL       string
	Provider      string
	DefaultBranch string
	PipelineName  string
	ConfigRaw     string
	Ref           string
	CommitSHA     string
	TriggeredBy   string
}

type RunService struct {
	store    MetadataStore
	resolver PipelineResolver
	starter  WorkflowStarter
}

func NewRunService(store MetadataStore, resolver PipelineResolver, starter WorkflowStarter) *RunService {
	return &RunService{
		store:    store,
		resolver: resolver,
		starter:  starter,
	}
}

func (s *RunService) CreateManualRun(ctx context.Context, input CreateManualRunInput) (domain.Run, error) {
	if err := validateCreateManualRunInput(input); err != nil {
		return domain.Run{}, err
	}

	resolved, err := s.resolver.Resolve(ctx, input.ConfigRaw)
	if err != nil {
		return domain.Run{}, err
	}

	project, err := s.store.EnsureProject(ctx, EnsureProjectParams{
		Name:          strings.TrimSpace(input.ProjectName),
		RepoURL:       strings.TrimSpace(input.RepoURL),
		Provider:      firstNonEmpty(strings.TrimSpace(input.Provider), "github"),
		DefaultBranch: firstNonEmpty(strings.TrimSpace(input.DefaultBranch), "main"),
	})
	if err != nil {
		return domain.Run{}, fmt.Errorf("ensure project: %w", err)
	}

	pipelineDef, err := s.store.EnsurePipelineDef(ctx, EnsurePipelineDefParams{
		ProjectID: project.ID,
		Name:      strings.TrimSpace(input.PipelineName),
	})
	if err != nil {
		return domain.Run{}, fmt.Errorf("ensure pipeline definition: %w", err)
	}

	pipelineVersion, err := s.store.CreatePipelineVersion(ctx, CreatePipelineVersionParams{
		PipelineDefID:     pipelineDef.ID,
		ConfigRaw:         strings.TrimSpace(input.ConfigRaw),
		ConfigHash:        resolved.ConfigHash,
		ParsedSummaryJSON: resolved.ParsedSummaryJSON,
	})
	if err != nil {
		return domain.Run{}, fmt.Errorf("create pipeline version: %w", err)
	}

	run, err := s.store.CreateRun(ctx, CreateRunParams{
		ProjectID:         project.ID,
		PipelineVersionID: pipelineVersion.ID,
		Status:            domain.RunStatusPending,
		Ref:               strings.TrimSpace(input.Ref),
		CommitSHA:         strings.TrimSpace(input.CommitSHA),
		TriggerType:       "manual",
		TriggeredBy:       firstNonEmpty(strings.TrimSpace(input.TriggeredBy), "manual"),
	})
	if err != nil {
		return domain.Run{}, fmt.Errorf("create run: %w", err)
	}

	handle, err := s.starter.StartPipelineRun(ctx, StartPipelineRunInput{
		RunID:             run.ID,
		ProjectID:         project.ID,
		ProjectName:       project.Name,
		PipelineDefID:     pipelineDef.ID,
		PipelineName:      pipelineDef.Name,
		PipelineVersionID: pipelineVersion.ID,
	})
	if err != nil {
		now := time.Now().UTC()
		_ = s.store.UpdateRunStatus(ctx, run.ID, domain.RunStatusFailed, &now)

		return domain.Run{}, fmt.Errorf("start workflow: %w", err)
	}

	if err := s.store.UpdateRunWorkflow(ctx, UpdateRunWorkflowParams{
		RunID:              run.ID,
		Status:             domain.RunStatusRunning,
		TemporalWorkflowID: handle.WorkflowID,
		TemporalRunID:      handle.RunID,
		StartedAt:          time.Now().UTC(),
	}); err != nil {
		return domain.Run{}, fmt.Errorf("update run workflow info: %w", err)
	}

	return s.store.GetRun(ctx, run.ID)
}

func (s *RunService) GetRun(ctx context.Context, runID string) (domain.Run, error) {
	if strings.TrimSpace(runID) == "" {
		return domain.Run{}, fmt.Errorf("%w: runID is required", ErrInvalidRunRequest)
	}

	return s.store.GetRun(ctx, runID)
}

func validateCreateManualRunInput(input CreateManualRunInput) error {
	switch {
	case strings.TrimSpace(input.ProjectName) == "":
		return fmt.Errorf("%w: projectName is required", ErrInvalidRunRequest)
	case strings.TrimSpace(input.PipelineName) == "":
		return fmt.Errorf("%w: pipelineName is required", ErrInvalidRunRequest)
	case strings.TrimSpace(input.ConfigRaw) == "":
		return fmt.Errorf("%w: configRaw is required", ErrInvalidRunRequest)
	case strings.TrimSpace(input.Ref) == "":
		return fmt.Errorf("%w: ref is required", ErrInvalidRunRequest)
	case strings.TrimSpace(input.CommitSHA) == "":
		return fmt.Errorf("%w: commitSha is required", ErrInvalidRunRequest)
	default:
		return nil
	}
}

func firstNonEmpty(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
