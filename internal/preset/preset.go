// Package preset defines the Preset data model and built-in presets.
// A preset bundles visual configuration only: widget layout, separator,
// render mode, theme, icon set, and directory style. It does not include
// data-source settings such as thresholds, git options, or speed windows.
package preset

import (
	"sort"

	"github.com/Jason-Adam/vitals/internal/config"
)

// Preset holds the visual configuration for a named layout.
type Preset struct {
	Name           string
	Lines          []config.Line
	Separator      string
	Icons          string
	Mode           string // plain, powerline, minimal
	Theme          string
	DirectoryStyle string
}

// Load returns the named built-in preset. Returns a zero-value Preset and
// false when name is not a known built-in.
func Load(name string) (Preset, bool) {
	p, ok := builtins[name]
	return p, ok
}

// BuiltinNames returns the names of all built-in presets in sorted order.
func BuiltinNames() []string {
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ApplyPreset overlays a preset's visual settings onto cfg.
// It replaces Lines and sets Style.Separator, Icons, Mode, Theme, and
// Directory.Style from the preset, then calls config.ResolveTheme so the
// updated palette is ready for rendering.
//
// User-preference fields are never touched: Thresholds, Context, Git, Speed,
// and Theme.Overrides remain exactly as the user configured them.
func ApplyPreset(cfg *config.Config, p Preset) {
	if len(p.Lines) > 0 {
		cfg.Lines = p.Lines
	}
	if p.Separator != "" {
		cfg.Style.Separator = p.Separator
	}
	if p.Icons != "" {
		cfg.Style.Icons = p.Icons
	}
	if p.Mode != "" {
		cfg.Style.Mode = p.Mode
	}
	if p.Theme != "" {
		cfg.Style.Theme = p.Theme
	}
	if p.DirectoryStyle != "" {
		cfg.Directory.Style = p.DirectoryStyle
	}
	config.ResolveTheme(cfg)
}

// LoadHudWithPreset loads the default config via config.LoadHud and then
// applies the named preset. If presetName is empty, or if no preset with
// that name exists in the built-in registry, the unmodified config is
// returned. This function never returns nil.
func LoadHudWithPreset(presetName string) *config.Config {
	cfg := config.LoadHud()
	if presetName == "" {
		return cfg
	}
	p, ok := Load(presetName)
	if !ok {
		return cfg
	}
	ApplyPreset(cfg, p)
	return cfg
}
