package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

func TestRender_ProducesOutput(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName:  "Sonnet",
		ContextWindowSize: 200000,
		ContextPercent:    50,
		Cwd:               "/Users/kyle/Code/project",
	}
	cfg := config.LoadHud()

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	out := buf.String()
	if out == "" {
		t.Fatal("Render produced no output")
	}

	// Default config line 1 has model, context, directory
	if !strings.Contains(out, "Sonnet") {
		t.Errorf("expected 'Sonnet' in output, got %q", out)
	}
	if !strings.Contains(out, "50%") {
		t.Errorf("expected '50%%' in output, got %q", out)
	}
	if !strings.Contains(out, "project") {
		t.Errorf("expected 'project' in output, got %q", out)
	}
}

func TestRender_SkipsUnknownWidgets(t *testing.T) {
	ctx := &model.RenderContext{ModelDisplayName: "Opus"}
	cfg := config.LoadHud()
	cfg.Lines = []config.Line{
		{Widgets: []string{"nonexistent", "model"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	out := buf.String()
	if !strings.Contains(out, "Opus") {
		t.Errorf("expected 'Opus' after skipping unknown widget, got %q", out)
	}
}

func TestRender_SkipsEmptyLines(t *testing.T) {
	ctx := &model.RenderContext{} // no data -> all widgets return ""
	cfg := config.LoadHud()

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	if buf.String() != "" {
		t.Errorf("expected no output for empty context, got %q", buf.String())
	}
}

func TestRender_UsesSeparator(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName:  "Opus",
		ContextWindowSize: 200000,
		ContextPercent:    42,
	}
	cfg := config.LoadHud()
	cfg.Style.Separator = " :: "
	cfg.Lines = []config.Line{
		{Widgets: []string{"model", "context"}},
	}

	var buf bytes.Buffer
	Render(&buf, ctx, cfg)

	out := buf.String()
	if !strings.Contains(out, " :: ") {
		t.Errorf("expected ' :: ' separator in output, got %q", out)
	}
}
