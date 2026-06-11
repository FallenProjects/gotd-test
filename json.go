package main

import (
	"encoding/json"

	"github.com/AshokShau/gotdbot"
)

func printJsonHandler(c *gotdbot.Client, update gotdbot.TlObject) error {
	var chatID int64
	var messageID int64
	var isOutgoing bool

	switch u := update.(type) {
	case *gotdbot.UpdateNewMessage:
		chatID = u.Message.ChatId
		messageID = u.Message.Id
		isOutgoing = u.Message.IsOutgoing
	case *gotdbot.UpdateNewBusinessMessage:
		chatID = u.Message.Message.ChatId
		messageID = u.Message.Message.Id
		isOutgoing = u.Message.Message.IsOutgoing
	case *gotdbot.UpdateBusinessMessageEdited:
		chatID = u.Message.Message.ChatId
		messageID = u.Message.Message.Id
		isOutgoing = u.Message.Message.IsOutgoing
	case *gotdbot.UpdateNewGuestQuery:
		data, _ := json.MarshalIndent(update, "", "  ")
		if err := sendResult(c, u.Id, string(data)); err != nil {
			c.Logger.Warnf("Failed to send JSON for guest query: %v", err)
		}
		return nil
	}

	if isOutgoing {
		return nil
	}

	if chatID != 0 && !isDebugEnabled(chatID) {
		return nil
	}

	data, marshalErr := json.MarshalIndent(update, "", "  ")
	if marshalErr != nil {
		c.Logger.Debugf("Failed to marshal update: %v", marshalErr)
		return nil
	}

	jsonStr := string(data)
	if chatID == 0 {
		c.Logger.Debugf("type=%s\n%s", update.GetType(), jsonStr)
		return nil
	}

	if err := sendJSON(c, chatID, messageID, jsonStr); err != nil {
		c.Logger.Warnf("Failed to send JSON: %v", err)
	}

	return nil
}
