package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"time"

	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/slack-go/slack"
)

const (
	maxDate     = 60
	imageWidth  = 400
	imageHeight = 80
	tmpDir      = "./tmp"

	defaultAlertThreshold = 30
)

func parseAlertThreshold() int {
	alertThreshold, err := strconv.Atoi(os.Getenv("ALERT_THREASHOLD"))
	if err != nil {
		return defaultAlertThreshold
	}

	return alertThreshold
}

type config struct {
	alertThreshold       int
	alertChannelID       string
	uploadImageChannelID string
	api                  *slack.Client
}

func createConfig() config {
	alertThreshold := parseAlertThreshold()
	alertChannelID := os.Getenv("SLACK_ALERT_CHANNEL")
	uploadImageChannelID := os.Getenv("SLACK_UPLOAD_IMAGE_CHANNEL")

	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)

	return config{
		alertThreshold:       alertThreshold,
		alertChannelID:       alertChannelID,
		uploadImageChannelID: uploadImageChannelID,
		api:                  api,
	}
}

func postBaseMessage(api slackactivity.SlackPostClient, channelID string, alertThreshold int) (string, error) {
	_, timestamp, err := api.PostMessage(channelID,
		slack.MsgOptionText(fmt.Sprintf("%d日以上メッセージのないチャネルのアラート", alertThreshold), false),
	)
	if err != nil {
		return "", fmt.Errorf("postBaseMessage failed: %w", err)
	}

	return timestamp, nil
}

func postAlertMessage(
	api slackactivity.SlackPostClient,
	channelID string,
	timestamp string,
	channel slack.Channel,
	imageURL string,
	lastMessageTime string,
) error {
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

	return fmt.Errorf("PostMessage failed: id=%s, %w", channel.ID, err)
}

func isSendAlert(count []slackactivity.MessageCount, alertThreshold int) bool {
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

func getMessageCount(
	api slackactivity.SlackChannelHistoryClient,
	channelID string,
	maxDate int,
) ([]slackactivity.MessageCount, error) {
	messages, err := slackactivity.GetChannelHistory(api, channelID)
	if err != nil {
		return nil, fmt.Errorf("GetChannelHistory failed: %w", err)
	}

	now := time.Now()

	count, err := slackactivity.CountMessage(messages, now)
	if err != nil {
		return nil, fmt.Errorf("CountMessage failed: %w", err)
	}

	start := int(math.Max(float64(len(count)-maxDate), 0))

	return count[start:], nil
}

func getLastMessageTime(count []slackactivity.MessageCount) string {
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

//nolint:cyclop,funlen
func _main() int {
	if _, err := os.Stat(tmpDir); err != nil {
		err := os.MkdirAll(tmpDir, os.ModePerm)
		log.Println("mkdir", tmpDir, err)

		if err != nil {
			return 1
		}
	}

	cfg := createConfig()

	ts, err := postBaseMessage(cfg.api, cfg.alertChannelID, cfg.alertThreshold)
	if err != nil {
		log.Println(err)

		return 1
	}

	channels, err := slackactivity.GetAllUnarchivedChannels(cfg.api)
	if err != nil {
		log.Println(err)

		return 1
	}

	log.Println("targetChannels", len(channels))

	for _, c := range channels {
		if !c.IsChannel || c.IsArchived {
			continue
		}

		log.Println(c.Name)

		messageCount, err := getMessageCount(cfg.api, c.ID, maxDate)
		if err != nil {
			log.Println(err)

			continue
		}

		if !isSendAlert(messageCount, cfg.alertThreshold) {
			continue
		}

		outputPath := path.Join(tmpDir, fmt.Sprintf("%s.png", c.Name))
		if err := slackactivity.GeneratePlot(messageCount, c, imageHeight, imageWidth, outputPath); err != nil {
			log.Println(err)

			continue
		}

		params := slack.FileUploadParameters{
			File:     outputPath,
			Channels: []string{cfg.uploadImageChannelID},
		}

		resp, err := cfg.api.UploadFile(params)
		if err != nil {
			log.Println(err)

			continue
		}

		permalink := resp.Permalink
		lastMessageTime := getLastMessageTime(messageCount)

		err = postAlertMessage(
			cfg.api,
			cfg.alertChannelID,
			ts,
			c,
			permalink,
			lastMessageTime,
		)
		if err != nil {
			log.Println(err)
		}
	}

	return 0
}

func main() {
	os.Exit(_main())
}
