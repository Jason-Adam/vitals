package widget

import (
	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Spinner renders an always-on braille spinner for testing refresh cadence.
// It is unconditional — it renders on every tick regardless of tool, agent,
// or thinking state. This lets us verify the statusline refresh rate sustains
// continuous animation without adding any behavioral logic.
func Spinner(_ *model.RenderContext, _ *config.Config) string {
	return yellowStyle.Render(spinnerFrame())
}
