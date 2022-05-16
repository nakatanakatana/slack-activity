package main

import (
	"log"
	"os"

	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/slack-go/slack"
)

func _main() int {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)

	channels, err := slackactivity.GetAllUnarchivedChannels(api)
	if err != nil {
		return 1
	}

	log.Println("allChannelsLen", len(channels))

	for _, c := range channels {
		if c.IsChannel && !c.IsArchived && !c.IsMember {
			_, _, _, err := api.JoinConversation(c.ID)
			if err != nil {
				log.Println(err)
			}
		}
	}

	return 0
}

func main() {
	os.Exit(_main())
}
