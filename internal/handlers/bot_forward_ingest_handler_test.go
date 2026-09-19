package handlers

import (
	"testing"

	"github.com/napuu/gpsp-bot/pkg/utils"
	tele "gopkg.in/telebot.v4"
)

func TestIsForwardedFromBot(t *testing.T) {
	const botID int64 = 12345
	const otherID int64 = 99999

	tests := []struct {
		name string
		msg  *tele.Message
		want bool
	}{
		{
			name: "nil message",
			msg:  nil,
			want: false,
		},
		{
			name: "not a forward",
			msg:  &tele.Message{},
			want: false,
		},
		{
			name: "legacy OriginalSender is bot",
			msg: &tele.Message{
				OriginalSender: &tele.User{ID: botID},
			},
			want: true,
		},
		{
			name: "legacy OriginalSender is other user",
			msg: &tele.Message{
				OriginalSender: &tele.User{ID: otherID},
			},
			want: false,
		},
		{
			name: "Origin.Sender is bot",
			msg: &tele.Message{
				Origin: &tele.MessageOrigin{
					Type:   "user",
					Sender: &tele.User{ID: botID},
				},
			},
			want: true,
		},
		{
			name: "Origin.Sender is other user",
			msg: &tele.Message{
				Origin: &tele.MessageOrigin{
					Type:   "user",
					Sender: &tele.User{ID: otherID},
				},
			},
			want: false,
		},
		{
			name: "hidden attribution has name only",
			msg: &tele.Message{
				OriginalSenderName: "Some Bot",
				Origin: &tele.MessageOrigin{
					Type:           "hidden_user",
					SenderUsername: "Some Bot",
				},
			},
			want: false,
		},
		{
			name: "Origin.Sender is bot even when OriginalSender differs",
			msg: &tele.Message{
				OriginalSender: &tele.User{ID: otherID},
				Origin: &tele.MessageOrigin{
					Type:   "user",
					Sender: &tele.User{ID: botID},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isForwardedFromBot(tt.msg, botID)
			if got != tt.want {
				t.Fatalf("isForwardedFromBot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasSurvivingMatch(t *testing.T) {
	tests := []struct {
		name       string
		pendingOCR string
		matches    []utils.FingerprintMatch
		want       bool
	}{
		{
			name:    "no matches",
			matches: nil,
			want:    false,
		},
		{
			name: "visual match without OCR",
			matches: []utils.FingerprintMatch{
				{MessageID: "1"},
			},
			want: true,
		},
		{
			name:       "OCR hashes differ — filtered out",
			pendingOCR: "aaa",
			matches: []utils.FingerprintMatch{
				{MessageID: "1", OCRTextHash: "bbb"},
			},
			want: false,
		},
		{
			name:       "OCR hashes match — survives",
			pendingOCR: "aaa",
			matches: []utils.FingerprintMatch{
				{MessageID: "1", OCRTextHash: "aaa"},
			},
			want: true,
		},
		{
			name:       "one OCR mismatch then one survival",
			pendingOCR: "aaa",
			matches: []utils.FingerprintMatch{
				{MessageID: "1", OCRTextHash: "bbb"},
				{MessageID: "2", OCRTextHash: "aaa"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasSurvivingMatch(tt.matches, tt.pendingOCR)
			if got != tt.want {
				t.Fatalf("hasSurvivingMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldIngestBotForward_RequiresDlFeature(t *testing.T) {
	t.Setenv("ENABLED_FEATURES", "ping")

	msg := &tele.Message{
		Chat: &tele.Chat{Type: tele.ChatGroup},
		Video: &tele.Video{
			File: tele.File{FileID: "file1"},
		},
		OriginalSender: &tele.User{ID: 42},
	}
	bot := &tele.Bot{Me: &tele.User{ID: 42}}

	ctx := &Context{
		Service:        Telegram,
		Telebot:        bot,
		TelebotContext: &stubTeleContext{msg: msg},
	}
	if shouldIngestBotForward(ctx) {
		t.Fatal("expected ingest to be disabled when dl is not in ENABLED_FEATURES")
	}

	t.Setenv("ENABLED_FEATURES", "ping;dl")
	if !shouldIngestBotForward(ctx) {
		t.Fatal("expected ingest when dl is enabled and message is a same-bot group forward with video")
	}
}

// stubTeleContext implements the small tele.Context surface used by shouldIngestBotForward.
type stubTeleContext struct {
	tele.Context
	msg *tele.Message
}

func (s *stubTeleContext) Message() *tele.Message { return s.msg }
func (s *stubTeleContext) Chat() *tele.Chat {
	if s.msg == nil {
		return nil
	}
	return s.msg.Chat
}
