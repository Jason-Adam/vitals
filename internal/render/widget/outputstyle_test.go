package widget

import (
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

func TestOutputStyleWidget_PresentStyleName(t *testing.T) {
	ctx := &model.RenderContext{OutputStyle: "default"}
	cfg := defaultCfg()

	got := OutputStyle(ctx, cfg)
	if got.Text != "default" {
		t.Errorf("OutputStyle: expected %q, got %q", "default", got.Text)
	}
}

func TestOutputStyleWidget_EmptyString(t *testing.T) {
	ctx := &model.RenderContext{OutputStyle: ""}
	cfg := defaultCfg()

	if got := OutputStyle(ctx, cfg); !got.IsEmpty() {
		t.Errorf("OutputStyle with empty string: expected empty, got %q", got.Text)
	}
}

func TestOutputStyleWidget_NilContext(t *testing.T) {
	// Simulate nil-equivalent: RenderContext with zero-value OutputStyle.
	ctx := &model.RenderContext{}
	cfg := defaultCfg()

	if got := OutputStyle(ctx, cfg); !got.IsEmpty() {
		t.Errorf("OutputStyle with zero-value context: expected empty, got %q", got.Text)
	}
}

func TestOutputStyleWidget_VariousStyleNames(t *testing.T) {
	tests := []struct {
		name  string
		style string
		want  string
	}{
		{"default style", "default", "default"},
		{"concise style", "concise", "concise"},
		{"verbose style", "verbose", "verbose"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &model.RenderContext{OutputStyle: tt.style}
			cfg := defaultCfg()
			got := OutputStyle(ctx, cfg)
			if got.Text != tt.want {
				t.Errorf("OutputStyle(%q): expected %q, got %q", tt.style, tt.want, got.Text)
			}
		})
	}
}

func TestOutputStyleWidget_RegisteredInRegistry(t *testing.T) {
	if _, ok := Registry["outputstyle"]; !ok {
		t.Error("Registry missing 'outputstyle' widget")
	}
}
