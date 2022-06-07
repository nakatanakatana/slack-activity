package report

import (
	"fmt"
	"math"
	"os"
	"path"
	"time"

	"github.com/kelseyhightower/envconfig"
	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/slack-go/slack"
)

const (
	imageWidth  = 400
	imageHeight = 80
)

type Config struct {
	AlertThreshold          int    `split_words:"true" default:"30"`
	SlackAlertChannel       string `split_words:"true"`
	SlackUploadImageChannel string `split_words:"true"`
	MaxDate                 int    `split_words:"true" default:"60"`
	TmpDir                  string `split_words:"true" default:"./tmp"`
}

func CreateConfig() (*Config, error) {
	var c Config

	err := envconfig.Process("", &c)
	if err != nil {
		return nil, fmt.Errorf("envconfig.Process fail: %w", err)
	}

	return &c, nil
}

func CreateTmpDir(cfg *Config) error {
	if _, err := os.Stat(cfg.TmpDir); err != nil {
		err := os.MkdirAll(cfg.TmpDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("mkdir error:%w", err)
		}
	}

	return nil
}

func CreateSlackClient(cfg *Config) *slack.Client {
	token := os.Getenv("SLACK_TOKEN")

	return slack.New(token)
}

func PostBaseMessage(api slackactivity.SlackPostClient, cfg *Config) (string, error) {
	_, timestamp, err := api.PostMessage(cfg.SlackAlertChannel,
		slack.MsgOptionText(
			fmt.Sprintf("%d日以上メッセージのないチャネルのアラート", cfg.AlertThreshold),
			false,
		),
	)
	if err != nil {
		return "", fmt.Errorf("postBaseMessage failed: %w", err)
	}

	return timestamp, nil
}

func PostReportMessage(
	api slackactivity.SlackPostClient,
	reportChannelID string,
	timestamp string,
	targetChannel slack.Channel,
	imageURL string,
	lastMessageTime string,
) error {
	attachment := slackactivity.CreatePostReportMessageAttachements(
		targetChannel,
		imageURL,
		lastMessageTime,
	)

	_, _, err := api.PostMessage(reportChannelID,
		slack.MsgOptionTS(timestamp),
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return fmt.Errorf("PostMessage failed: id=%s, %w", targetChannel.ID, err)
	}

	return nil
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

func GetMessageCount(
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

func GetLastMessageTime(count []slackactivity.MessageCount, maxDate int) string {
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

func SendChannelReport(
	channel slack.Channel,
	cfg *Config,
	reportChannelID string,
	ts string,
	historyClient slackactivity.SlackChannelHistoryClient,
	postClient slackactivity.SlackPostClient,
) error {
	if !channel.IsChannel || channel.IsArchived {
		return nil
	}

	messageCount, err := GetMessageCount(historyClient, channel.ID, cfg.MaxDate)
	if err != nil {
		return err
	}

	if !isSendAlert(messageCount, cfg.AlertThreshold) {
		return nil
	}

	outputPath := path.Join(cfg.TmpDir, fmt.Sprintf("%s.png", channel.Name))
	if err := slackactivity.GeneratePlot(messageCount, channel, imageHeight, imageWidth, outputPath); err != nil {
		return fmt.Errorf("GeneratePlot error: %w", err)
	}

	params := slack.FileUploadParameters{
		File:     outputPath,
		Channels: []string{cfg.SlackUploadImageChannel},
	}

	resp, err := postClient.UploadFile(params)
	if err != nil {
		return fmt.Errorf("UploadFile error: %w", err)
	}

	permalink := resp.Permalink
	lastMessageTime := GetLastMessageTime(messageCount, cfg.MaxDate)

	err = PostReportMessage(
		postClient,
		reportChannelID,
		ts,
		channel,
		permalink,
		lastMessageTime,
	)
	if err != nil {
		return fmt.Errorf("PostReportMessage error: %w", err)
	}

	return nil
}
