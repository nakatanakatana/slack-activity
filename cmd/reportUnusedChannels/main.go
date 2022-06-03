package main

import (
	"log"
	"os"
	"sync"

	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/nakatanakatana/slack-activity/report"
)

const (
	reportParallelLimit = 10
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

	slots := make(chan struct{}, reportParallelLimit)

	var wg sync.WaitGroup

	wg.Add(len(channels))

	for _, channel := range channels {
		c := channel
		log.Println("target channel:", c.Name)

		slots <- struct{}{}

		go func() {
			err := report.SendChannelReport(c, cfg, ts, api, api)
			if err != nil {
				log.Println(err)
			}

			<-slots
			wg.Done()
		}()
	}

	wg.Wait()

	return 0
}

func main() {
	os.Exit(_main())
}
