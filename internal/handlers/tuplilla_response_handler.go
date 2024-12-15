package handlers

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/napuu/gpsp-bot/pkg/utils"
	"golang.org/x/exp/rand"
	tele "gopkg.in/telebot.v4"
)

type TuplillaResponseHandler struct {
	next ContextHandler
}

func (r *TuplillaResponseHandler) Execute(m *Context) {
	slog.Debug("Entering TuplillaResponseHandler")
	if m.action == Tuplilla {
		switch m.Service {
		case Telegram:
			chatId := tele.ChatID(utils.S2I(m.chatId))

			cube1Response, err := m.Telebot.Send(chatId, tele.Cube)

			if err != nil {
				slog.Error(err.Error())
			}

			time.Sleep(2 * time.Second)

			cube2Response, err := m.Telebot.Send(chatId, tele.Cube)

			if err != nil {
				slog.Error(err.Error())
			}

			if cube1Response.Dice.Value == cube2Response.Dice.Value {
				m.gotDubz = true
			}

			m.lastCubeThrownTime = time.Now()
			m.dubzNegation = make(chan string)
			go func() {
				m.dubzNegation <- utils.GetNegation(m.parsedText)
			}()

		case Discord:
			message := &discordgo.MessageReference{
				ChannelID: m.chatId,
				MessageID: m.id,
			}

			cube1 := rand.Intn(6) + 1
			cube2 := rand.Intn(6) + 1

			dubzNegation := make(chan string)
			go func() {
				if cube1 == cube2 {
					dubzNegation <- m.parsedText
				} else {
					dubzNegation <- utils.GetNegation(m.parsedText)
				}
			}()
			msgContent := fmt.Sprintf("Noppa 1: %d", cube1)
			msg, _ := m.DiscordSession.ChannelMessageSendReply(m.chatId, msgContent, message)

			// The reply above aborts the typing
			m.SendTyping()

			time.Sleep(2 * time.Second)

			msgContent += "\n" + fmt.Sprintf("Noppa 2: %d", cube2)
			m.DiscordSession.ChannelMessageEdit(m.chatId, msg.ID, msgContent)
			time.Sleep(2 * time.Second)

			finalMessage := <-dubzNegation
			if cube1 == cube2 {
				msgContent += "\nTuplat tuli 😎, "
			} else {
				msgContent += "\nEi tuplia 😿, "
			}
			msgContent += finalMessage
			m.DiscordSession.ChannelMessageEdit(m.chatId, msg.ID, msgContent)

		}
	}

	r.next.Execute(m)
}

func (u *TuplillaResponseHandler) SetNext(next ContextHandler) {
	u.next = next
}
