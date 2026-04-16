// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package domain

import "encoding/json"

type PipelineDocument struct {
	Version  int               `yaml:"version" json:"version"`
	Triggers []PipelineTrigger `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Stages   []PipelineStage   `yaml:"stages" json:"stages"`
	Jobs     []PipelineJob     `yaml:"jobs" json:"jobs"`
}

type PipelineTrigger struct {
	Type     string   `yaml:"type" json:"type"`
	Branches []string `yaml:"branches,omitempty" json:"branches,omitempty"`
}

type PipelineStage struct {
	Name string `yaml:"name" json:"name"`
}

type PipelineJob struct {
	Name        string         `yaml:"name" json:"name"`
	Stage       string         `yaml:"stage" json:"stage"`
	Needs       []string       `yaml:"needs,omitempty" json:"needs,omitempty"`
	RunsOn      string         `yaml:"runs_on,omitempty" json:"runsOn,omitempty"`
	Environment string         `yaml:"environment,omitempty" json:"environment,omitempty"`
	Approval    string         `yaml:"approval,omitempty" json:"approval,omitempty"`
	Steps       []PipelineStep `yaml:"steps" json:"steps"`
}

type PipelineStep struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Run  string `yaml:"run,omitempty" json:"run,omitempty"`
}

type ResolvedPipeline struct {
	Document          PipelineDocument `json:"document"`
	ConfigHash        string           `json:"configHash"`
	ParsedSummaryJSON json.RawMessage  `json:"parsedSummary"`
}
