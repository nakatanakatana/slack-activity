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
	BotSettings     BotSettings
	Token           string
	TaskConcurrency int
}

func (f *TaskFactory) init(Token string) (error) {
	errMsg := errFuncMsg("init")
	f.Token = Token
	f.TaskConcurrency = taskConcurrency
	cli := slack.New(f.Token)
	var err error
	auth, err := cli.AuthTest()
	if err != nil {
		return fmt.Errorf(errMsg, "authError", err)
	}
	botSettings := BotSettings{}
	botSettings.Name = auth.User
	botSettings.Id = auth.UserID

	u, err := cli.GetUserInfo(auth.UserID)
	if err != nil {
		return fmt.Errorf(errMsg, "getBotInfoFailed", err)
	}
	botSettings.IconURL = u.Profile.ImageOriginal

	f.BotSettings = botSettings
	err = f.setChannels(cli)
	if err != nil {
		 return fmt.Errorf(errMsg, "setUnArchiveChannelsFailed", err)
	}
	err = f.setAllChannels(cli)
	if err != nil {
		return fmt.Errorf(errMsg, "setAllChannelsFailed", err)
	}
	err = f.setUsers(cli)
	if err != nil {
		return fmt.Errorf(errMsg, "setUsersFailed", err)
	}

	return nil
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

func (f *TaskFactory) NewTask(target ChannelID, imageTargetChannel ChannelName, channelNames, excludeChannelName []ChannelName, archived, useThread bool, threshold int64) *Task {
	cli := slack.New(f.Token)
	timeNow := time.Now()
	return &Task{
		cli,
		f.BotSettings,
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
