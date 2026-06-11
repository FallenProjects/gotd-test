package main

//go:generate go run github.com/AshokShau/gotdbot/scripts/tools

import (
	"log"
	"strconv"

	"github.com/AshokShau/gotdbot"
)

func main() {
	apiID, err := strconv.Atoi(ApiId)
	if err != nil {
		log.Fatalln(err)
	}

	bot, err := gotdbot.NewClient(int32(apiID), ApiHash, Token, &gotdbot.ClientOpts{
		LibraryPath: "./libtdjson.so.1.8.65",
	})

	if err != nil {
		log.Fatalf("Failed to create bot client: %v", err)
	}

	bot.OnCommand("eval", evalCommandHandler)
	bot.OnCommand("debug", debugCommandHandler)
	bot.AddHandler(&catchAllHandler{fn: printJsonHandler})

	err = bot.Start()
	if err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}

	me := bot.Me
	username := ""
	if me.Usernames != nil {
		username = me.Usernames.EditableUsername
	}

	bot.Logger.Info("Logged in", "username", username, "id", me.Id)
	bot.Idle()
}

type catchAllHandler struct {
	fn func(*gotdbot.Client, gotdbot.TlObject) error
}

func (h *catchAllHandler) CheckUpdate(_ *gotdbot.Client, _ gotdbot.TlObject) bool { return true }
func (h *catchAllHandler) HandleUpdate(c *gotdbot.Client, update gotdbot.TlObject) error {
	return h.fn(c, update)
}
