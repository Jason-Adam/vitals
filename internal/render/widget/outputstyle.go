package widget

import (
	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// OutputStyle renders the current Claude Code output style name.
// Returns an empty WidgetResult when ctx.OutputStyle is empty.
func OutputStyle(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.OutputStyle == "" {
		return WidgetResult{}
	}
	return WidgetResult{Text: ctx.OutputStyle}
}
