// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dieter8322-arch/convey-engine/apps/server/internal/app"
	"github.com/dieter8322-arch/convey-engine/apps/server/internal/domain"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ErrRunNotFound = errors.New("run not found")

type Store struct {
	db *gorm.DB
}

func OpenPostgres(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	return db, nil
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) EnsureProject(ctx context.Context, params app.EnsureProjectParams) (domain.Project, error) {
	record := ProjectRecord{
		ID:            uuid.NewString(),
		Name:          params.Name,
		RepoURL:       params.RepoURL,
		Provider:      params.Provider,
		DefaultBranch: params.DefaultBranch,
		Status:        string(domain.ProjectStatusActive),
	}

	if err := s.db.WithContext(ctx).
		Where("name = ?", params.Name).
		Attrs(record).
		FirstOrCreate(&record).Error; err != nil {
		return domain.Project{}, err
	}

	return projectFromRecord(record), nil
}

func (s *Store) EnsurePipelineDef(ctx context.Context, params app.EnsurePipelineDefParams) (domain.PipelineDef, error) {
	record := PipelineDefRecord{
		ID:        uuid.NewString(),
		ProjectID: params.ProjectID,
		Name:      params.Name,
		Status:    string(domain.PipelineDefStatusActive),
	}

	if err := s.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", params.ProjectID, params.Name).
		Attrs(record).
		FirstOrCreate(&record).Error; err != nil {
		return domain.PipelineDef{}, err
	}

	return pipelineDefFromRecord(record), nil
}

func (s *Store) CreatePipelineVersion(ctx context.Context, params app.CreatePipelineVersionParams) (domain.PipelineVersion, error) {
	nextVersion, err := s.nextPipelineVersion(ctx, params.PipelineDefID)
	if err != nil {
		return domain.PipelineVersion{}, err
	}

	record := PipelineVersionRecord{
		ID:                uuid.NewString(),
		PipelineDefID:     params.PipelineDefID,
		Version:           nextVersion,
		ConfigRaw:         params.ConfigRaw,
		ConfigHash:        params.ConfigHash,
		ParsedSummaryJSON: params.ParsedSummaryJSON,
		CreatedAt:         time.Now().UTC(),
	}

	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return domain.PipelineVersion{}, err
	}

	return pipelineVersionFromRecord(record), nil
}

func (s *Store) CreateRun(ctx context.Context, params app.CreateRunParams) (domain.Run, error) {
	now := time.Now().UTC()
	record := RunRecord{
		ID:                uuid.NewString(),
		ProjectID:         params.ProjectID,
		PipelineVersionID: params.PipelineVersionID,
		Status:            string(params.Status),
		Ref:               params.Ref,
		CommitSHA:         params.CommitSHA,
		TriggerType:       params.TriggerType,
		TriggeredBy:       params.TriggeredBy,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return domain.Run{}, err
	}

	return runFromRecord(record), nil
}

func (s *Store) UpdateRunWorkflow(ctx context.Context, params app.UpdateRunWorkflowParams) error {
	workflowID := params.TemporalWorkflowID
	runID := params.TemporalRunID

	updates := map[string]any{
		"status":               string(params.Status),
		"temporal_workflow_id": &workflowID,
		"temporal_run_id":      &runID,
		"started_at":           params.StartedAt,
		"updated_at":           time.Now().UTC(),
	}

	result := s.db.WithContext(ctx).
		Model(&RunRecord{}).
		Where("id = ?", params.RunID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRunNotFound
	}

	return nil
}

func (s *Store) UpdateRunStatus(ctx context.Context, runID string, status domain.RunStatus, finishedAt *time.Time) error {
	updates := map[string]any{
		"status":     string(status),
		"updated_at": time.Now().UTC(),
	}
	if finishedAt != nil {
		updates["finished_at"] = *finishedAt
	}

	result := s.db.WithContext(ctx).
		Model(&RunRecord{}).
		Where("id = ?", runID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRunNotFound
	}

	return nil
}

func (s *Store) GetRun(ctx context.Context, runID string) (domain.Run, error) {
	var record RunRecord
	if err := s.db.WithContext(ctx).
		Where("id = ?", runID).
		Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Run{}, ErrRunNotFound
		}
		return domain.Run{}, err
	}

	return runFromRecord(record), nil
}

func (s *Store) nextPipelineVersion(ctx context.Context, pipelineDefID string) (int, error) {
	var record PipelineVersionRecord
	err := s.db.WithContext(ctx).
		Where("pipeline_def_id = ?", pipelineDefID).
		Order("version DESC").
		Take(&record).Error
	if err == nil {
		return record.Version + 1, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 1, nil
	}

	return 0, err
}
