package main

import (
	"context"
	"net/http"

	"github.com/supperdoggy/SmartHomeServer/music-services/album-queue/pkg/config"
	"github.com/supperdoggy/SmartHomeServer/music-services/album-queue/pkg/db"
	"github.com/supperdoggy/SmartHomeServer/music-services/album-queue/pkg/handler"
	"go.uber.org/zap"
	"gopkg.in/tucnak/telebot.v2"
)

func main() {
	ctx := context.Background()

	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	log.Info("Loaded config", zap.Any("config", cfg))

	bot, err := telebot.NewBot(telebot.Settings{
		Token: cfg.BotToken,
	})
	if err != nil {
		log.Fatal("Failed to create bot", zap.Error(err))
	}

	db, err := db.NewDatabase(ctx, log, cfg.DatabaseURL, cfg.DatabaseName)
	if err != nil {
		log.Fatal("Failed to create database connection", zap.Error(err))
	}

	// app health check api
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		http.ListenAndServe(":8080", nil)
	}()

	log.Info("Database connection established")

	h := handler.NewHandler(db, log, bot, cfg.BotWhitelist)

	bot.Handle("/start", h.Start)
	bot.Handle(telebot.OnText, h.HandleText)
	bot.Handle("/queue", h.HandleQueue)
	bot.Handle("/deactivate", h.HandleDeactivate)
	bot.Handle("/p", h.HandlePlaylist)

	log.Info("Bot is running", zap.String("username", bot.Me.Username))
	bot.Poller = &telebot.LongPoller{Timeout: 10}
	bot.Start()
}
