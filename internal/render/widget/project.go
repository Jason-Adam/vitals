package widget

import (
	"fmt"
	"strings"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Project renders a merged directory + git segment.
// Format: '{directory} {branch}{dirty}{ahead}{behind}'
// e.g. 'tail-claude-hud main*' or 'tail-claude-hud feat/auth↑2'
// Directory is magenta bold; branch name is cyan; dirty/ahead/behind are dim.
// Returns an empty WidgetResult when ctx.Cwd is empty.
// When ctx.Git is nil, renders directory only (no git suffix).
// FgColor is left empty because the widget composes multiple styles internally;
// the renderer passes the pre-styled Text through as-is.
func Project(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.Cwd == "" {
		return WidgetResult{}
	}

	levels := cfg.Directory.Levels
	if levels <= 0 {
		levels = 1
	}

	dirName := lastNSegments(ctx.Cwd, levels)
	dir := dirStyle.Render(dirName)

	if ctx.Git == nil {
		return WidgetResult{Text: dir}
	}

	g := ctx.Git
	branch := gitBranchStyle.Render(g.Branch)

	// Build the dim suffix: dirty indicator, ahead, behind.
	var dimParts strings.Builder
	if g.IsDirty() {
		dimParts.WriteString("*")
	}
	if g.AheadBy > 0 {
		dimParts.WriteString(fmt.Sprintf("↑%d", g.AheadBy))
	}
	if g.BehindBy > 0 {
		dimParts.WriteString(fmt.Sprintf("↓%d", g.BehindBy))
	}

	suffix := dimParts.String()
	if suffix != "" {
		return WidgetResult{Text: dir + " " + branch + gitDimStyle.Render(suffix)}
	}
	return WidgetResult{Text: dir + " " + branch}
}
