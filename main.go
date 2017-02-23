package main

import (
	"fmt"
	"os"

	"github.com/nlopes/slack"
)

func _main() int {
	errMsg := errFuncMsg("_main")

	factory := &TaskFactory{}
	factory.Token = SLACK_TOKEN
	factory.TaskConcurrency = 10
	cli := slack.New(factory.Token)
	var err error
	err = factory.setChannels(cli)
	if err != nil {
		fmt.Printf(errMsg, "setUnArchiveChannelsFailed", err)
		return 1
	}
	err = factory.setAllChannels(cli)
	if err != nil {
		fmt.Printf(errMsg, "setAllChannelsFailed", err)
		return 1
	}
	err = factory.setUsers(cli)
	if err != nil {
		fmt.Printf(errMsg, "setUsersFailed", err)
		return 1
	}

	task := factory.NewTask(
		"general",
		"_bot_post_image",
		[]ChannelName{":all", },
		[]ChannelName{"random", "general", },
		false,
		true,
		14,
	)
	task.run()

	//rtm := cli.NewRTM()
	//logger := log.New(os.Stdout, "slack-bot: ", log.LstdFlags)
	//slack.SetLogger(logger)
	//go rtm.ManageConnection()
	//
	//for msg := range rtm.IncomingEvents {
	//	switch ev := msg.Data.(type) {
	//
	//	case *slack.TeamJoinEvent:
	//		err := factory.setUsers(cli)
	//		if err != nil {
	//			fmt.Printf("userDataUpdateFailed, %s", err)
	//		}
	//
	//	case *slack.ChannelCreatedEvent:
	//		err := factory.setAllChannels(cli)
	//		if err != nil {
	//			fmt.Printf("allChannelDataUpdateFailed, %s", err)
	//		}
	//		err = factory.setChannels(cli)
	//		if err != nil {
	//			fmt.Printf("channelDataUpdateFailed, %s", err)
	//		}
	//
	//	case *slack.ChannelArchiveEvent:
	//		err := factory.setChannels(cli)
	//		if err != nil {
	//			fmt.Printf("channelDataUpdateFailed, %s", err)
	//		}
	//
	//	case *slack.RTMError:
	//		fmt.Printf("Error: %s\n", ev.Error())
	//
	//	case *slack.InvalidAuthEvent:
	//		fmt.Printf("Invalid credentials")
	//		return 1
	//
	//	}
	//}
	return 0
}

func main() {
	os.Exit(_main())
}
