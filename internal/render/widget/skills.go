package widget

import (
	"strings"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Skills renders the list of skill names invoked in the current session.
// Skills are extracted from "Skill" tool_use blocks in the transcript.
//
// The widget shows a comma-separated list of unique skill names seen so far,
// newest-first. Returns an empty WidgetResult when ctx.Transcript is nil or no skills
// have been invoked.
// FgColor is left empty because dimStyle uses faint rather than a foreground color;
// the renderer passes the pre-styled Text through as-is.
func Skills(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.Transcript == nil || len(ctx.Transcript.SkillNames) == 0 {
		return WidgetResult{}
	}

	// Deduplicate while preserving most-recent-first order. Walk the slice
	// in reverse (newest last → newest first after reversal) and keep the
	// first occurrence of each name.
	seen := make(map[string]bool, len(ctx.Transcript.SkillNames))
	unique := make([]string, 0, len(ctx.Transcript.SkillNames))
	for i := len(ctx.Transcript.SkillNames) - 1; i >= 0; i-- {
		name := ctx.Transcript.SkillNames[i]
		if !seen[name] {
			seen[name] = true
			unique = append(unique, name)
		}
	}

	list := strings.Join(unique, ", ")
	return WidgetResult{Text: MutedStyle.Render(list)}
}
