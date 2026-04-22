package main

import (
	"bytes"
	"context"
	"fmt"
	"go/build"
	"io"
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
	parts := strings.SplitN(text, " ", 2)
	var code string
	if len(parts) > 1 {
		code = strings.TrimSpace(parts[1])
	}

	if code == "" {
		_, _ = m.ReplyText(c, "No code provided. Usage: /eval <code ", nil)
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
		fmt.Println("Error using stdlib symbols:", err)
	}

	if err := i.Use(Symbols); err != nil {
		fmt.Println("Error using custom symbols:", err)
	}

	customSymbols := map[string]map[string]reflect.Value{
		"eval/eval": {
			"C":       reflect.ValueOf(c),
			"Client":  reflect.ValueOf(c),
			"Ctx":     reflect.ValueOf(ctx),
			"Context": reflect.ValueOf(ctx),
			"M":       reflect.ValueOf(m),
			"Message": reflect.ValueOf(m),
		},
	}

	if err := i.Use(customSymbols); err != nil {
		fmt.Println("failed to use custom eval symbols:", err)
	}

	if !strings.Contains(code, "package ") && !strings.Contains(code, "func ") {
		code = fmt.Sprintf(`package main
import (
	e "eval/eval"
	"fmt"
	"os"
	"github.com/AshokShau/gotdbot"
)

func runSnippet() (res any) {
	c, client, Client := e.C, e.C, e.C
	ctx, Context := e.Ctx, e.Ctx
	m, message, Message := e.M, e.M, e.M

	_ = c; _ = client; _ = Client
	_ = ctx; _ = Context
	_ = m; _ = message; _ = Message

	%s

	return res
}

func main() {
	if res := runSnippet(); res != nil {
		fmt.Println(res)
	}
}`, code)
	}

	_, err := i.EvalWithContext(context.Background(), code)
	if err != nil {
		_, err := m.ReplyText(c, fmt.Sprintf("<b>#EVALERR:</b> <code>%s</code>", gotdbot.EscapeHTML(err.Error())), &gotdbot.SendTextMessageOpts{
			ParseMode: gotdbot.ParseModeHTML,
		})

		if err != nil {
			log.Printf("Failed to send error message: %v", err)
			return gotdbot.EndGroups
		}

		return gotdbot.EndGroups
	}

	var output string
	if stdout.Len() > 0 {
		output = stdout.String()
	}
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	if strings.TrimSpace(output) == "" {
		output = "No Output"
	}

	if len(output) > maxMessageLen {
		file, _ := os.Create("output.txt")
		defer file.Close()
		_, _ = io.WriteString(file, output)
		defer os.Remove(file.Name())

		_, err := m.ReplyDocument(c, &gotdbot.InputFileLocal{Path: file.Name()}, &gotdbot.SendDocumentOpts{
			Caption: "Output",
		})

		if err != nil {
			log.Printf("Failed to send document: %v", err)
			return gotdbot.EndGroups
		}

		return gotdbot.EndGroups
	}

	_, err = m.ReplyText(c, fmt.Sprintf("<b>#EVALOut:</b>\n<code>%s</code>", gotdbot.EscapeHTML(strings.TrimSpace(output))), &gotdbot.SendTextMessageOpts{
		ParseMode: gotdbot.ParseModeHTML,
	})

	if err != nil {
		log.Printf("Failed to send output message: %v", err)
		return gotdbot.EndGroups
	}

	return gotdbot.EndGroups
}
