package main

import (
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

var (
	Token         = os.Getenv("TOKEN")
	ApiId         = os.Getenv("API_ID")
	ApiHash       = os.Getenv("API_HASH")
	devIDs        = parseDevIDs(os.Getenv("DEV_IDS"))
	maxMessageLen = 4000
)

func parseDevIDs(list string) []int64 {
	if list == "" {
		return nil
	}

	var ids []int64
	for _, s := range strings.Split(list, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			log.Printf("invalid DEV_ID: %s", s)
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

var Symbols = map[string]map[string]reflect.Value{}

func IsDev(userID int64) bool {
	for _, id := range devIDs {
		if id == userID {
			return true
		}
	}
	return false
}
