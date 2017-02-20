package main

import (
	"github.com/nlopes/slack"
	"time"
)

type TaskFactory struct {
	AllChannels     Channels
	Channels        Channels
	Users           Users
	Token           string
	TaskConcurrency int
}

func (f *TaskFactory) setAllChannels(cli *slack.Client) error {
	allChannels, err := cli.GetChannels(false)
	if err != nil {
		return err
	}
	f.AllChannels = MakeChannels(allChannels)
	return nil
}

func (f *TaskFactory) setChannels(cli *slack.Client) error {
	Channels, err := cli.GetChannels(true)
	if err != nil {
		return err
	}
	f.Channels = MakeChannels(Channels)
	return nil
}

func (f *TaskFactory) setUsers(cli *slack.Client) error {
	allUsers, err := cli.GetUsers()
	if err != nil {
		return err
	}
	f.Users = MakeUsers(allUsers)
	return nil
}

func (f *TaskFactory) NewTask(target, imageTargetChannel string, channelName, excludeChannelName []string, archived, useThread bool) *Task {
	cli := slack.New(f.Token)
	timeNow := time.Now()
	return &Task{
		cli,
		f.TaskConcurrency,
		timeNow,
		f.Channels,
		f.Users,
		target,
		channelName,
		excludeChannelName,
		archived,
		map[string]slack.History{},
		map[string]Summary{},
		map[string]string{},
		map[string]slack.File{},
		imageTargetChannel,
		useThread,
		"",

	}
}
