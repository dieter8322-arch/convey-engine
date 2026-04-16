// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dieter8322-arch/convey-engine/apps/server/internal/domain"
	"gopkg.in/yaml.v3"
)

var ErrInvalidPipeline = errors.New("invalid pipeline config")

type PipelineResolver interface {
	Resolve(ctx context.Context, configRaw string) (domain.ResolvedPipeline, error)
}

type YAMLPipelineResolver struct{}

func NewYAMLPipelineResolver() *YAMLPipelineResolver {
	return &YAMLPipelineResolver{}
}

func (r *YAMLPipelineResolver) Resolve(_ context.Context, configRaw string) (domain.ResolvedPipeline, error) {
	trimmed := strings.TrimSpace(configRaw)
	if trimmed == "" {
		return domain.ResolvedPipeline{}, fmt.Errorf("%w: configRaw is required", ErrInvalidPipeline)
	}

	var document domain.PipelineDocument
	if err := yaml.Unmarshal([]byte(trimmed), &document); err != nil {
		return domain.ResolvedPipeline{}, fmt.Errorf("%w: decode yaml: %w", ErrInvalidPipeline, err)
	}

	if err := validatePipelineDocument(document); err != nil {
		return domain.ResolvedPipeline{}, err
	}

	summary, err := buildPipelineSummary(document)
	if err != nil {
		return domain.ResolvedPipeline{}, fmt.Errorf("%w: marshal summary: %w", ErrInvalidPipeline, err)
	}

	sum := sha256.Sum256([]byte(trimmed))

	return domain.ResolvedPipeline{
		Document:          document,
		ConfigHash:        hex.EncodeToString(sum[:]),
		ParsedSummaryJSON: summary,
	}, nil
}

func validatePipelineDocument(document domain.PipelineDocument) error {
	if document.Version <= 0 {
		return fmt.Errorf("%w: version must be greater than zero", ErrInvalidPipeline)
	}
	if len(document.Stages) == 0 {
		return fmt.Errorf("%w: at least one stage is required", ErrInvalidPipeline)
	}
	if len(document.Jobs) == 0 {
		return fmt.Errorf("%w: at least one job is required", ErrInvalidPipeline)
	}

	stageNames := make(map[string]struct{}, len(document.Stages))
	for _, stage := range document.Stages {
		name := strings.TrimSpace(stage.Name)
		if name == "" {
			return fmt.Errorf("%w: stage name cannot be empty", ErrInvalidPipeline)
		}
		if _, exists := stageNames[name]; exists {
			return fmt.Errorf("%w: duplicated stage %q", ErrInvalidPipeline, name)
		}
		stageNames[name] = struct{}{}
	}

	jobNames := make(map[string]struct{}, len(document.Jobs))
	for _, job := range document.Jobs {
		name := strings.TrimSpace(job.Name)
		if name == "" {
			return fmt.Errorf("%w: job name cannot be empty", ErrInvalidPipeline)
		}
		if _, exists := jobNames[name]; exists {
			return fmt.Errorf("%w: duplicated job %q", ErrInvalidPipeline, name)
		}
		if _, exists := stageNames[job.Stage]; !exists {
			return fmt.Errorf("%w: job %q references unknown stage %q", ErrInvalidPipeline, name, job.Stage)
		}
		if len(job.Steps) == 0 {
			return fmt.Errorf("%w: job %q must contain at least one step", ErrInvalidPipeline, name)
		}

		jobNames[name] = struct{}{}
	}

	for _, job := range document.Jobs {
		for _, dependency := range job.Needs {
			if _, exists := jobNames[dependency]; !exists {
				return fmt.Errorf("%w: job %q references unknown dependency %q", ErrInvalidPipeline, job.Name, dependency)
			}
		}
	}

	return nil
}

func buildPipelineSummary(document domain.PipelineDocument) (json.RawMessage, error) {
	stageNames := make([]string, 0, len(document.Stages))
	for _, stage := range document.Stages {
		stageNames = append(stageNames, stage.Name)
	}

	jobNames := make([]string, 0, len(document.Jobs))
	for _, job := range document.Jobs {
		jobNames = append(jobNames, job.Name)
	}

	triggerTypes := make([]string, 0, len(document.Triggers))
	for _, trigger := range document.Triggers {
		triggerTypes = append(triggerTypes, trigger.Type)
	}

	summary, err := json.Marshal(map[string]any{
		"version":      document.Version,
		"stageCount":   len(document.Stages),
		"jobCount":     len(document.Jobs),
		"triggerCount": len(document.Triggers),
		"stages":       stageNames,
		"jobs":         jobNames,
		"triggers":     triggerTypes,
	})
	if err != nil {
		return nil, err
	}

	return summary, nil
}
