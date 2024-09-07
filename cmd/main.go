package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"uncord-bot-go/config"
	"uncord-bot-go/handlers"

	"uncord-bot-go/lavalink"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgolink/v3/disgolink"
)

func main() {
	slog.Info("Starting uncord-bot-go...")

	// Load enviornment variables from .env file (for local development)
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file", "err", err)
	}

	config.LoadConfig() // load configuration
	config.ConnectDB()  // connect to database

	// Create the Disgo client with the appropriate intents and event listener
	client, err := disgo.New(config.AppConfig.DiscordToken,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuildMessages,         // Listen for guild message events
				gateway.IntentMessageContent,        // Listen for message content (for reading message text)
				gateway.IntentGuildMessageReactions, // Listen for reactions on messages
				gateway.IntentGuildVoiceStates,      // Listen for voice state updates
			),
		),
		bot.WithEventListenerFunc(handlers.OnReactionAdd),
		bot.WithEventListenerFunc(handlers.OnMessageCreate),
		bot.WithEventListenerFunc(handlers.OnReactionRemove),
	)
	if err != nil {
		slog.Error("Error while building client", slog.Any("err", err))
		return
	}

	// Initialize Lavalink client
	lavalinkClient, err := lavalink.NewClient(disgolink.NodeConfig{
		Name:     "local",
		Address:  "localhost:2333",
		Password: "youshallnotpass",
		Secure:   false,
	}, client)

	if err != nil {
		slog.Error("Error initializing Lavalink client", slog.Any("err", err))
		return
	}

	handler := handlers.NewHandler(lavalinkClient, client)
	client.AddEventListeners(handler)

	defer client.Close(context.TODO())

	if err = client.OpenGateway(context.TODO()); err != nil {
		slog.Error("Error connecting to Discord gateway", slog.Any("err", err))
		return
	}

	slog.Info("uncord-bot-go is now running. Press CTRL-C to exit.")

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
