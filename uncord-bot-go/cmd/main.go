package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"uncord-bot-go/handlers"  // Local module import for handlers
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
)

func main() {
	// Log starting the bot
	slog.Info("Starting uncord-bot-go...")

	// Get the token from the environment variable
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		slog.Error("No token provided. Set the DISCORD_TOKEN environment variable.")
		return
	}

	// Create the Disgo client with the appropriate intents and event listener
	client, err := disgo.New(token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuildMessages,    // Listen for guild message events
				gateway.IntentMessageContent,   // Listen for message content (for reading message text)
			),
		),
		bot.WithEventListenerFunc(handlers.OnMessageCreate), // Correct function name
	)
	if err != nil {
		slog.Error("Error while building disgo client", slog.Any("err", err))
		return
	}

	// Ensure the client closes gracefully
	defer client.Close(context.TODO())

	// Connect to the Discord gateway
	if err = client.OpenGateway(context.TODO()); err != nil {
		slog.Error("Error connecting to Discord gateway", slog.Any("err", err))
		return
	}

	// Bot is now running
	slog.Info("uncord-bot-go is now running. Press CTRL-C to exit.")

	// Listen for termination signals to shut down gracefully
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}