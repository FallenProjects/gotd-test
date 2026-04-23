package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/AshokShau/gotdbot"
)

func sendJSON(c *gotdbot.Client, chatID, replyToID int64, output string) error {
	language := "json"
	if len([]rune(output)) <= maxMessageLen {
		escaped := gotdbot.EscapeHTML(output)
		text := "<pre language=\"" + language + "\">" + escaped + "</pre>"
		_, err := c.SendTextMessage(chatID, text, &gotdbot.SendTextMessageOpts{
			ParseMode:        gotdbot.ParseModeHTML,
			ReplyToMessageID: replyToID,
		})
		return err
	}

	tmpFile, err := os.CreateTemp("", "update-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.WriteString(output); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	docOpts := &gotdbot.SendDocumentOpts{
		ReplyToMessageID: replyToID,
	}
	_, err = c.SendDocument(chatID, gotdbot.InputFileLocal{Path: tmpFile.Name()}, docOpts)
	return err
}

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

	if err := sendJSON(c, chatID, 0, jsonStr); err != nil {
		log.Printf("[ERROR] Failed to send JSON: %v", err)
	}

	return nil
}
