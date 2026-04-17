// Package telereactions bridges a gap in gopkg.in/telebot.v4: the library
// exposes message_reaction update fields but does not dispatch them to a
// Handle() endpoint. Wrap attaches a filter to a LongPoller that diffs each
// update's OldReaction/NewReaction and invokes per-emoji add/remove callbacks.
package telereactions

import (
	tele "gopkg.in/telebot.v4"
)

// AllowedUpdate is the Telegram update type this package consumes. Wrap
// ensures it is present in the inner poller's AllowedUpdates.
const AllowedUpdate = "message_reaction"

// Event describes a single emoji change on a message. One MessageReaction
// update with N changed emoji becomes N events.
type Event struct {
	Chat      *tele.Chat
	MessageID int
	User      *tele.User
	ActorChat *tele.Chat
	Emoji     string
}

// Handlers holds the callbacks Wrap invokes. Either field may be nil.
type Handlers struct {
	OnAdd    func(Event)
	OnRemove func(Event)
}

// Wrap returns a MiddlewarePoller that diffs message_reaction updates and
// fires Handlers for each changed emoji. It mutates inner.AllowedUpdates to
// include "message_reaction" if missing.
func Wrap(inner *tele.LongPoller, h Handlers) *tele.MiddlewarePoller {
	ensureAllowed(inner)
	return tele.NewMiddlewarePoller(inner, func(u *tele.Update) bool {
		dispatch(u, h)
		return true
	})
}

func ensureAllowed(p *tele.LongPoller) {
	for _, u := range p.AllowedUpdates {
		if u == AllowedUpdate {
			return
		}
	}
	p.AllowedUpdates = append(p.AllowedUpdates, AllowedUpdate)
}

func dispatch(u *tele.Update, h Handlers) {
	if u == nil || u.MessageReaction == nil {
		return
	}
	mr := u.MessageReaction
	base := Event{
		Chat:      mr.Chat,
		MessageID: mr.MessageID,
		User:      mr.User,
		ActorChat: mr.ActorChat,
	}
	if h.OnAdd != nil {
		for _, r := range mr.NewReaction {
			if !containsEmoji(mr.OldReaction, r.Emoji) {
				e := base
				e.Emoji = r.Emoji
				h.OnAdd(e)
			}
		}
	}
	if h.OnRemove != nil {
		for _, r := range mr.OldReaction {
			if !containsEmoji(mr.NewReaction, r.Emoji) {
				e := base
				e.Emoji = r.Emoji
				h.OnRemove(e)
			}
		}
	}
}

func containsEmoji(rs []tele.Reaction, emoji string) bool {
	for _, r := range rs {
		if r.Emoji == emoji {
			return true
		}
	}
	return false
}
