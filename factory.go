package main

import (
	"fmt"
	"time"

	"github.com/nlopes/slack"
)

type TaskFactory struct {
	AllChannels     Channels
	Channels        Channels
	Users           Users
	Token           string
	TaskConcurrency int
}

func (f *TaskFactory) setAllChannels(cli *slack.Client) error {
	errMsg := errFuncMsg("setAllChannels")
	allChannels, err := cli.GetChannels(false)
	if err != nil {
		return fmt.Errorf(errMsg, "Get Channels Failed", err)
	}
	f.AllChannels = MakeChannels(allChannels)
	return nil
}

func (f *TaskFactory) setChannels(cli *slack.Client) error {
	errMsg := errFuncMsg("setChannels")
	Channels, err := cli.GetChannels(true)
	if err != nil {
		return fmt.Errorf(errMsg, "Get Channels Failed", err)
	}
	f.Channels = MakeChannels(Channels)
	return nil
}

func (f *TaskFactory) setUsers(cli *slack.Client) error {
	errMsg := errFuncMsg("setUsers")
	allUsers, err := cli.GetUsers()
	if err != nil {
		return fmt.Errorf(errMsg, "Get Users Failed", err)
	}
	f.Users = MakeUsers(allUsers)
	return nil
}

func (f *TaskFactory) NewTask(target, imageTargetChannel ChannelName, channelNames, excludeChannelName []ChannelName, archived, useThread bool, threshold int64) *Task {
	cli := slack.New(f.Token)
	timeNow := time.Now()
	return &Task{
		cli,
		f.TaskConcurrency,
		timeNow,
		f.Channels,
		f.Users,
		target,
		channelNames,
		excludeChannelName,
		archived,
		map[ChannelName]slack.History{},
		map[ChannelName]Summary{},
		map[ChannelName]string{},
		map[ChannelName]slack.File{},
		imageTargetChannel,
		useThread,
		threshold,
		"",

	}
}
