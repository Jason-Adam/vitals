package widget

import (
	"strings"
	"testing"

	"github.com/Jason-Adam/vitals/internal/model"
)

func TestProjectWidget_MergedOutputWithDirtyRepo(t *testing.T) {
	ctx := &model.RenderContext{
		Cwd: "/Users/kyle/Code/vitals",
		Git: &model.GitStatus{Branch: "main", Dirty: true},
	}
	cfg := defaultCfg()
	cfg.Directory.Levels = 1

	got := Project(ctx, cfg).Text

	if !strings.Contains(got, "vitals") {
		t.Errorf("expected directory 'vitals' in output, got %q", got)
	}
	if !strings.Contains(got, "main") {
		t.Errorf("expected branch 'main' in output, got %q", got)
	}
	if !strings.Contains(got, "*") {
		t.Errorf("expected dirty indicator '*' in output, got %q", got)
	}
}

func TestProjectWidget_DirectoryOnlyWhenGitNil(t *testing.T) {
	ctx := &model.RenderContext{
		Cwd: "/Users/kyle/Code/vitals",
		Git: nil,
	}
	cfg := defaultCfg()
	cfg.Directory.Levels = 1

	got := Project(ctx, cfg).Text

	if !strings.Contains(got, "vitals") {
		t.Errorf("expected directory 'vitals' in output, got %q", got)
	}
	// No branch should appear.
	if strings.Contains(got, "main") {
		t.Errorf("expected no branch when git is nil, got %q", got)
	}
}

func TestProjectWidget_EmptyCwdReturnsEmpty(t *testing.T) {
	ctx := &model.RenderContext{
		Cwd: "",
		Git: &model.GitStatus{Branch: "main"},
	}
	cfg := defaultCfg()

	if got := Project(ctx, cfg); !got.IsEmpty() {
		t.Errorf("expected empty result when Cwd is empty, got %q", got.Text)
	}
}

func TestProjectWidget_AheadBehindShownWhenNonzero(t *testing.T) {
	ctx := &model.RenderContext{
		Cwd: "/Users/kyle/Code/vitals",
		Git: &model.GitStatus{Branch: "feat/auth", AheadBy: 2, BehindBy: 1},
	}
	cfg := defaultCfg()
	cfg.Directory.Levels = 1

	got := Project(ctx, cfg).Text

	if !strings.Contains(got, "feat/auth") {
		t.Errorf("expected branch 'feat/auth', got %q", got)
	}
	if !strings.Contains(got, "↑2") {
		t.Errorf("expected '↑2' for ahead count, got %q", got)
	}
	if !strings.Contains(got, "↓1") {
		t.Errorf("expected '↓1' for behind count, got %q", got)
	}
}

func TestProjectWidget_ZeroAheadBehindNotShown(t *testing.T) {
	ctx := &model.RenderContext{
		Cwd: "/Users/kyle/Code/vitals",
		Git: &model.GitStatus{Branch: "main", AheadBy: 0, BehindBy: 0},
	}
	cfg := defaultCfg()
	cfg.Directory.Levels = 1

	got := Project(ctx, cfg).Text

	if strings.Contains(got, "↑") {
		t.Errorf("expected no ahead indicator when AheadBy is 0, got %q", got)
	}
	if strings.Contains(got, "↓") {
		t.Errorf("expected no behind indicator when BehindBy is 0, got %q", got)
	}
}

func TestProjectWidget_CleanRepoNoDirtyIndicator(t *testing.T) {
	ctx := &model.RenderContext{
		Cwd: "/Users/kyle/Code/vitals",
		Git: &model.GitStatus{Branch: "main", Dirty: false},
	}
	cfg := defaultCfg()
	cfg.Directory.Levels = 1

	got := Project(ctx, cfg).Text

	if strings.Contains(got, "*") {
		t.Errorf("expected no dirty indicator for clean repo, got %q", got)
	}
}

func TestProjectWidget_MultipleSegments(t *testing.T) {
	ctx := &model.RenderContext{
		Cwd: "/Users/kyle/Code/my-projects/vitals",
		Git: &model.GitStatus{Branch: "main"},
	}
	cfg := defaultCfg()
	cfg.Directory.Levels = 2

	got := Project(ctx, cfg).Text

	if !strings.Contains(got, "my-projects/vitals") {
		t.Errorf("expected 2-segment path 'my-projects/vitals', got %q", got)
	}
}

func TestProjectWidget_RegisteredInRegistry(t *testing.T) {
	if _, ok := Registry["project"]; !ok {
		t.Error("Registry missing 'project' widget")
	}
}
