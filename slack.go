package slackactivity

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

const (
	getChannelsMaxPages    = 30
	getChannelsPageLimit   = 1000
	getChannelHistoryLimit = 1000
)

func GetAllUnarchivedChannels(api *slack.Client) ([]slack.Channel, error) {
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

func GetChannelHistory(api *slack.Client, id string) ([]slack.Message, error) {
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

func FilterMessage(message []slack.Message, fn func(slack.Message) bool) []slack.Message {
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

func IgnoreMessageSubType(message slack.Message) bool {
	return message.SubType != "channel_join" && message.SubType != "channel_leave"
}

const (
	timeFormat  = "2006/01/02"
	hoursPerDay = 24
)

type MessageCount struct {
	Key   string
	Count int
}

var ErrNoMessages = errors.New("no messages")

func CountMessage(message []slack.Message) ([]MessageCount, error) {
	filtered := FilterMessage(message, IgnoreMessageSubType)
	if len(filtered) == 0 {
		return nil, ErrNoMessages
	}

	oldestTime, err := SlackTimestampToTime(filtered[len(filtered)-1].Timestamp)
	if err != nil {
		return nil, err
	}

	oldestDate := time.Date(oldestTime.Year(), oldestTime.Month(), oldestTime.Day(), 0, 0, 0, 0, oldestTime.Location())
	dateNum := int(math.Ceil(time.Since(oldestDate).Hours() / hoursPerDay))
	result := make([]MessageCount, dateNum)

	for i := 0; i < dateNum; i++ {
		d := oldestDate.AddDate(0, 0, i).Format(timeFormat)
		result[i].Key = d
	}

	for _, m := range filtered {
		// ignore error
		date, _ := SlackTimestampToTime(m.Timestamp)
		diff := date.Sub(oldestDate)
		i := int(math.Floor(diff.Hours() / hoursPerDay))
		result[i].Count++
	}

	return result, nil
}

func SlackTimestampToTime(st string) (time.Time, error) {
	spl := strings.Split(st, ".")
	//nolint:gomnd
	timestamp, err := strconv.ParseInt(spl[0], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("timestamp sec parse failed: %w", err)
	}

	//nolint:gomnd
	timestampMillisec, err := strconv.ParseInt(spl[1], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("timestamp millisec parse failed: %w", err)
	}

	//nolint:gomnd
	times := time.Unix(timestamp, timestampMillisec*1000)

	return times, nil
}
