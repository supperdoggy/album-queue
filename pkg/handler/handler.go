package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/DigitalIndependence/models/spotify"
	"github.com/supperdoggy/SmartHomeServer/music-services/album-queue/pkg/db"
	"github.com/supperdoggy/SmartHomeServer/music-services/album-queue/pkg/utils"
	"go.uber.org/zap"
	"gopkg.in/tucnak/telebot.v2"
)

type Handler interface {
	Start(m *telebot.Message)
	HandleText(m *telebot.Message)
	HandleQueue(m *telebot.Message)
	HandleDeactivate(m *telebot.Message)
	HandlePlaylist(m *telebot.Message)
}

type handler struct {
	db             db.Database
	spotifyService spotify.SpotifyService
	whiteList      []int64
	bot            *telebot.Bot
	log            *zap.Logger
}

func NewHandler(db db.Database, log *zap.Logger, bot *telebot.Bot, spotifyService spotify.SpotifyService, whiteList []int64) Handler {
	return &handler{
		db:             db,
		log:            log,
		bot:            bot,
		whiteList:      whiteList,
		spotifyService: spotifyService,
	}
}

func (h *handler) Start(m *telebot.Message) {
	if !utils.InWhiteList(m.Sender.ID, h.whiteList) {
		h.log.Info("Unauthorized user", zap.Int64("user_id", m.Sender.ID))
		return
	}

	h.bot.Reply(m, "Привіііііііііт, я бот який кочає музіку на сєрвер, скинь мені урлу на спотік і я додам в чергу на скачування ❤️")
}

func (h *handler) HandleText(m *telebot.Message) {
	if !utils.InWhiteList(m.Sender.ID, h.whiteList) {
		h.log.Info("Unauthorized user", zap.Int64("user_id", m.Sender.ID))
		return
	}

	h.log.Info("Received message", zap.Any("message", m.Text))

	// Check if the message is a valid Spotify URL
	if !utils.IsValidSpotifyURL(m.Text) {
		h.bot.Reply(m, "о ніііііі, це не посилання на спотіфай.... 💔😭")
		return
	}

	name, err := h.spotifyService.GetObjectName(context.Background(), m.Text)
	if err != nil {
		h.log.Error("Failed to get object name from Spotify", zap.Error(err))
		h.bot.Reply(m, "не получилося дістати назву з спотіфай... 💔😭")
		return
	}

	// Add the download request to the database
	err = h.db.NewDownloadRequest(context.Background(), m.Text, name, m.Sender.ID)
	if err != nil {
		h.log.Error("Failed to add download request to database", zap.Error(err))
		h.bot.Reply(m, "не получилось додати в чергу, скажи максиму шо шось не так...")
		return
	}

	h.bot.Reply(m, "Ураураура успішно додали пісню в чергу!!!!")
}

func (h *handler) HandleQueue(m *telebot.Message) {
	if !utils.InWhiteList(m.Sender.ID, h.whiteList) {
		h.log.Info("Unauthorized user", zap.Int64("user_id", m.Sender.ID))
		return
	}

	requests, err := h.db.GetActiveRequests(context.Background())
	if err != nil {
		h.log.Error("Failed to get active download requests", zap.Error(err))
		h.bot.Reply(m, "не получилося дістати чергу... 💔😭")
		return
	}

	if len(requests) == 0 {
		h.bot.Reply(m, "немає активних запитів на скачування...")
		return
	}

	response := "Активні запити на скачування:\n"
	for _, r := range requests {
		response += fmt.Sprintf("%s: %s. Active: %v, SyncCount: %v, Errored: %v, RetryCount: %v\n", r.ID, r.Name, r.Active, r.SyncCount, r.Errored, r.RetryCount)
	}

	h.bot.Reply(m, response)
}

func (h *handler) HandleDeactivate(m *telebot.Message) {
	if !utils.InWhiteList(m.Sender.ID, h.whiteList) {
		h.log.Info("Unauthorized user", zap.Int64("user_id", m.Sender.ID))
		return
	}

	s := strings.Split(m.Text, " ")
	if len(s) != 2 {
		h.bot.Reply(m, "не розумію цю команду. Пліз юзай /deactivate <request_id>.")
		return
	}

	id := s[1]
	h.log.Info("Deactivating request", zap.String("id", id))

	err := h.db.DeactivateRequest(context.Background(), id)
	if err != nil {
		h.log.Error("Failed to deactivate request", zap.Error(err))
		h.bot.Reply(m, "не получилося деактивувати запит. Пліз спробуй ще раз пізніше.")
		return
	}

	h.bot.Reply(m, "Запит деактивовано, всьо капец.")
}

func (h *handler) HandlePlaylist(m *telebot.Message) {
	if !utils.InWhiteList(m.Sender.ID, h.whiteList) {
		h.log.Info("Unauthorized user", zap.Int64("user_id", m.Sender.ID))
		return
	}

	h.log.Info("Received playlist request", zap.Any("message", m.Text))

	// get playlist link

	msg := strings.Split(m.Text, " ")
	if len(msg) != 2 {
		h.bot.Reply(m, "не розумію цю команду. Пліз юзай /playlist <playlist_id>.")
		return
	}

	playlistURL := msg[1]

	if !utils.IsValidSpotifyURL(playlistURL) {
		h.bot.Reply(m, "о ніііііі, це не посилання на спотіфай.... 💔😭")
		return
	}

	if err := h.db.NewPlaylistRequest(context.Background(), playlistURL, m.Sender.ID); err != nil {
		h.log.Error("Failed to add playlist request to database", zap.Error(err))
		h.bot.Reply(m, "не получилось додати в чергу, скажи максиму шо шось не так...")
		return
	}

	h.bot.Reply(m, "Ураураура успішно додали плейлист в чергу!!!!")
}
