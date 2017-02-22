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
	"sync"
)

type ChannelName string

type Task struct {
	cli                    *slack.Client
	Concurrency            int
	TaskStartTime          time.Time
	Channels               Channels
	Users                  Users
	PostTarget             ChannelName
	ChannelNames           []ChannelName
	ExcludeChannelNames    []ChannelName
	ExcludeArchivedChannel bool
	Histories              map[ChannelName]slack.History
	Summaries              map[ChannelName]Summary
	FileNames              map[ChannelName]string
	UploadedFiles          map[ChannelName]slack.File
	fileUploadChannel      ChannelName
	UseThread              bool
	BaseThreadTimestamp    string
}

func (t *Task) getChannelMessages(channelName ChannelName) error {
	errMsg := errFuncMsg("getChannelMessage")
	chID, err := t.Channels.getIDbyName(channelName)
	if err != nil {
		return fmt.Errorf(errMsg, fmt.Sprintf("Channel:%s is not exists", channelName), err)
	}
	params := slack.NewHistoryParameters()
	params.Count = 1000
	history, err := t.cli.GetChannelHistory(chID, params)
	if err != nil {
		return fmt.Errorf(errMsg, fmt.Sprintf("get %s HistoryError", channelName), err)
	}
	t.Histories[channelName] = *history
	return nil
}

func (t *Task) makeSummary(channelName ChannelName) error {
	errMsg := errFuncMsg("makeSummary")
	latest := time.Time{}
	messages := t.Histories[channelName].Messages
	activity := map[string]float64{}
	oldest := time.Now()
	for _, m := range messages {
		mTime, err := SlackTimestampstrToTime(m.Timestamp)
		if err != nil {
			return fmt.Errorf(errMsg, "timestamp Parse Error", err)
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
		return fmt.Errorf(errMsg, "create Max date Failed", err)
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
	errMsg := errFuncMsg("run")
	var err error
	targetChannelID, err := t.Channels.getIDbyName(t.PostTarget)
	if err != nil {
		return fmt.Errorf(errMsg, fmt.Sprintf("targetChannelNotFound: %s", t.PostTarget), err)
	}
	if t.UseThread {
		params := slack.NewPostMessageParameters()
		params.Username = "channel-activity"
		params.IconURL = "https://avatars.slack-edge.com/2017-02-21/144176451349_c3cd9d3c4fcda7f4c6a2_72.png"
		_, thread_ts, err := t.cli.PostMessage(targetChannelID, "ChannelActivityReport", params)
		if err != nil {
			return fmt.Errorf(errMsg, "baseMessagePostFailed", err)
		}
		t.BaseThreadTimestamp = thread_ts
	}
	// if ChannelNames is nil or :all, target are all
	if t.ChannelNames == nil || t.ChannelNames[0] == ":all" {
		names := t.Channels.keys()
		t.ChannelNames = names
	}

	// distinct and exclude channelNames
	target := map[ChannelName]struct{}{}
	for _, s := range t.ChannelNames {
		target[s] = struct{}{}
	}
	for _, e := range t.ExcludeChannelNames{
		if _, ok := target[e]; ok{
			delete(target, e)
		}
	}

	tmpDirName, err := ioutil.TempDir("", t.TaskStartTime.Format("2006-01-02-"))
	if err != nil {
		return fmt.Errorf(errMsg, "mktempDirFailed", err)
	}
	defer os.RemoveAll(tmpDirName)
	errChan := make(chan error, len(target))
	semaphore := make(chan int, t.Concurrency)
	defer close(errChan)

	var wg sync.WaitGroup
	wg.Add(len(target))
	for chName := range target {
		go func(chName ChannelName) {
			defer wg.Done()
			semaphore <- 1
			defer func() { <-semaphore }()
			fmt.Printf("%s: start\n", chName)
			end := make(chan int)
			defer close(end)

			err = t.getChannelMessages(chName)
			if err != nil {
				fmt.Printf("%s: %s", chName, err)
				errChan <- fmt.Errorf(errMsg, "getMessageError", err)
				return
			}

			err = t.makeSummary(chName)
			if err != nil {
				fmt.Printf("%s: %s", chName, err)
				errChan <- fmt.Errorf(errMsg, "makeSummaryError", err)
				return
			}

			s := t.Summaries[chName]
			fileName, err := s.createBarChartImage(imageWidth, imageHeight, tmpDirName)
			if err != nil {
				fmt.Printf("%s: %s", chName, err)
				errChan <- fmt.Errorf(errMsg, "createImageError", err)
				return
			}
			t.FileNames[chName] = fileName
			err = t.postImage(chName)
			if err != nil {
				fmt.Printf("%s: %s", chName, err)
				errChan <- fmt.Errorf(errMsg, "postImageError", err)
				return
			}

			err = t.postMessage(chName)
			if err != nil {
				fmt.Printf("%s: %s", chName, err)
				errChan <- fmt.Errorf(errMsg, "postMessageError", err)
				return
			}

			errChan <- nil
		}(chName)
	}
	wg.Wait()

	// TODO: add handle when goroutine error
	return nil
}

func (t *Task) postImage(channelName ChannelName) error {
	var target ChannelName
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

func (t *Task) postMessage(channelName ChannelName) error {
	errMsg := errFuncMsg("postMessage")
	userID := t.Channels[channelName].Creator
	userName, err := t.Users.getNamebyID(userID)
	if err != nil {
		return fmt.Errorf(errMsg, "", err)
	}
	channelID := t.Channels[channelName].ID
	params := slack.NewPostMessageParameters()
	if t.UseThread {
		params.ThreadTimestamp = t.BaseThreadTimestamp
	}

	params.Attachments = []slack.Attachment{
		slack.Attachment{
			Title:    fmt.Sprintf("<#%s|%s>", channelID, channelName),
			ImageURL: t.UploadedFiles[channelName].Permalink,
			Fields: []slack.AttachmentField{
				slack.AttachmentField{
					Title: "Creator",
					Value: fmt.Sprintf("<@%s|%s>", userID, userName),
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
	t.cli.PostMessage(string(t.PostTarget), "", params)

	return nil
}

type Summary struct {
	LatestMessageDate time.Time
	ChannelName       ChannelName
	Activity          map[string]float64
}

func (s *Summary) createBarChartImage(width, height int, tmpDirName string) (string, error) {
	// x axis = relative date
	// y axis = activity ( message count )

	errMsg := errFuncMsg("createBarChartImage")
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
		return "", fmt.Errorf(errMsg, "make Plot Failed", err)
	}
	p.Title.Text = fmt.Sprintf("#%s Activity [reported at %s]", s.ChannelName, time.Now().Format("2006-01-02"))

	barwidth := float64(width) / float64(len(s.Activity))
	w := vg.Points(barwidth)

	bars, err := plotter.NewBarChart(dataValues, w)
	if err != nil {
		return "", fmt.Errorf(errMsg, "bar chart Error", err)
	}
	bars.LineStyle.Width = vg.Length(0)
	bars.Color = plotutil.Color(1)
	bars.XMin = float64(-1 * len(s.Activity))

	p.Add(bars)

	outputPath := path.Join(tmpDirName, fmt.Sprintf("%s.png", s.ChannelName))
	if err := p.Save(vg.Length(width), vg.Length(height), outputPath); err != nil {
		return "", fmt.Errorf(errMsg, "save to file failed", err)
	}
	return outputPath, nil
}
