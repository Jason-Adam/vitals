package widget

import (
	"github.com/Jason-Adam/vitals/internal/config"
	"github.com/Jason-Adam/vitals/internal/model"
)

// Service renders the repository name (e.g. "vitals") so the user always
// knows which project they are working in. Returns empty when the name
// is not available.
func Service(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.ServiceName == "" {
		return WidgetResult{}
	}
	icons := IconsFor(cfg.Style.Icons)
	plain := icons.Folder + " " + ctx.ServiceName
	return WidgetResult{
		Text:      dirStyle.Render(plain),
		PlainText: plain,
		FgColor:   "13",
	}
}
