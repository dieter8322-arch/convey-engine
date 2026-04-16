// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package domain

import (
	"encoding/json"
	"time"
)

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
)

type PipelineDefStatus string

const (
	PipelineDefStatusActive   PipelineDefStatus = "active"
	PipelineDefStatusDisabled PipelineDefStatus = "disabled"
)

type RunStatus string

const (
	RunStatusPending          RunStatus = "pending"
	RunStatusRunning          RunStatus = "running"
	RunStatusAwaitingApproval RunStatus = "awaiting_approval"
	RunStatusSucceeded        RunStatus = "succeeded"
	RunStatusFailed           RunStatus = "failed"
	RunStatusCanceled         RunStatus = "canceled"
)

type Project struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	RepoURL       string        `json:"repoUrl"`
	Provider      string        `json:"provider"`
	DefaultBranch string        `json:"defaultBranch"`
	Status        ProjectStatus `json:"status"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

type PipelineDef struct {
	ID        string            `json:"id"`
	ProjectID string            `json:"projectId"`
	Name      string            `json:"name"`
	Status    PipelineDefStatus `json:"status"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

type PipelineVersion struct {
	ID                string          `json:"id"`
	PipelineDefID     string          `json:"pipelineDefId"`
	Version           int             `json:"version"`
	ConfigRaw         string          `json:"configRaw"`
	ConfigHash        string          `json:"configHash"`
	ParsedSummaryJSON json.RawMessage `json:"parsedSummary"`
	CreatedAt         time.Time       `json:"createdAt"`
}

type Run struct {
	ID                 string    `json:"id"`
	ProjectID          string    `json:"projectId"`
	PipelineVersionID  string    `json:"pipelineVersionId"`
	Status             RunStatus `json:"status"`
	Ref                string    `json:"ref"`
	CommitSHA          string    `json:"commitSha"`
	TriggerType        string    `json:"triggerType"`
	TriggeredBy        string    `json:"triggeredBy"`
	TemporalWorkflowID string    `json:"temporalWorkflowId"`
	TemporalRunID      string    `json:"temporalRunId"`
	StartedAt          time.Time `json:"startedAt"`
	FinishedAt         time.Time `json:"finishedAt"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}
