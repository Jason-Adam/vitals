package widget

import (
	"strings"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// TestLinesWidget_BothZeroReturnsEmpty covers the nil/zero data case.
func TestLinesWidget_BothZeroReturnsEmpty(t *testing.T) {
	ctx := &model.RenderContext{}
	cfg := defaultCfg()

	if got := Lines(ctx, cfg); !got.IsEmpty() {
		t.Errorf("Lines both zero: expected empty, got %q", got.Text)
	}
}

// TestLinesWidget_OnlyAdditions renders only the green addition count.
func TestLinesWidget_OnlyAdditions(t *testing.T) {
	ctx := &model.RenderContext{LinesAdded: 42}
	cfg := defaultCfg()

	got := Lines(ctx, cfg)
	if !strings.Contains(got.Text, "+42") {
		t.Errorf("Lines only additions: expected '+42' in output, got %q", got.Text)
	}
	if strings.Contains(got.Text, "-") {
		t.Errorf("Lines only additions: unexpected '-' in output, got %q", got.Text)
	}
}

// TestLinesWidget_OnlyRemovals renders only the red removal count.
func TestLinesWidget_OnlyRemovals(t *testing.T) {
	ctx := &model.RenderContext{LinesRemoved: 17}
	cfg := defaultCfg()

	got := Lines(ctx, cfg)
	if !strings.Contains(got.Text, "-17") {
		t.Errorf("Lines only removals: expected '-17' in output, got %q", got.Text)
	}
	if strings.Contains(got.Text, "+") {
		t.Errorf("Lines only removals: unexpected '+' in output, got %q", got.Text)
	}
}

// TestLinesWidget_BothPresent renders both counts separated by a space.
func TestLinesWidget_BothPresent(t *testing.T) {
	ctx := &model.RenderContext{LinesAdded: 100, LinesRemoved: 23}
	cfg := defaultCfg()

	got := Lines(ctx, cfg)
	if !strings.Contains(got.Text, "+100") {
		t.Errorf("Lines both: expected '+100' in output, got %q", got.Text)
	}
	if !strings.Contains(got.Text, "-23") {
		t.Errorf("Lines both: expected '-23' in output, got %q", got.Text)
	}
}

// TestLinesWidget_NilCostData covers the case where StdinData.Cost was nil
// and both RenderContext fields were left at their zero values.
func TestLinesWidget_NilCostData(t *testing.T) {
	// When gather sees a nil Cost pointer, LinesAdded and LinesRemoved remain 0.
	ctx := &model.RenderContext{LinesAdded: 0, LinesRemoved: 0}
	cfg := defaultCfg()

	if got := Lines(ctx, cfg); !got.IsEmpty() {
		t.Errorf("Lines nil cost data: expected empty, got %q", got.Text)
	}
}
