package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"uncord-bot-go/config"
	"uncord-bot-go/handlers"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
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
		bot.WithCacheConfigOpts(
			cache.WithCaches(
				cache.CacheGuildVoiceStates,
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
	b.Client = client

	b.Lavalink = disgolink.New(client.ApplicationID())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = client.OpenGateway(ctx); err != nil {
		slog.Error("Error connecting to Discord gateway", slog.Any("err", err))
		return
	}
	defer client.Close(context.TODO())

	// Register commands after connecting to the gateway
	if err = b.RegisterCommands(client); err != nil {
		slog.Error("Failed to register commands", slog.Any("err", err))
		return
	}

	b.RegisterGuildCommands(client, snowflake.ID(1112943203755233350))

	node, err := b.Lavalink.AddNode(ctx, disgolink.NodeConfig{
		Name:     "local",
		Address:  "localhost:2333",
		Password: "youshallnotpass",
		Secure:   false,
	})
	if err != nil {
		slog.Error("Failed to add node", slog.Any("err", err))
		return
	}

	version, err := node.Version(ctx)
	if err != nil {
		slog.Error("Failed to get node version", slog.Any("err", err))
		return
	}

	slog.Info("uncord-bot-go is now running. Press CTRL-C to exit.", slog.String("node_version", version), slog.String("node_session_id", node.SessionID()))

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
