package handlers

import (
	"testing"

	"github.com/napuu/gpsp-bot/internal/dayvideo"
)

type noopHandler struct{}

func (noopHandler) Execute(*Context)       {}
func (noopHandler) SetNext(ContextHandler) {}

func TestDayVideoHandlerSkipsWhenSchedulerNil(t *testing.T) {
	h := NewDayVideoHandler(nil)
	h.SetNext(noopHandler{})
	h.Execute(&Context{Service: Discord, chatId: "123", guildId: "guild"})
}

func TestDayVideoHandlerSkipsDiscordDM(t *testing.T) {
	t.Setenv("ENABLED_FEATURES", "daymeme")
	sched := dayvideo.NewSchedulerForTest("", nil, nil, nil, nil, nil, 0)
	h := NewDayVideoHandler(sched)
	h.SetNext(noopHandler{})
	h.Execute(&Context{Service: Discord, chatId: "123", guildId: ""})
}
