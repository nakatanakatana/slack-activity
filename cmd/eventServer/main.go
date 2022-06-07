package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/nakatanakatana/slack-activity/report"
	"github.com/slack-go/slack"
)

const (
	APIBaseURL = "/api/v1"
)

func main() {
	botToken := os.Getenv("SLACK_TOKEN")
	slackClient := slack.New(botToken)

	cfg, err := report.CreateConfig()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	cfg.AlertThreshold = 0

	signingSecret := slackactivity.SigningSecret(os.Getenv("SLACK_SIGNING_SECRET"))
	mux := http.NewServeMux()
	verifier := slackactivity.NewSecretsVerifierMiddleware(signingSecret)
	eventHandlerFunc := createEventHandlerFunc(slackClient, slackClient, slackClient, cfg)
	mux.Handle(fmt.Sprintf("%s/events", APIBaseURL), verifier(eventHandlerFunc))

	log.Printf("starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
