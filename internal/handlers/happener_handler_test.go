package handlers

import (
	"testing"

	"github.com/napuu/gpsp-bot/pkg/utils"
)

func TestHappenerCommandIsParsedFromBothSpellings(t *testing.T) {
	t.Setenv("ENABLED_FEATURES", "ping;happener")

	for _, rawText := range []string{"/happener", "/häppener", "!häppener"} {
		t.Run(rawText, func(t *testing.T) {
			ctx := &Context{rawText: rawText}

			genericMessageHandler := &GenericMessageHandler{}
			genericMessageHandler.SetNext(&mockHandler{})
			genericMessageHandler.Execute(ctx)

			if ctx.action != Happener {
				t.Errorf("Expected action %q, got %q", Happener, ctx.action)
			}
		})
	}
}

func TestHappenerCommandIgnoredWhenFeatureDisabled(t *testing.T) {
	t.Setenv("ENABLED_FEATURES", "ping")

	ctx := &Context{rawText: "/häppener"}

	genericMessageHandler := &GenericMessageHandler{}
	genericMessageHandler.SetNext(&mockHandler{})
	genericMessageHandler.Execute(ctx)

	if ctx.action != "" {
		t.Errorf("Expected no action, got %q", ctx.action)
	}
}

func TestHappenerPagesAreValidPageIDs(t *testing.T) {
	if len(happenerPages) == 0 {
		t.Fatal("Expected at least one teletext page to pick from")
	}

	seenNames := map[string]bool{}
	seenIDs := map[string]bool{}
	for _, page := range happenerPages {
		if !utils.IsValidTeletextPageID(page.id) {
			t.Errorf("Page %q has invalid id %q", page.name, page.id)
		}
		if seenNames[page.name] {
			t.Errorf("Duplicate page name %q", page.name)
		}
		if seenIDs[page.id] {
			t.Errorf("Duplicate page id %q", page.id)
		}
		seenNames[page.name] = true
		seenIDs[page.id] = true
	}
}

func TestPickHappenerPageStaysInListAndVaries(t *testing.T) {
	known := map[string]bool{}
	for _, page := range happenerPages {
		known[page.id] = true
	}

	picked := map[string]bool{}
	for range 200 {
		page := pickHappenerPage()
		if !known[page.id] {
			t.Fatalf("Picked unknown page %q", page.id)
		}
		picked[page.id] = true
	}

	if len(picked) != len(happenerPages) {
		t.Errorf("Picked %d distinct pages out of %d in 200 draws", len(picked), len(happenerPages))
	}
}

func TestHappenerHandlerIgnoresOtherActions(t *testing.T) {
	ctx := &Context{action: Ping}

	end := &mockHandler{}
	happenerHandler := &HappenerHandler{}
	happenerHandler.SetNext(end)
	happenerHandler.Execute(ctx)

	if ctx.finalImagePath != "" {
		t.Errorf("Expected no image path, got %q", ctx.finalImagePath)
	}
	if !end.executed {
		t.Error("Expected chain to continue to next handler")
	}
}
