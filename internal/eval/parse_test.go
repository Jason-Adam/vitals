package eval

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/render"
)

func TestParseSimpleColor(t *testing.T) {
	segs := Parse("\x1b[31mhello\x1b[0m")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d: %+v", len(segs), segs)
	}
	seg := segs[0]
	if seg.Text != "hello" {
		t.Errorf("expected text 'hello', got %q", seg.Text)
	}
	if seg.Fg.Type != ColorANSI16 || seg.Fg.Index != 1 {
		t.Errorf("expected ANSI16 fg index 1 (red), got %+v", seg.Fg)
	}
}

func TestParseXterm256(t *testing.T) {
	segs := Parse("\x1b[38;5;75mtext\x1b[0m")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d: %+v", len(segs), segs)
	}
	seg := segs[0]
	if seg.Text != "text" {
		t.Errorf("expected text 'text', got %q", seg.Text)
	}
	if seg.Fg.Type != ColorXterm256 || seg.Fg.Index != 75 {
		t.Errorf("expected Xterm256 fg index 75, got %+v", seg.Fg)
	}
}

func TestParseMultipleSegments(t *testing.T) {
	// Two words with different ANSI16 colors.
	input := "\x1b[32mgreen\x1b[0m \x1b[31mred\x1b[0m"
	segs := Parse(input)

	// We expect three segments: "green", " ", "red".
	// The space between resets is unstyled and distinct from the colored words.
	if len(segs) < 2 {
		t.Fatalf("expected at least 2 segments, got %d: %+v", len(segs), segs)
	}

	// Find green and red segments by color.
	var greenFound, redFound bool
	for _, s := range segs {
		if s.Fg.Type == ColorANSI16 && s.Fg.Index == 2 && s.Text == "green" {
			greenFound = true
		}
		if s.Fg.Type == ColorANSI16 && s.Fg.Index == 1 && s.Text == "red" {
			redFound = true
		}
	}
	if !greenFound {
		t.Errorf("did not find green segment in %+v", segs)
	}
	if !redFound {
		t.Errorf("did not find red segment in %+v", segs)
	}
}

func TestParseBoldFaint(t *testing.T) {
	// Bold text.
	segs := Parse("\x1b[1mBOLD\x1b[0m")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	if !segs[0].Bold {
		t.Error("expected Bold=true")
	}
	if segs[0].Faint {
		t.Error("expected Faint=false")
	}

	// Faint text.
	segs2 := Parse("\x1b[2mfaint\x1b[0m")
	if len(segs2) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs2))
	}
	if segs2[0].Faint != true {
		t.Error("expected Faint=true")
	}
	if segs2[0].Bold {
		t.Error("expected Bold=false")
	}
}

func TestParseReset(t *testing.T) {
	// Color set, then reset, then plain text.
	segs := Parse("\x1b[31mred\x1b[0mplain")
	// Expect two segments: "red" (ANSI16 fg 1) and "plain" (default).
	if len(segs) != 2 {
		t.Fatalf("expected 2 segments, got %d: %+v", len(segs), segs)
	}
	if segs[0].Fg.Type != ColorANSI16 || segs[0].Fg.Index != 1 {
		t.Errorf("first segment should be red ANSI16, got %+v", segs[0].Fg)
	}
	if segs[1].Fg.Type != ColorDefault {
		t.Errorf("second segment should have default fg after reset, got %+v", segs[1].Fg)
	}
	if segs[1].Text != "plain" {
		t.Errorf("expected text 'plain', got %q", segs[1].Text)
	}
}

func TestParseMergeAdjacent(t *testing.T) {
	// Two consecutive sequences with the same color — the plain text between
	// them is identical in style, so all three parts merge.
	input := "\x1b[32mfoo\x1b[32mbar\x1b[0m"
	segs := Parse(input)
	// Both "foo" and "bar" share ANSI16 fg 2, so they must be a single segment.
	if len(segs) != 1 {
		t.Fatalf("expected 1 merged segment, got %d: %+v", len(segs), segs)
	}
	if segs[0].Text != "foobar" {
		t.Errorf("expected merged text 'foobar', got %q", segs[0].Text)
	}
}

func TestParseEmptyText(t *testing.T) {
	// Escape-only input — no visible text.
	segs := Parse("\x1b[31m\x1b[0m")
	if len(segs) != 0 {
		t.Errorf("expected no segments for escape-only input, got %d: %+v", len(segs), segs)
	}
}

// TestParseRealOutput exercises Parse against the actual output of render.Render().
func TestParseRealOutput(t *testing.T) {
	ctx := &model.RenderContext{
		ModelDisplayName:  "Claude Sonnet 4",
		ContextPercent:    42,
		ContextWindowSize: 200000,
	}
	cfg := config.LoadHud()

	var buf bytes.Buffer
	render.Render(&buf, ctx, cfg)

	raw := buf.String()
	if raw == "" {
		t.Fatal("Render produced no output")
	}

	segs := Parse(raw)
	if len(segs) < 2 {
		t.Fatalf("expected at least 2 segments, got %d from %q", len(segs), raw)
	}

	// The first segment should contain visible text that includes "Sonnet".
	found := false
	for _, s := range segs {
		if strings.Contains(s.Text, "Sonnet") {
			found = true
			break
		}
	}
	if !found {
		texts := make([]string, len(segs))
		for i, s := range segs {
			texts[i] = s.Text
		}
		t.Errorf("no segment contains 'Sonnet'; segment texts: %v", texts)
	}
}
