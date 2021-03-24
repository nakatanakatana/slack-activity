package main

import (
	"fmt"
	"math"
	"os"
	"path"

	"github.com/nakatanakatana/slack-activity"
	"github.com/slack-go/slack"
)

const (
	maxDate        = 60
	alertThreshold = 14
	imageWidth     = 400
	imageHeight    = 80
)

var alertChannelID string
var uploadImageChannelID string

func init() {
	alertChannelID = os.Getenv("SLACK_ALERT_CHANNEL")
	uploadImageChannelID = os.Getenv("SLACK_UPLOAD_IMAGE_CHANNEL")
}

func getCount(api *slack.Client, channelID string) ([]slackActivity.MessageCount, error) {
	messages, err := slackActivity.GetChannelHistory(api, channelID)
	if err != nil {
		return nil, err
	}
	count, err := slackActivity.CountMessage(messages)
	if err != nil {
		return nil, err
	}
	start := int(math.Max(float64(len(count)-maxDate), 0))
	return count[start:], nil
}

func postBaseMessage(api *slack.Client, channelID string) (timestamp string, err error) {
	_, timestamp, err = api.PostMessage(channelID,
		slack.MsgOptionText(fmt.Sprintf("%d日以上メッセージのないチャネルのアラート", alertThreshold), false),
	)
	return timestamp, err
}

func postAlertMessage(api *slack.Client, channelID string, timestamp string, channel slack.Channel, imageURL string, lastMessageTime string) error {
	_, _, err := api.PostMessage(channelID,
		slack.MsgOptionTS(timestamp),
		slack.MsgOptionAttachments(slack.Attachment{
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
		}),
	)
	return err
}

func isSendAlert(count []slackActivity.MessageCount) bool {
	if len(count) >= alertThreshold {
		for i := 0; i < alertThreshold; i++ {
			if count[len(count)-1-i].Count != 0 {
				return false
			}
		}
		return true
	}
	return false
}

func getLastMessageTime(count []slackActivity.MessageCount) string {
	lastMessageTime := fmt.Sprintf("%d日 以上前", maxDate)
	for i := 0; i < len(count); i++ {
		cur := len(count) - 1 - i
		if count[cur].Count != 0 {
			lastMessageTime = count[cur].Key
			break
		}
	}
	return lastMessageTime
}

func postFile(api *slack.Client, channelID string, filePath string) (permalink string, err error) {
	return slackActivity.PostFile(api, channelID, filePath)
}

func _main() (code int) {
	api := slackActivity.SlackAPI
	ts, err := postBaseMessage(api, alertChannelID)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	channels, err := slackActivity.GetAllUnarchivedChannels(api)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println("targetChannels", len(channels))
	for _, c := range channels {
		if c.IsChannel && !c.IsArchived {
			fmt.Println(c.Name)
			result, err := getCount(api, c.ID)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(result)
			if isSendAlert(result) {
				outputPath := path.Join("./tmp", fmt.Sprintf("%s.png", c.Name))
				if err := slackActivity.GeneratePlot(result, c, imageHeight, imageWidth, outputPath); err != nil {
					fmt.Println(err)
					continue
				}
				permalink, err := postFile(api, uploadImageChannelID, outputPath)
				if err != nil {
					fmt.Println(err)
					continue
				}
				lastMessageTime := getLastMessageTime(result)
				postAlertMessage(api, alertChannelID, ts, c, permalink, lastMessageTime)
			}
		}
	}
	return 0
}

func main() {
	os.Exit(_main())
}
