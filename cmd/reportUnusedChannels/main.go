package main

import (
	"log"
	"os"

	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/nakatanakatana/slack-activity/report"
)

func _main() int {
	cfg := report.CreateConfig()
	err := report.CreateTmpDir(cfg)
	if err != nil {
		log.Println(err)

		return 1
	}

	api := report.CreateSlackClient(cfg)

	ts, err := report.PostBaseMessage(api, cfg)
	if err != nil {
		log.Println(err)

		return 1
	}

	channels, err := slackactivity.GetAllUnarchivedChannels(api)
	if err != nil {
		log.Println(err)

		return 1
	}

	log.Println("targetChannels", len(channels))

	for _, channel := range channels {
		c := channel
		log.Println("target channel:", c.Name)

		err := report.SendChannelReport(c, cfg, ts, api, api)
		if err != nil {
			log.Println(err)
		}
	}

	return 0
}

func main() {
	os.Exit(_main())
}
