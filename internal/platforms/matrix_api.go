package platforms

import (
	"context"

	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/handlers"
	"golang.org/x/exp/slog"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func wrapMatrixHandler(client *mautrix.Client, chain *chain.HandlerChain) func(ctx context.Context, evt *event.Event) {
	return func(ctx context.Context, evt *event.Event) {
		if evt.Sender == client.UserID {
			return // Ignore messages from the bot itself
		}
		chain.Process(&handlers.Context{
			MatrixEvent:  evt,
			MatrixClient: client,
			Service:      handlers.Matrix,
		})
	}
}

func RunMatrixBot() {
	client := getMatrixClient()
	chain := chain.NewChainOfResponsibility()

	syncer := client.Syncer.(*mautrix.DefaultSyncer)

	// Handle text messages
	syncer.OnEventType(event.EventMessage, wrapMatrixHandler(client, chain))

	// Optional: Auto-join on invites
	syncer.OnEventType(event.StateMember, func(ctx context.Context, evt *event.Event) {
		if evt.GetStateKey() == client.UserID.String() && evt.Content.AsMember().Membership == event.MembershipInvite {
			// _, err := client.JoinRoomByID(ctx, evt.RoomID)
			client.JoinRoomByID(ctx, evt.RoomID)
			// if err != nil {
			// 	slog.Error("Failed to join room", slog.String("room_id", evt.RoomID.String()), slog.Error(err))
			// } else {
			// 	slog.Info("Joined room", slog.String("room_id", evt.RoomID.String()))
			// }
		}
	})

	slog.Info("Starting Matrix bot...")
	err := client.Sync()
	if err != nil {
		slog.Error("Sync failed")
	}
}

func getMatrixClient() *mautrix.Client {
	cfg := config.FromEnv()
	// client, err := mautrix.NewClient(cfg.MATRIX_HOMESERVER, "gpsp-bot", "")
	client, err := mautrix.NewClient(cfg.MATRIX_HOMESERVER, "", "")
	if err != nil {
		panic(err)
	}

	// Logging
	// client.Log = slog.Default().WithGroup("matrix")

	// Login
	resp, err := client.Login(context.TODO(), &mautrix.ReqLogin{
		Type:       mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: "gpsp-bot"},
		Password:   "",
	})
	if err != nil {
		panic(err)
	}

	client.SetCredentials(resp.UserID, resp.AccessToken)
	return client
}
