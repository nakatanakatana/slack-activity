package main

import (
	"github.com/nlopes/slack"
	"strings"
	"time"
	"strconv"
	"fmt"
)

func errFuncMsg(funcName string) string {
	return fmt.Sprintf("%s: %s\n%s", funcName, "%s", "%s")
}

func MakeChannels(ch []slack.Channel) Channels {
	channels := map[ChannelName]slack.Channel{}
	for _, c := range ch {
		channels[ChannelName(c.Name)] = c
	}
	return channels
}

type Channels map[ChannelName]slack.Channel

func (c Channels) getIDbyName(channelName ChannelName) (string, error) {
	errMsg := errFuncMsg("getIDbyName")
	if ch, ok := c[channelName]; ok {
		return ch.ID, nil
	} else {
		return "", fmt.Errorf(errMsg, fmt.Sprintf("%s NotFound", channelName), "")
	}
}

func (c Channels) keys() ([]ChannelName) {
	list := []ChannelName{}
	for ch := range c {
		list = append(list, ch)
	}
	return list
}

func MakeUsers(ch []slack.User) Users {
	users := map[string]slack.User{}
	for _, u := range ch {
		users[u.ID] = u
	}
	return users
}

type Users map[string]slack.User

func (usr Users) getNamebyID(userID string) (string, error) {
	errMsg := errFuncMsg("getIDbyName")
	u, ok := usr[userID]
	if ok {
		return u.Name, nil
	} else {
		return "", fmt.Errorf(errMsg, fmt.Sprintf("%s NotFound", userID), "")
	}
}

func (usr Users) keys() []string {
	list := []string{}
	for usr := range usr {
		list = append(list, usr)
	}
	return list
}

func SlackTimestampstrToTime(st string) (time.Time, error) {
	errMsg := errFuncMsg("SlackTimestampstrToTime")
	spl := strings.Split(st, ".")
	timestamp, err := strconv.ParseInt(spl[0], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf(errMsg, "timestampSec Parse Failed", err)
	}
	timestampMillisec, err := strconv.ParseInt(spl[1], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf(errMsg, "timestampMilliSec Parse Failed", err)
	}
	times := time.Unix(timestamp, timestampMillisec*1000)
	return times, nil
}
