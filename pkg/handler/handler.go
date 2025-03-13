package handler

import (
	"context"
	"fmt"
	"strings"

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
}

type handler struct {
	db        db.Database
	whiteList []int64
	bot       *telebot.Bot
	log       *zap.Logger
}

func NewHandler(db db.Database, log *zap.Logger, bot *telebot.Bot, whiteList []int64) Handler {
	return &handler{
		db:        db,
		log:       log,
		bot:       bot,
		whiteList: whiteList,
	}
}

func (h *handler) Start(m *telebot.Message) {
	if !utils.InWhiteList(m.Sender.ID, h.whiteList) {
		h.log.Info("Unauthorized user", zap.Int64("user_id", m.Sender.ID))
		return
	}

	h.bot.Reply(m, "Hello! I'm the album queue bot. Send me a Spotify album link and I'll add it to the download queue.")
}

func (h *handler) HandleText(m *telebot.Message) {
	if !utils.InWhiteList(m.Sender.ID, h.whiteList) {
		h.log.Info("Unauthorized user", zap.Int64("user_id", m.Sender.ID))
		return
	}

	h.log.Info("Received message", zap.Any("message", m.Text))

	// Check if the message is a valid Spotify URL
	if !utils.IsValidSpotifyURL(m.Text) {
		h.bot.Reply(m, "Invalid Spotify URL. Please send a valid Spotify album link.")
		return
	}

	// TODO: get name from spotify

	// Add the download request to the database
	err := h.db.NewDownloadRequest(context.Background(), m.Text, "", m.Sender.ID)
	if err != nil {
		h.log.Error("Failed to add download request to database", zap.Error(err))
		h.bot.Reply(m, "Failed to add download request to database. Please try again later.")
		return
	}

	h.bot.Reply(m, "Download request added to the queue.")
}

func (h *handler) HandleQueue(m *telebot.Message) {
	if !utils.InWhiteList(m.Sender.ID, h.whiteList) {
		h.log.Info("Unauthorized user", zap.Int64("user_id", m.Sender.ID))
		return
	}

	requests, err := h.db.GetActiveRequests(context.Background())
	if err != nil {
		h.log.Error("Failed to get active download requests", zap.Error(err))
		h.bot.Reply(m, "Failed to get active download requests. Please try again later.")
		return
	}

	if len(requests) == 0 {
		h.bot.Reply(m, "No active download requests.")
		return
	}

	response := "Active download requests:\n"
	for _, r := range requests {
		response += fmt.Sprintf("%s: %s - %s\n", r.ID, r.Name, r.SpotifyURL)
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
		h.bot.Reply(m, "Invalid command. Please use /deactivate <request_id>.")
		return
	}

	id := s[1]
	h.log.Info("Deactivating request", zap.String("id", id))

	err := h.db.DeactivateRequest(context.Background(), id)
	if err != nil {
		h.log.Error("Failed to deactivate request", zap.Error(err))
		h.bot.Reply(m, "Failed to deactivate request. Please try again later.")
		return
	}

	h.bot.Reply(m, "Request deactivated.")
}
