package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/AshokShau/gotdbot"
)

const (
	adminID = int64(5938660179)
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

	_, err = c.SendDocument(chatID, gotdbot.InputFileLocal{Path: tmpFile.Name()}, &gotdbot.SendDocumentOpts{ReplyToMessageID: replyToID})
	return err
}

func sendResult(c *gotdbot.Client, queryID int64, output string) error {
	language := "json"
	if len([]rune(output)) <= maxMessageLen {
		escaped := gotdbot.EscapeHTML(output)
		text := "<pre language=\"" + language + "\">" + escaped + "</pre>"
		ftext, err := c.GetFormattedText(text, nil, gotdbot.ParseModeHTML)
		if err != nil {
			return fmt.Errorf("get formatted text: %w", err)
		}

		result := gotdbot.InputInlineQueryResultArticle{
			Id:    fmt.Sprintf("json-%d", time.Now().UnixNano()),
			Title: "JSON Result",
			InputMessageContent: &gotdbot.InputMessageText{
				Text: ftext,
			},
		}

		_, err = c.AnswerGuestQuery(queryID, result)
		return err
	}

	tmpFile, err := os.CreateTemp("", "guest-query-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.WriteString(output); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}

	_ = tmpFile.Close()
	msg, err := c.SendDocument(adminID, gotdbot.InputFileLocal{Path: tmpFile.Name()}, &gotdbot.SendDocumentOpts{Caption: fmt.Sprintf("JSON result for guest query %d", queryID)})
	if err != nil {
		return fmt.Errorf("upload document: %w", err)
	}

	result := &gotdbot.InputInlineQueryResultDocument{
		Id:          strconv.FormatInt(time.Now().UnixNano(), 10),
		Title:       "JSON Result",
		Description: "Large JSON output",
		DocumentUrl: msg.RemoteFileID(),
		InputMessageContent: &gotdbot.InputMessageDocument{
			Document: &gotdbot.InputDocument{Document: gotdbot.InputFileRemote{Id: msg.RemoteFileID()}},
			Caption: &gotdbot.FormattedText{
				Text: fmt.Sprintf("JSON result for guest query %d", queryID),
			},
		},
	}

	_, err = c.AnswerGuestQuery(queryID, result)
	if err != nil {
		c.Logger.Warnf("failed to answer guest query: %v", err)
		return fmt.Errorf("answer guest query: %w", err)
	}

	return nil
}
