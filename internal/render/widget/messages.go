package widget

import (
	"fmt"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Messages renders the number of conversational turns in the current session.
// Tool_result entries are excluded because they carry tool output back to the
// model rather than representing a human or assistant turn.
// Returns an empty WidgetResult when ctx.Transcript is nil or no turns have been counted yet.
// FgColor is left empty because dimStyle uses faint rather than a foreground color;
// the renderer passes the pre-styled Text through as-is.
func Messages(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.Transcript == nil || ctx.Transcript.MessageCount == 0 {
		return WidgetResult{}
	}
	plain := fmt.Sprintf("%d msgs", ctx.Transcript.MessageCount)
	return WidgetResult{
		Text:      MutedStyle.Render(plain),
		PlainText: plain,
		FgColor:   "8",
	}
}
