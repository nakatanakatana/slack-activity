package main

import (
	"fmt"
	"os"

	"github.com/nlopes/slack"
	"log"
	"strings"
)

func _main() int {
	errMsg := errFuncMsg("_main")

	SlackToken := os.Getenv("SLACK_TOKEN")

	factory := &TaskFactory{}
	if err := factory.init(SlackToken); err != nil {
		fmt.Printf(errMsg, "factoryInit Failed", err)
		return 1
	}

	//task := factory.NewTask(
	//	"general",
	//	"_bot_post_image",
	//	[]ChannelName{":all", },
	//	[]ChannelName{"random", "general", },
	//	false,
	//	true,
	//	14,
	//)
	//task.run()

	cli := slack.New(SlackToken)
	a, _ := cli.AuthTest()
	fmt.Println(a)
	rtm := cli.NewRTM()
	logger := log.New(os.Stdout, "slack-bot: ", log.LstdFlags)
	slack.SetLogger(logger)
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.TeamJoinEvent:
			err := factory.setUsers(cli)
			if err != nil {
				fmt.Printf("userDataUpdateFailed, %s", err)
			}

		case *slack.ChannelCreatedEvent:
			err := factory.setAllChannels(cli)
			if err != nil {
				fmt.Printf("allChannelDataUpdateFailed, %s", err)
			}
			err = factory.setChannels(cli)
			if err != nil {
				fmt.Printf("channelDataUpdateFailed, %s", err)
			}

		case *slack.ChannelArchiveEvent:
			err := factory.setChannels(cli)
			if err != nil {
				fmt.Printf("channelDataUpdateFailed, %s", err)
			}
		case *slack.MessageEvent:
			toMe := strings.Contains(ev.Msg.Text, factory.BotSettings.Id)
			var checkChannels []ChannelName
			for channelName := range factory.AllChannels {
				if strings.Contains(ev.Msg.Text, factory.AllChannels[channelName].ID) {
					checkChannels = append(checkChannels, channelName)
				}
			}
			if strings.Contains(ev.Msg.Text, ":all") {
				checkChannels = append(checkChannels, ":all")
			}
			if ev.Msg.SubType == "" && toMe && len(checkChannels) > 0 {
				task := factory.NewTask(
					ChannelID(ev.Channel),
					postImageTarget,
					checkChannels,
					[]ChannelName{"", },
					false,
					false,
					0,

				)
				task.run()
			}
			fmt.Println(ev.Msg.SubType, toMe, len(checkChannels))

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return 1

		}
	}
	return 0
}

func main() {
	os.Exit(_main())
}
