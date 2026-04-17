package telereactions

import (
	"reflect"
	"testing"

	tele "gopkg.in/telebot.v4"
)

func emoji(s string) tele.Reaction {
	return tele.Reaction{Type: tele.ReactionTypeEmoji, Emoji: s}
}

func TestDispatch_AddRemoveDiff(t *testing.T) {
	var added, removed []string
	h := Handlers{
		OnAdd:    func(e Event) { added = append(added, e.Emoji) },
		OnRemove: func(e Event) { removed = append(removed, e.Emoji) },
	}

	u := &tele.Update{
		MessageReaction: &tele.MessageReaction{
			Chat:        &tele.Chat{ID: -1001234567890},
			MessageID:   42,
			OldReaction: []tele.Reaction{emoji("👍"), emoji("❤")},
			NewReaction: []tele.Reaction{emoji("👍"), emoji("🔥")},
		},
	}
	dispatch(u, h)

	if !reflect.DeepEqual(added, []string{"🔥"}) {
		t.Errorf("added = %v, want [🔥]", added)
	}
	if !reflect.DeepEqual(removed, []string{"❤"}) {
		t.Errorf("removed = %v, want [❤]", removed)
	}
}

func TestDispatch_NilHandlersDoNotPanic(t *testing.T) {
	u := &tele.Update{
		MessageReaction: &tele.MessageReaction{
			Chat:        &tele.Chat{ID: 1},
			NewReaction: []tele.Reaction{emoji("👍")},
		},
	}
	dispatch(u, Handlers{}) // OnAdd + OnRemove both nil
}

func TestDispatch_IgnoresNonReactionUpdates(t *testing.T) {
	called := false
	h := Handlers{OnAdd: func(Event) { called = true }}
	dispatch(&tele.Update{}, h)
	dispatch(nil, h)
	if called {
		t.Error("OnAdd fired for update without MessageReaction")
	}
}

func TestDispatch_CarriesChatAndMessageID(t *testing.T) {
	var got Event
	h := Handlers{OnAdd: func(e Event) { got = e }}
	u := &tele.Update{
		MessageReaction: &tele.MessageReaction{
			Chat:        &tele.Chat{ID: -100999},
			MessageID:   777,
			User:        &tele.User{ID: 55},
			NewReaction: []tele.Reaction{emoji("👎")},
		},
	}
	dispatch(u, h)

	if got.Chat.ID != -100999 || got.MessageID != 777 || got.User.ID != 55 || got.Emoji != "👎" {
		t.Errorf("event = %+v", got)
	}
}

func TestEnsureAllowed_AppendsWhenMissing(t *testing.T) {
	p := &tele.LongPoller{AllowedUpdates: []string{"message"}}
	ensureAllowed(p)
	want := []string{"message", "message_reaction"}
	if !reflect.DeepEqual(p.AllowedUpdates, want) {
		t.Errorf("AllowedUpdates = %v, want %v", p.AllowedUpdates, want)
	}
}

func TestEnsureAllowed_NoDuplicate(t *testing.T) {
	p := &tele.LongPoller{AllowedUpdates: []string{"message", "message_reaction"}}
	ensureAllowed(p)
	if len(p.AllowedUpdates) != 2 {
		t.Errorf("AllowedUpdates = %v, expected no duplicate", p.AllowedUpdates)
	}
}

func TestWrap_ReturnsMiddlewarePoller(t *testing.T) {
	inner := &tele.LongPoller{}
	mp := Wrap(inner, Handlers{})
	if mp == nil || mp.Poller != inner {
		t.Error("Wrap did not return a MiddlewarePoller wrapping the inner poller")
	}
	if len(inner.AllowedUpdates) == 0 || inner.AllowedUpdates[0] != "message_reaction" {
		t.Errorf("Wrap did not inject message_reaction; got %v", inner.AllowedUpdates)
	}
}
