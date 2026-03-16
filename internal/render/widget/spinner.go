package widget

import (
	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Spinner renders an always-on braille spinner for testing refresh cadence.
// It is unconditional — it renders on every tick regardless of tool, agent,
// or thinking state. This lets us verify the statusline refresh rate sustains
// continuous animation without adding any behavioral logic.
//
// The frame is driven by a monotonic invocation counter stored in the transcript
// snapshot rather than wall-clock time, so successive invocations always produce
// a different frame even when they fall within the same 200ms window.
func Spinner(ctx *model.RenderContext, _ *config.Config) string {
	if ctx.Transcript != nil {
		return yellowStyle.Render(spinnerFrameFromCounter(ctx.Transcript.SpinnerFrame))
	}
	// Fall back to time-based frame when no transcript data is available.
	return yellowStyle.Render(spinnerFrame())
}
