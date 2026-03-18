package widget

import (
	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Project composes the Directory and Git widgets into a single segment
// without a separator between them.
// Format: '{directory} {branch}{dirty}{ahead}{behind}'
// e.g. 'tail-claude-hud main*' or 'tail-claude-hud feat/auth↑2'
// Returns an empty WidgetResult when both sub-widgets are empty.
// When Git has no data, renders directory only.
// FgColor is left empty because the sub-widgets compose multiple styles;
// the renderer passes the pre-styled Text through as-is.
func Project(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	dir := Directory(ctx, cfg)
	if dir.IsEmpty() {
		return WidgetResult{}
	}

	git := Git(ctx, cfg)
	if git.IsEmpty() {
		return dir
	}

	return WidgetResult{
		Text:      dir.Text + " " + git.Text,
		PlainText: dir.PlainText + " " + git.PlainText,
		FgColor:   "13", // inherit directory's dominant color
	}
}
