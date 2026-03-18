package widget

import (
	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Session renders the current session name with dim styling.
// Returns an empty WidgetResult when ctx.Transcript is nil or SessionName is empty.
// FgColor is left empty because dimStyle uses faint rather than a foreground color;
// the renderer passes the pre-styled Text through as-is.
func Session(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.Transcript == nil || ctx.Transcript.SessionName == "" {
		return WidgetResult{}
	}
	name := ctx.Transcript.SessionName
	return WidgetResult{
		Text:      MutedStyle.Render(name),
		PlainText: name,
		FgColor:   "8",
	}
}
