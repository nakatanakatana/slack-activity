package slackactivity

import (
	"github.com/slack-go/slack"
)

type (
	MessageTimestamp string
	FilePermalink    string
)

type SlackPostClient interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	UploadFile(params slack.FileUploadParameters) (file *slack.File, err error)
}

var _ SlackPostClient = &slack.Client{}
