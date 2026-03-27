package widget

import (
	"strings"
	"testing"

	"github.com/Jason-Adam/vitals/internal/model"
)

func TestMessagesWidget_NilTranscriptReturnsEmpty(t *testing.T) {
	ctx := &model.RenderContext{Transcript: nil}
	cfg := defaultCfg()

	if got := Messages(ctx, cfg); !got.IsEmpty() {
		t.Errorf("Messages with nil Transcript: expected empty, got %q", got.Text)
	}
}

func TestMessagesWidget_ZeroCountReturnsEmpty(t *testing.T) {
	ctx := &model.RenderContext{Transcript: &model.TranscriptData{MessageCount: 0}}
	cfg := defaultCfg()

	if got := Messages(ctx, cfg); !got.IsEmpty() {
		t.Errorf("Messages with zero count: expected empty, got %q", got.Text)
	}
}

func TestMessagesWidget_NonZeroCountRendersCount(t *testing.T) {
	ctx := &model.RenderContext{Transcript: &model.TranscriptData{MessageCount: 7}}
	cfg := defaultCfg()

	got := Messages(ctx, cfg)
	if !strings.Contains(got.Text, "7") {
		t.Errorf("Messages: expected '7' in output, got %q", got.Text)
	}
	if !strings.Contains(got.Text, "msgs") {
		t.Errorf("Messages: expected 'msgs' in output, got %q", got.Text)
	}
}

func TestMessagesWidget_ExactFormat(t *testing.T) {
	ctx := &model.RenderContext{Transcript: &model.TranscriptData{MessageCount: 3}}
	cfg := defaultCfg()

	got := Messages(ctx, cfg)
	want := MutedStyle.Render("3 msgs")
	if got.Text != want {
		t.Errorf("Messages: expected %q, got %q", want, got.Text)
	}
}

func TestMessagesWidget_RegisteredInRegistry(t *testing.T) {
	if _, ok := Registry["messages"]; !ok {
		t.Error("Registry missing 'messages' widget")
	}
}
