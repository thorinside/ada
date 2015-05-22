package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/guregu/kami"
	"github.com/kyokomi/slackbot"
	"github.com/kyokomi/slackbot/plugins"
	_ "github.com/thorinside/ada/plugins/adabot"
	"golang.org/x/net/context"
)

func main() {
	var token string
	flag.StringVar(&token, "token", os.Getenv("SLACK_BOT_TOKEN"), "SlackBot Token")
	flag.Parse()

	ctx := plugins.Context()

	c := slackbot.DefaultConfig()
	c.Name = "ada"
	c.SlackToken = token
	slackbot.WebSocketRTM(ctx, c)

	kami.Get("/", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	kami.Serve()
}
