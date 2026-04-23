package main

import (
	"bytes"
	"context"
	"fmt"
	"go/build"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/AshokShau/gotdbot"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func evalCommandHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	m := ctx.EffectiveMessage
	if !IsDev(m.SenderID()) {
		return gotdbot.EndGroups
	}

	text := m.GetText()
	var code string
	if idx := strings.IndexAny(text, " \n\t"); idx > 0 {
		code = strings.TrimSpace(text[idx+1:])
	}

	if code == "" {
		_, _ = m.ReplyText(c, "No code provided. Usage: /eval &lt;code&gt;", &gotdbot.SendTextMessageOpts{ParseMode: gotdbot.ParseModeHTML})
		return gotdbot.EndGroups
	}

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = build.Default.GOPATH
	}

	var stdout, stderr bytes.Buffer
	i := interp.New(interp.Options{
		Stdout: &stdout,
		Stderr: &stderr,
		GoPath: goPath,
	})

	if err := i.Use(stdlib.Symbols); err != nil {
		log.Printf("Error using stdlib symbols: %v", err)
	}

	if err := i.Use(Symbols); err != nil {
		log.Printf("Error using custom symbols: %v", err)
	}

	customSymbols := map[string]map[string]reflect.Value{
		"eval/eval": {
			"C":   reflect.ValueOf(c),
			"Ctx": reflect.ValueOf(ctx),
			"M":   reflect.ValueOf(m),
		},
	}

	if err := i.Use(customSymbols); err != nil {
		log.Printf("Failed to use custom eval symbols: %v", err)
	}

	if !strings.Contains(code, "package ") && !strings.Contains(code, "func ") {
		code = fmt.Sprintf(`package main

import (
	e "eval/eval"
	"encoding/json"
	"fmt"
	"os"
	"github.com/AshokShau/gotdbot"
)

func runSnippet() (res any) {
	c := e.C
	ctx := e.Ctx
	m := e.M
	_ = c
	_ = ctx
	_ = m
	%s
	return res
}

func main() {
	if res := runSnippet(); res != nil {
		if data, err := json.MarshalIndent(res, "", "  "); err == nil {
			fmt.Println(string(data))
		} else {
			fmt.Println(res)
		}
	}
}`, code)
	}

	_, err := i.EvalWithContext(context.Background(), code)
	if err != nil {
		errMsg := fmt.Sprintf(
			"<pre language=\"go\">%s</pre>",
			gotdbot.EscapeHTML(err.Error()),
		)
		if _, sendErr := m.ReplyText(c, errMsg, &gotdbot.SendTextMessageOpts{
			ParseMode: gotdbot.ParseModeHTML,
		}); sendErr != nil {
			log.Printf("Failed to send error message: %v", sendErr)
		}
		return gotdbot.EndGroups
	}

	output := buildOutput(stdout, stderr)
	trimmed := strings.TrimSpace(output)

	if err = sendJSON(c, m.ChatId, m.Id, trimmed); err != nil {
		log.Printf("Failed to send output message: %v", err)
	}

	return gotdbot.EndGroups
}

func buildOutput(stdout, stderr bytes.Buffer) string {
	var sb strings.Builder
	if stdout.Len() > 0 {
		sb.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(stderr.String())
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "No Output"
	}
	return out
}
