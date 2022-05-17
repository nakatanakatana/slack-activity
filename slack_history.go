package slackactivity

import (
	"fmt"

	"github.com/slack-go/slack"
)

type SlackChannelHistoryClient interface {
	GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)
}

var _ SlackChannelHistoryClient = &slack.Client{}

const (
	getChannelHistoryLimit = 1000
)

func GetChannelHistory(
	api SlackChannelHistoryClient,
	id string,
) ([]slack.Message, error) {
	params := &slack.GetConversationHistoryParameters{
		ChannelID: id,
		Limit:     getChannelHistoryLimit,
		Cursor:    "",
		Inclusive: false,
		Latest:    "",
		Oldest:    "",
	}

	result, err := api.GetConversationHistory(params)
	if err != nil {
		return nil, fmt.Errorf("GetConversationHistory failed: %w", err)
	}

	return result.Messages, nil
}

func FilterMessage(
	message []slack.Message,
	fn func(slack.Message) bool,
) []slack.Message {
	count := 0

	for _, m := range message {
		if fn(m) {
			count++
		}
	}

	result := make([]slack.Message, count)
	count = 0

	for _, m := range message {
		if fn(m) {
			result[count] = m
			count++
		}
	}

	return result
}
