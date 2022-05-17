package slackactivity

import (
	"fmt"

	"github.com/slack-go/slack"
)

const (
	getChannelsMaxPages  = 30
	getChannelsPageLimit = 1000
)

type SlackGetChannelsClient interface {
	GetConversations(
		params *slack.GetConversationsParameters,
	) (channels []slack.Channel, nextCursor string, err error)
}

var _ SlackGetChannelsClient = &slack.Client{}

func GetAllUnarchivedChannels(
	api SlackGetChannelsClient,
) ([]slack.Channel, error) {
	params := &slack.GetConversationsParameters{
		Limit:           getChannelsPageLimit,
		ExcludeArchived: true,
		Types:           []string{"public_channel"},
		Cursor:          "",
		TeamID:          "",
	}
	channels := make([]slack.Channel, 0)

	for i := 0; i < getChannelsMaxPages; i++ {
		result, cursor, err := api.GetConversations(params)
		if err != nil {
			return nil, fmt.Errorf("GetConversation failed: %w", err)
		}

		params.Cursor = cursor

		channels = append(channels, result...)

		if cursor == "" {
			break
		}
	}

	return channels, nil
}
