package platforms

import (
	"context"
	"log/slog"

	"github.com/napuu/gpsp-bot/internal/chain"
	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/napuu/gpsp-bot/internal/handlers"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func wrapMatrixHandler(client *mautrix.Client, chain *chain.HandlerChain) func(ctx context.Context, evt *event.Event) {
	return func(ctx context.Context, evt *event.Event) {
		slog.Debug("Matrix event received", "type", evt.Type, "sender", evt.Sender)
		
		// Ignore messages from the bot itself
		if evt.Sender == id.UserID(config.FromEnv().MATRIX_USER_ID) {
			slog.Debug("Ignoring message from bot itself")
			return
		}

		// Only handle room messages
		if evt.Type != event.EventMessage {
			slog.Debug("Ignoring non-message event", "type", evt.Type)
			return
		}

		slog.Debug("Processing Matrix message")
		
		// Wrap the context
		var clientInterface interface{} = client
		var evtInterface interface{} = evt
		chain.Process(&handlers.Context{
			MatrixClient: &clientInterface,
			MatrixEvent:  &evtInterface,
			Service:      handlers.Matrix,
		})
	}
}

func RunMatrixBot() {
	cfg := config.FromEnv()
	
	if cfg.MATRIX_TOKEN == "" {
		slog.Error("MATRIX_TOKEN is required")
		panic("MATRIX_TOKEN is required")
	}
	if cfg.MATRIX_HOMESERVER == "" {
		slog.Error("MATRIX_HOMESERVER is required")
		panic("MATRIX_HOMESERVER is required")
	}
	if cfg.MATRIX_USER_ID == "" {
		slog.Error("MATRIX_USER_ID is required")
		panic("MATRIX_USER_ID is required")
	}

	client, err := mautrix.NewClient(cfg.MATRIX_HOMESERVER, id.UserID(cfg.MATRIX_USER_ID), cfg.MATRIX_TOKEN)
	if err != nil {
		slog.Error("Error creating Matrix client", "error", err)
		panic(err)
	}

	client.Store = mautrix.NewMemorySyncStore()

	// Create the chain of responsibility
	chain := chain.NewChainOfResponsibility()

	// Set up the handler
	syncer := mautrix.NewDefaultSyncer()
	syncer.OnEventType(event.EventMessage, wrapMatrixHandler(client, chain))
	client.Syncer = syncer

	// Get the current sync token to skip old messages
	slog.Info("Getting current sync state to skip old messages...")
	resp, err := client.SyncRequest(context.Background(), 0, "", "", false, event.PresenceOnline)
	if err != nil {
		slog.Warn("Could not get initial sync state", "error", err)
	} else {
		// Store the next_batch token so we start from "now"
		client.Store.SaveNextBatch(context.Background(), id.UserID(cfg.MATRIX_USER_ID), resp.NextBatch)
		slog.Info("Skipped old messages, starting from current state", "next_batch", resp.NextBatch)
	}

	slog.Info("Now listening for new messages...")

	// Start syncing - will only process new messages from this point
	err = client.Sync()
	if err != nil {
		slog.Error("Error syncing Matrix client", "error", err)
		panic(err)
	}
}
