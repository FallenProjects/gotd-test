package main

import (
	"encoding/json"
	"log"

	"github.com/AshokShau/gotdbot"
)

func printJsonHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	if ctx.EffectiveMessage != nil && ctx.EffectiveMessage.IsOutgoing {
		return gotdbot.EndGroups
	}

	if chatID := ctx.EffectiveChatId; chatID != 0 && !isDebugEnabled(chatID) {
		return nil
	}

	data, marshalErr := json.MarshalIndent(ctx.RawUpdate, "", "  ")
	if marshalErr != nil {
		log.Printf("[ERROR] Failed to marshal update: %v", marshalErr)
		return nil
	}

	jsonStr := string(data)
	chatID := ctx.EffectiveChatId
	if chatID == 0 {
		log.Printf("[UPDATE] type=%s\n%s", ctx.RawUpdate.GetType(), jsonStr)
		return nil
	}

	guest := ctx.Update.UpdateNewGuestQuery
	if guest != nil {
		if err := sendResult(c, guest.Id, jsonStr); err != nil {
			log.Printf("[ERROR] Failed to send JSON for guest query: %v", err)
		}
		return nil
	}

	if err := sendJSON(c, chatID, 0, jsonStr); err != nil {
		log.Printf("[ERROR] Failed to send JSON: %v", err)
	}

	return nil
}
