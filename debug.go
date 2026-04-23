package main

import (
	"strings"
	"sync"

	"github.com/AshokShau/gotdbot"
)

var chatDebugState = struct {
	mu    sync.RWMutex
	state map[int64]bool
}{
	state: make(map[int64]bool),
}

func isDebugEnabled(chatID int64) bool {
	if chatID == 0 {
		return true
	}

	chatDebugState.mu.RLock()
	enabled, ok := chatDebugState.state[chatID]
	chatDebugState.mu.RUnlock()
	if !ok {
		return true
	}
	return enabled
}

func setDebugEnabled(chatID int64, enabled bool) {
	if chatID == 0 {
		return
	}

	chatDebugState.mu.Lock()
	chatDebugState.state[chatID] = enabled
	chatDebugState.mu.Unlock()
}

func debugCommandHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	m := ctx.EffectiveMessage
	chatID := ctx.EffectiveChatId
	parts := strings.Fields(m.GetText())
	if len(parts) < 2 {
		_, _ = m.ReplyText(c, "Usage: /debug on|off", nil)
		return gotdbot.EndGroups
	}

	switch strings.ToLower(parts[1]) {
	case "on":
		setDebugEnabled(chatID, true)
		_, _ = m.ReplyText(c, "Debug logger is now ON for this chat.", nil)
	case "off":
		setDebugEnabled(chatID, false)
		_, _ = m.ReplyText(c, "Debug logger is now OFF for this chat.", nil)
	default:
		_, _ = m.ReplyText(c, "Usage: /debug on|off", nil)
	}

	return gotdbot.EndGroups
}
