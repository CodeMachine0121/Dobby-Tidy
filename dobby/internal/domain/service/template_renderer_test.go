package service_test

import (
	"testing"
	"time"

	"github.com/dobby/filemanager/internal/domain/job"
	"github.com/dobby/filemanager/internal/domain/rule"
	"github.com/dobby/filemanager/internal/domain/service"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// S-11: 渲染包含所有支援變數的命名樣版
// ─────────────────────────────────────────────────────────────────────────────

// S-11-1
func TestTemplateRenderer_RenderName_AllVariables(t *testing.T) {
	// Arrange
	renderer := service.NewTemplateRenderer()
	tmpl := createNamingTemplate("{project}-{type}-{YYYY}{MM}{DD}-{seq}.{ext}")
	ctx := createFullProcessingContext()

	// Act
	result := renderer.RenderName(tmpl, ctx)

	// Assert
	assert.Equal(t, "my-app-screenshot-20260416-001.png", result)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-12: 渲染包含 {original} 變數的命名樣版
// ─────────────────────────────────────────────────────────────────────────────

// S-12-1
func TestTemplateRenderer_RenderName_OriginalVariable(t *testing.T) {
	// Arrange
	renderer := service.NewTemplateRenderer()
	tmpl := createNamingTemplate("{original}-{YYYY}{MM}{DD}.{ext}")
	ctx := createContextWithOriginal("report-draft", "pdf")

	// Act
	result := renderer.RenderName(tmpl, ctx)

	// Assert
	assert.Equal(t, "report-draft-20260416.pdf", result)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-13: 渲染目標路徑樣版（含動態 {project} 路徑）
// ─────────────────────────────────────────────────────────────────────────────

// S-13-1
func TestTemplateRenderer_RenderPath_ProjectVariable(t *testing.T) {
	// Arrange
	renderer := service.NewTemplateRenderer()
	tmpl := createTargetPathTemplate("~/Projects/{project}/assets/")
	ctx := createContextWithProject("my-app")

	// Act
	result := renderer.RenderPath(tmpl, ctx)

	// Assert
	assert.Equal(t, "~/Projects/my-app/assets/", result)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func createNamingTemplate(s string) rule.NamingTemplate {
	return rule.NamingTemplate{TemplateString: s}
}

func createTargetPathTemplate(s string) rule.TargetPathTemplate {
	return rule.TargetPathTemplate{PathTemplate: s}
}

func createFullProcessingContext() job.ProcessingContext {
	return job.ProcessingContext{
		Project:      "my-app",
		TypeLabel:    "screenshot",
		Date:         time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC),
		Seq:          "001",
		OriginalName: "Untitled",
		Extension:    "png",
	}
}

func createContextWithOriginal(originalName, ext string) job.ProcessingContext {
	return job.ProcessingContext{
		OriginalName: originalName,
		Extension:    ext,
		Date:         time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC),
	}
}

func createContextWithProject(project string) job.ProcessingContext {
	return job.ProcessingContext{
		Project: project,
	}
}
