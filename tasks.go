package main

import (
	"github.com/nlopes/slack"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/plotutil"
	"path"
	"time"
	"errors"
)

type Task struct {
	cli                    *slack.Client
	Concurrency            int
	TaskStartTime          time.Time
	Channels               Channels
	Users                  Users
	PostTarget             string
	ChannelNames           []string
	ExcludeChannelNames    []string
	ExcludeArchivedChannel bool
	Histories              map[string]slack.History
	Summaries              map[string]Summary
	FileNames              map[string]string
	UploadedFiles          map[string]slack.File
	fileUploadChannel      string
	UseThread              bool
	BaseThreadTimestamp    string
}

func (t *Task) getChannelMessages(channelName string) error {
	chID, err := t.Channels.getIDbyName(channelName)
	if err != nil {
		return fmt.Errorf("Channel:%s is not exists", channelName)
	}
	params := slack.NewHistoryParameters()
	params.Count = 1000
	history, err := t.cli.GetChannelHistory(chID, params)
	if err != nil {
		return fmt.Errorf("getHistoryError, %s", err)
	}
	t.Histories[channelName] = *history
	return nil
}

func (t *Task) makeSummary(channelName string) error {
	latest := time.Time{}
	messages := t.Histories[channelName].Messages
	activity := map[string]float64{}
	oldest := time.Now()
	for _, m := range messages {
		mTime, err := SlackTimestampstrToTime(m.Timestamp)
		if err != nil {
			return err
		}
		activity[mTime.Format("2006-01-02")] += 1
		if oldest.Unix() > mTime.Unix() {
			oldest = mTime
		}
		if latest.Unix() < mTime.Unix() {
			latest = mTime
		}
	}
	maxTimeStr := t.TaskStartTime.Add(24 * time.Hour).Format("2006-01-02")
	maxTime, err := time.Parse("2006-01-02", maxTimeStr)
	if err != nil {
		return err
	}
	for ti := oldest; ti.Unix() < maxTime.Unix(); ti = ti.Add(24 * time.Hour) {
		date := ti.Format("2006-01-02")
		if _, ok := activity[date]; !ok {
			activity[date] = 0
		}
	}
	t.Summaries[channelName] = Summary{
		LatestMessageDate: latest,
		ChannelName:       channelName,
		Activity:          activity,
	}
	return nil
}

func (t *Task) run() error {
	var err error
	targetChannelID, err := t.Channels.getIDbyName(t.PostTarget)
	if err != nil {
		return fmt.Errorf("targetChannelNotFound: %s", t.PostTarget)
	}
	if t.UseThread {
		params := slack.NewPostMessageParameters()
		params.Username = "ChannelActivityCheckBot"
		_, thread_ts, err := t.cli.PostMessage(targetChannelID, "ChannelActivityReport", params)
		if err != nil {
			return errors.New("baseMessagePostFailed")
		}
		t.BaseThreadTimestamp = thread_ts
	}
	// if ChannelNames is nil or :all, target are all
	if t.ChannelNames == nil || t.ChannelNames[0] == ":all" {
		names, err := t.Channels.keys()
		if err != nil {
			return err
		}
		t.ChannelNames = names
	}

	// distinct channelNames
	target := map[string]struct{}{}
	for _, s := range t.ChannelNames {
		target[s] = struct{}{}
	}
	tmpDirName, err := ioutil.TempDir("", t.TaskStartTime.Format("2006-01-02-"))
	if err != nil {
		return fmt.Errorf("mktempDirFailed, %s", err)
	}
	defer os.RemoveAll(tmpDirName)
	for chName := range target {
		err = t.getChannelMessages(chName)
		if err != nil {
			return fmt.Errorf("getMessageError, %s", err)
		}
		err = t.makeSummary(chName)
		if err != nil {
			return fmt.Errorf("makeSummaryError, %s", err)
		}
		s := t.Summaries[chName]
		fileName, err := s.createBarChartImage(imageWidth, imageHeight, tmpDirName)
		if err != nil {
			return fmt.Errorf("createImageError, %s", err)
		}
		t.FileNames[chName] = fileName
		err = t.postImage(chName)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		err = t.postMessage(chName)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
	}

	return nil
}

func (t *Task) postImage(channelName string) error {
	var target string
	if t.fileUploadChannel != "" {
		target = t.fileUploadChannel
	} else {
		target = t.PostTarget
	}
	channelID, err := t.Channels.getIDbyName(target)
	if err != nil {
		return err
	}
	fileName := t.FileNames[channelName]
	params := slack.FileUploadParameters{
		Title:    fileName,
		File:     fileName,
		Channels: []string{channelID, },
	}
	file, err := t.cli.UploadFile(params)
	if err != nil {
		return err
	}
	t.UploadedFiles[channelName] = *file
	return nil
}

func (t *Task) postMessage(channelName string) error {
	userName, err := t.Users.getNamebyID(t.Channels[channelName].Creator)
	if err != nil {
		return err
	}
	params := slack.NewPostMessageParameters()
	if t.UseThread {
		params.ThreadTimestamp = t.BaseThreadTimestamp
	}

	params.Attachments = []slack.Attachment{
		slack.Attachment{
			Text:     fmt.Sprintf("#%s Activity", channelName),
			ImageURL: t.UploadedFiles[channelName].Permalink,
			Fields: []slack.AttachmentField{
				slack.AttachmentField{
					Title: "Creator",
					Value: fmt.Sprintf("@%s", userName),
					Short: true,
				},
				slack.AttachmentField{
					Title: "LastMessageTime",
					Value: t.Summaries[channelName].LatestMessageDate.Format("2006-01-02"),
					Short: true,
				},
			},
		},
	}
	t.cli.PostMessage(t.PostTarget, "", params)

	return nil
}

type Summary struct {
	LatestMessageDate time.Time
	ChannelName       string
	Activity          map[string]float64
}

func (s *Summary) createBarChartImage(width, height int, tmpDirName string) (string, error) {
	// x axis = relative date
	// y axis = activity ( message count )

	keys := make([]string, len(s.Activity))
	i := 0
	for k := range s.Activity {
		keys[i] = k
		i += 1
	}
	sort.Strings(keys)
	values := make([]float64, len(s.Activity))
	for i := 0; i < len(s.Activity); i += 1 {
		values[i] = s.Activity[keys[i]]
	}

	dataValues := plotter.Values(values)
	p, err := plot.New()
	if err != nil {
		return "", err
	}
	p.Title.Text = fmt.Sprintf("#%s Activity [reported at %s]", s.ChannelName, time.Now().Format("2006-01-02"))

	barwidth := float64(width) / float64(len(s.Activity))
	w := vg.Points(barwidth)

	bars, err := plotter.NewBarChart(dataValues, w)
	if err != nil {
		return "", err
	}
	bars.LineStyle.Width = vg.Length(0)
	bars.Color = plotutil.Color(1)
	bars.XMin = float64(-1 * len(s.Activity))

	p.Add(bars)

	outputPath := path.Join(tmpDirName, fmt.Sprintf("%s.png", s.ChannelName))
	if err := p.Save(vg.Length(width), vg.Length(height), outputPath); err != nil {
		return "", err
	}
	return outputPath, nil
}

