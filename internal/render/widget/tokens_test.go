package widget

import (
	"strings"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

func TestTokensWidget_AllThreePresent(t *testing.T) {
	ctx := &model.RenderContext{
		InputTokens:   45100,
		CacheCreation: 8000,
		CacheRead:     4300,
	}
	cfg := defaultCfg()

	got := Tokens(ctx, cfg)
	if got.IsEmpty() {
		t.Fatal("expected non-empty output when all token fields are set")
	}

	// 45100 → "45.1k in", cache = 8000+4300 = 12300 → "12.3k cache"
	for _, want := range []string{"45.1k in", "12.3k cache", "·"} {
		if !strings.Contains(got.Text, want) {
			t.Errorf("Tokens all present: output %q does not contain %q", got.Text, want)
		}
	}
}

func TestTokensWidget_CacheZero(t *testing.T) {
	ctx := &model.RenderContext{
		InputTokens:   20000,
		CacheCreation: 0,
		CacheRead:     0,
	}
	cfg := defaultCfg()

	got := Tokens(ctx, cfg)
	if got.IsEmpty() {
		t.Fatal("expected non-empty output when only InputTokens is set")
	}

	if !strings.Contains(got.Text, "20.0k in") {
		t.Errorf("Tokens cache-zero: expected '20.0k in', got %q", got.Text)
	}
	// No cache section should appear when both cache fields are zero.
	if strings.Contains(got.Text, "cache") {
		t.Errorf("Tokens cache-zero: output %q should not contain 'cache' when cache counts are zero", got.Text)
	}
}

func TestTokensWidget_AllZero(t *testing.T) {
	ctx := &model.RenderContext{
		InputTokens:   0,
		CacheCreation: 0,
		CacheRead:     0,
	}
	cfg := defaultCfg()

	if got := Tokens(ctx, cfg); !got.IsEmpty() {
		t.Errorf("Tokens all-zero: expected empty, got %q", got.Text)
	}
}

func TestTokensWidget_ZeroValueContext(t *testing.T) {
	// A zero-value RenderContext has all token fields as zero; must return empty.
	ctx := &model.RenderContext{}
	cfg := defaultCfg()

	if got := Tokens(ctx, cfg); !got.IsEmpty() {
		t.Errorf("Tokens zero-value context: expected empty, got %q", got.Text)
	}
}

func TestTokensWidget_RegisteredInRegistry(t *testing.T) {
	fn, ok := Registry["tokens"]
	if !ok {
		t.Fatal("'tokens' not found in widget.Registry")
	}
	if fn == nil {
		t.Fatal("'tokens' registry entry is nil")
	}
}

func TestTokensWidget_SmallCounts(t *testing.T) {
	// Counts below 1000 should render without a 'k' suffix.
	ctx := &model.RenderContext{
		InputTokens:   500,
		CacheCreation: 200,
		CacheRead:     0,
	}
	cfg := defaultCfg()

	got := Tokens(ctx, cfg)
	if !strings.Contains(got.Text, "500 in") {
		t.Errorf("Tokens small counts: expected '500 in', got %q", got.Text)
	}
	if !strings.Contains(got.Text, "200 cache") {
		t.Errorf("Tokens small counts: expected '200 cache', got %q", got.Text)
	}
}
