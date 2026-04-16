// SPDX-FileCopyrightText: 2026 dieter8322-arch
// SPDX-License-Identifier: MIT

package api

import (
	"errors"
	"net/http"

	"github.com/dieter8322-arch/convey-engine/apps/server/internal/app"
	"github.com/dieter8322-arch/convey-engine/apps/server/internal/repository"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	runService *app.RunService
}

type createManualRunRequest struct {
	ProjectName   string `json:"projectName"`
	RepoURL       string `json:"repoUrl"`
	Provider      string `json:"provider"`
	DefaultBranch string `json:"defaultBranch"`
	PipelineName  string `json:"pipelineName"`
	ConfigRaw     string `json:"configRaw"`
	Ref           string `json:"ref"`
	CommitSHA     string `json:"commitSha"`
	TriggeredBy   string `json:"triggeredBy"`
}

func NewHandler(runService *app.RunService) *Handler {
	return &Handler{runService: runService}
}

func (h *Handler) CreateManualRun(c *gin.Context) {
	var request createManualRunRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	run, err := h.runService.CreateManualRun(c.Request.Context(), app.CreateManualRunInput{
		ProjectName:   request.ProjectName,
		RepoURL:       request.RepoURL,
		Provider:      request.Provider,
		DefaultBranch: request.DefaultBranch,
		PipelineName:  request.PipelineName,
		ConfigRaw:     request.ConfigRaw,
		Ref:           request.Ref,
		CommitSHA:     request.CommitSHA,
		TriggeredBy:   request.TriggeredBy,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"run": run})
}

func (h *Handler) GetRun(c *gin.Context) {
	run, err := h.runService.GetRun(c.Request.Context(), c.Param("runID"))
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"run": run})
}

func writeError(c *gin.Context, err error) {
	_ = c.Error(err)

	switch {
	case errors.Is(err, app.ErrInvalidRunRequest), errors.Is(err, app.ErrInvalidPipeline):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, repository.ErrRunNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
