// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package repository

import (
	"encoding/json"
	"time"

	"github.com/dieter8322-arch/convey-engine/apps/server/internal/domain"
)

type ProjectRecord struct {
	ID            string    `gorm:"column:id;primaryKey;type:uuid"`
	Name          string    `gorm:"column:name;size:128;not null;uniqueIndex:uk_projects_name"`
	RepoURL       string    `gorm:"column:repo_url;size:512;not null"`
	Provider      string    `gorm:"column:provider;size:32;not null;index:idx_projects_provider"`
	DefaultBranch string    `gorm:"column:default_branch;size:128;not null"`
	Status        string    `gorm:"column:status;size:32;not null"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null"`
}

func (ProjectRecord) TableName() string {
	return "projects"
}

type PipelineDefRecord struct {
	ID        string    `gorm:"column:id;primaryKey;type:uuid"`
	ProjectID string    `gorm:"column:project_id;type:uuid;not null;uniqueIndex:uk_pipeline_defs_project_name;index:idx_pipeline_defs_project_status"`
	Name      string    `gorm:"column:name;size:128;not null;uniqueIndex:uk_pipeline_defs_project_name"`
	Status    string    `gorm:"column:status;size:32;not null;index:idx_pipeline_defs_project_status"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (PipelineDefRecord) TableName() string {
	return "pipeline_defs"
}

type PipelineVersionRecord struct {
	ID                string    `gorm:"column:id;primaryKey;type:uuid"`
	PipelineDefID     string    `gorm:"column:pipeline_def_id;type:uuid;not null;uniqueIndex:uk_pipeline_versions_pipeline_version"`
	Version           int       `gorm:"column:version;not null;uniqueIndex:uk_pipeline_versions_pipeline_version"`
	ConfigRaw         string    `gorm:"column:config_raw;type:text;not null"`
	ConfigHash        string    `gorm:"column:config_hash;size:128;not null;index:idx_pipeline_versions_config_hash"`
	ParsedSummaryJSON []byte    `gorm:"column:parsed_summary_json;type:json;not null"`
	CreatedAt         time.Time `gorm:"column:created_at;not null"`
}

func (PipelineVersionRecord) TableName() string {
	return "pipeline_versions"
}

type RunRecord struct {
	ID                 string    `gorm:"column:id;primaryKey;type:uuid"`
	ProjectID          string    `gorm:"column:project_id;type:uuid;not null;index:idx_runs_project_status_created"`
	PipelineVersionID  string    `gorm:"column:pipeline_version_id;type:uuid;not null;index:idx_runs_pipeline_version_created"`
	Status             string    `gorm:"column:status;size:32;not null;index:idx_runs_project_status_created"`
	Ref                string    `gorm:"column:ref;size:255;not null"`
	CommitSHA          string    `gorm:"column:commit_sha;size:64;not null"`
	TriggerType        string    `gorm:"column:trigger_type;size:32;not null"`
	TriggeredBy        string    `gorm:"column:triggered_by;size:128;not null"`
	TemporalWorkflowID *string   `gorm:"column:temporal_workflow_id;size:255;uniqueIndex:uk_runs_temporal_workflow_id"`
	TemporalRunID      *string   `gorm:"column:temporal_run_id;size:255"`
	StartedAt          time.Time `gorm:"column:started_at"`
	FinishedAt         time.Time `gorm:"column:finished_at"`
	CreatedAt          time.Time `gorm:"column:created_at;not null;index:idx_runs_project_status_created;index:idx_runs_pipeline_version_created"`
	UpdatedAt          time.Time `gorm:"column:updated_at;not null"`
}

func (RunRecord) TableName() string {
	return "runs"
}

func projectFromRecord(record ProjectRecord) domain.Project {
	return domain.Project{
		ID:            record.ID,
		Name:          record.Name,
		RepoURL:       record.RepoURL,
		Provider:      record.Provider,
		DefaultBranch: record.DefaultBranch,
		Status:        domain.ProjectStatus(record.Status),
		CreatedAt:     record.CreatedAt,
		UpdatedAt:     record.UpdatedAt,
	}
}

func pipelineDefFromRecord(record PipelineDefRecord) domain.PipelineDef {
	return domain.PipelineDef{
		ID:        record.ID,
		ProjectID: record.ProjectID,
		Name:      record.Name,
		Status:    domain.PipelineDefStatus(record.Status),
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func pipelineVersionFromRecord(record PipelineVersionRecord) domain.PipelineVersion {
	return domain.PipelineVersion{
		ID:                record.ID,
		PipelineDefID:     record.PipelineDefID,
		Version:           record.Version,
		ConfigRaw:         record.ConfigRaw,
		ConfigHash:        record.ConfigHash,
		ParsedSummaryJSON: json.RawMessage(record.ParsedSummaryJSON),
		CreatedAt:         record.CreatedAt,
	}
}

func runFromRecord(record RunRecord) domain.Run {
	temporalWorkflowID := ""
	if record.TemporalWorkflowID != nil {
		temporalWorkflowID = *record.TemporalWorkflowID
	}

	temporalRunID := ""
	if record.TemporalRunID != nil {
		temporalRunID = *record.TemporalRunID
	}

	return domain.Run{
		ID:                 record.ID,
		ProjectID:          record.ProjectID,
		PipelineVersionID:  record.PipelineVersionID,
		Status:             domain.RunStatus(record.Status),
		Ref:                record.Ref,
		CommitSHA:          record.CommitSHA,
		TriggerType:        record.TriggerType,
		TriggeredBy:        record.TriggeredBy,
		TemporalWorkflowID: temporalWorkflowID,
		TemporalRunID:      temporalRunID,
		StartedAt:          record.StartedAt,
		FinishedAt:         record.FinishedAt,
		CreatedAt:          record.CreatedAt,
		UpdatedAt:          record.UpdatedAt,
	}
}
