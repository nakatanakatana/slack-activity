package slackactivity

import (
	"github.com/slack-go/slack"
)

type SlackPostMessageClient interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}

var _ SlackPostMessageClient = &slack.Client{}
