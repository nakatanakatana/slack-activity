package slackActivity

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

var SlackAPI *slack.Client

func init() {
	token := os.Getenv("SLACK_TOKEN")
	SlackAPI = slack.New(token)
}

func createSlackClient(token string) *slack.Client {
	return slack.New(token)
}

func GetAllUnarchivedChannels(api *slack.Client) (channels []slack.Channel, err error) {
	params := &slack.GetConversationsParameters{
		Limit:           1000,
		ExcludeArchived: "true",
	}
	channels = make([]slack.Channel, 0)
	for i := 0; i < 30; i++ {
		result, cursor, err := api.GetConversations(params)
		if err != nil {
			return nil, err
		}
		params.Cursor = cursor
		channels = append(channels, result...)
		if cursor == "" {
			break
		}
	}
	return channels, nil
}

func GetChannelHistory(api *slack.Client, id string) (messages []slack.Message, err error) {
	params := &slack.GetConversationHistoryParameters{
		ChannelID: id,
		Limit:     1000,
	}
	result, err := api.GetConversationHistory(params)
	if err != nil {
		return nil, err
	}
	return result.Messages, nil
}

func FilterMessage(message []slack.Message, fn func(slack.Message) bool) (result []slack.Message) {
	count := 0
	for _, m := range message {
		if fn(m) {
			count++
		}
	}
	result = make([]slack.Message, count)
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

const timeFormat = "2006/01/02"

type MessageCount struct {
	Key   string
	Count int
}

func CountMessage(message []slack.Message) ([]MessageCount, error) {
	filtered := FilterMessage(message, IgnoreMessageSubType)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no messages")
	}
	oldestTime, err := SlackTimestampToTime(filtered[len(filtered)-1].Timestamp)
	if err != nil {
		return nil, err
	}
	oldestDate := time.Date(oldestTime.Year(), oldestTime.Month(), oldestTime.Day(), 0, 0, 0, 0, oldestTime.Location())
	dateNum := int(math.Ceil(time.Since(oldestDate).Hours() / 24))
	result := make([]MessageCount, dateNum)
	for i := 0; i < dateNum; i++ {
		d := oldestDate.AddDate(0, 0, i).Format(timeFormat)
		result[i].Key = d
	}
	for _, m := range filtered {
		// ignore error
		date, _ := SlackTimestampToTime(m.Timestamp)
		diff := date.Sub(oldestDate)
		i := int(math.Floor(diff.Hours() / 24))
		result[i].Count++
	}
	return result, nil
}

func SlackTimestampToTime(st string) (time.Time, error) {
	spl := strings.Split(st, ".")
	timestamp, err := strconv.ParseInt(spl[0], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	timestampMillisec, err := strconv.ParseInt(spl[1], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	times := time.Unix(timestamp, timestampMillisec*1000)
	return times, nil
}
