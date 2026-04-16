package service

import (
	"fmt"
	"strings"

	"github.com/dobby/filemanager/internal/domain/job"
	"github.com/dobby/filemanager/internal/domain/rule"
)

// TemplateRenderer is a Domain Service that substitutes template variables with context values.
//
// Supported variables: {project}, {type}, {YYYY}, {MM}, {DD}, {seq}, {original}, {ext}
type TemplateRenderer struct{}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

// RenderName renders a NamingTemplate using the given ProcessingContext.
func (r *TemplateRenderer) RenderName(tmpl rule.NamingTemplate, ctx job.ProcessingContext) string {
	return render(tmpl.TemplateString, ctx)
}

// RenderPath renders a TargetPathTemplate using the given ProcessingContext.
func (r *TemplateRenderer) RenderPath(tmpl rule.TargetPathTemplate, ctx job.ProcessingContext) string {
	return render(tmpl.PathTemplate, ctx)
}

func render(template string, ctx job.ProcessingContext) string {
	replacer := strings.NewReplacer(
		"{project}", ctx.Project,
		"{type}", ctx.TypeLabel,
		"{YYYY}", fmt.Sprintf("%04d", ctx.Date.Year()),
		"{MM}", fmt.Sprintf("%02d", ctx.Date.Month()),
		"{DD}", fmt.Sprintf("%02d", ctx.Date.Day()),
		"{seq}", ctx.Seq,
		"{original}", ctx.OriginalName,
		"{ext}", ctx.Extension,
	)
	return replacer.Replace(template)
}
