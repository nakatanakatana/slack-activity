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

const (
	timestampBase   = 10
	timestampIntBit = 64
	milliToNano     = 1000
)

func SlackTimestampToTime(st string) (time.Time, error) {
	spl := strings.Split(st, ".")

	timestamp, err := strconv.ParseInt(spl[0], timestampBase, timestampIntBit)
	if err != nil {
		return time.Time{}, fmt.Errorf("timestamp sec parse failed: %w", err)
	}

	timestampMillisec, err := strconv.ParseInt(spl[1], timestampBase, timestampIntBit)
	if err != nil {
		return time.Time{}, fmt.Errorf("timestamp millisec parse failed: %w", err)
	}

	times := time.Unix(timestamp, timestampMillisec*milliToNano)

	return times, nil
}
