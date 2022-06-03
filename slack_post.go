package slackactivity

import (
	"fmt"

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

func CreatePostReportMessageAttachements(
	channel slack.Channel,
	imageURL string,
	lastMessageTime string,
) slack.Attachment {
	return slack.Attachment{
		Title:    fmt.Sprintf("<#%s>", channel.ID),
		ImageURL: imageURL,
		Fields: []slack.AttachmentField{
			{
				Title: "Creator",
				Value: fmt.Sprintf("<@%s>", channel.Creator),
				Short: true,
			},
			{
				Title: "LastMessageTime",
				Value: lastMessageTime,
				Short: true,
			},
		},
	}
}
