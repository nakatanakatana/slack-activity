package main

import (
	"github.com/nlopes/slack"
	"strings"
	"errors"
	"time"
	"strconv"
)

func MakeChannels(ch []slack.Channel) Channels {
	channels := map[string]slack.Channel{}
	for _, c := range ch {
		channels[c.Name] = c
	}
	return channels
}

type Channels map[string]slack.Channel

func (c Channels) getIDbyName(channelName string) (string, error) {
	if ch, ok := c[channelName]; ok {
		return ch.ID, nil
	} else {
		return "", errors.New("NotFound")
	}
}

func (c Channels) keys() ([]string, error) {
	list := []string{}
	for ch := range c {
		list = append(list, ch)
	}
	return list, nil
}

func MakeUsers(ch []slack.User) Users {
	users := map[string]slack.User{}
	for _, u := range ch {
		users[u.ID] = u
	}
	return users
}

type Users map[string]slack.User

func (c Users) getNamebyID(userID string) (string, error) {
	u, ok := c[userID]
	if ok {
		return u.Name, nil
	} else {
		return "", errors.New("NotFound")
	}
}

func (u Users) keys() ([]string, error) {
	list := []string{}
	for usr := range u {
		list = append(list, usr)
	}
	return list, nil
}

func SlackTimestampstrToTime(st string) (time.Time, error) {
	spl := strings.Split(st, ".")
	timestamp, err := strconv.ParseInt(spl[0], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	timestampMillisec, err := strconv.ParseInt(spl[1], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	times := time.Unix(timestamp, timestampMillisec*1000)
	return times, nil
}
