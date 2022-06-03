package report

import (
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"time"

	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/slack-go/slack"
)

const (
	imageWidth  = 400
	imageHeight = 80

	defaultTmpDir         = "./tmp"
	defaultMaxDate        = 60
	defaultAlertThreshold = 30
)

type Config struct {
	alertThreshold       int
	alertChannelID       string
	uploadImageChannelID string
	maxDate              int
	tmpDir               string
}

func parseTmpDir() string {
	tmpDir := os.Getenv("TMP_DIR")
	if tmpDir == "" {
		return defaultTmpDir
	}

	return tmpDir
}

func parseAlertThreshold() int {
	alertThreshold, err := strconv.Atoi(os.Getenv("ALERT_THREASHOLD"))
	if err != nil {
		return defaultAlertThreshold
	}

	return alertThreshold
}

func parseMaxDate() int {
	maxDate, err := strconv.Atoi(os.Getenv("MAX_DATE"))
	if err != nil {
		return defaultMaxDate
	}

	return maxDate
}

func CreateConfig() Config {
	alertThreshold := parseAlertThreshold()
	maxDate := parseMaxDate()
	alertChannelID := os.Getenv("SLACK_ALERT_CHANNEL")
	uploadImageChannelID := os.Getenv("SLACK_UPLOAD_IMAGE_CHANNEL")
	tmpDir := parseTmpDir()

	return Config{
		alertThreshold:       alertThreshold,
		alertChannelID:       alertChannelID,
		uploadImageChannelID: uploadImageChannelID,
		maxDate:              maxDate,
		tmpDir:               tmpDir,
	}
}

func CreateTmpDir(cfg Config) error {
	if _, err := os.Stat(cfg.tmpDir); err != nil {
		err := os.MkdirAll(cfg.tmpDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("mkdir error:%w", err)
		}
	}

	return nil
}

func CreateSlackClient(cfg Config) *slack.Client {
	token := os.Getenv("SLACK_TOKEN")

	return slack.New(token)
}

func PostBaseMessage(api slackactivity.SlackPostClient, cfg Config) (string, error) {
	_, timestamp, err := api.PostMessage(cfg.alertChannelID,
		slack.MsgOptionText(
			fmt.Sprintf("%d日以上メッセージのないチャネルのアラート", cfg.alertThreshold),
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
	channelID string,
	timestamp string,
	channel slack.Channel,
	imageURL string,
	lastMessageTime string,
) error {
	attachment := slackactivity.CreatePostReportMessageAttachements(channel, imageURL, lastMessageTime)

	_, _, err := api.PostMessage(channelID,
		slack.MsgOptionTS(timestamp),
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return fmt.Errorf("PostMessage failed: id=%s, %w", channel.ID, err)
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
	cfg Config,
	ts string,
	historyClient slackactivity.SlackChannelHistoryClient,
	postClient slackactivity.SlackPostClient,
) error {
	if !channel.IsChannel || channel.IsArchived {
		return nil
	}

	messageCount, err := GetMessageCount(historyClient, channel.ID, cfg.maxDate)
	if err != nil {
		return err
	}

	if !isSendAlert(messageCount, cfg.alertThreshold) {
		return nil
	}

	outputPath := path.Join(cfg.tmpDir, fmt.Sprintf("%s.png", channel.Name))
	if err := slackactivity.GeneratePlot(messageCount, channel, imageHeight, imageWidth, outputPath); err != nil {
		return fmt.Errorf("GeneratePlot error: %w", err)
	}

	params := slack.FileUploadParameters{
		File:     outputPath,
		Channels: []string{cfg.uploadImageChannelID},
	}

	resp, err := postClient.UploadFile(params)
	if err != nil {
		return fmt.Errorf("UploadFile error: %w", err)
	}

	permalink := resp.Permalink
	lastMessageTime := GetLastMessageTime(messageCount, cfg.maxDate)

	err = PostReportMessage(
		postClient,
		cfg.alertChannelID,
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
