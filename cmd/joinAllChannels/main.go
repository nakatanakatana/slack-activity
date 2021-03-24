package main

import (
	"fmt"
	"os"

	"github.com/nakatanakatana/slack-activity"
)

func _main() (code int) {
	api := slackActivity.SlackAPI
	channels, err := slackActivity.GetAllUnarchivedChannels(api)
	if err != nil {
		return 1
	}
	fmt.Println("allChannelsLen", len(channels))
	for _, c := range channels {
		if c.IsChannel && !c.IsArchived && !c.IsMember {
			_, _, _, err := api.JoinConversation(c.ID)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return 0
}

func main() {
	os.Exit(_main())
}
