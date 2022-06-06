package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/avast/retry-go"
	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/nakatanakatana/slack-activity/report"
	"github.com/slack-go/slack"
)

const (
	reportParallelLimit = 10
	chanLogInterval     = 10 * time.Second
	maxRetry            = 5
	retryDelay          = 1 * time.Second
	rateLimit           = 15
	rateLimitWindowTime = 10 * time.Second
)

//nolint:funlen
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

	rate := make(chan struct{}, rateLimit)
	slots := make(chan struct{}, reportParallelLimit)

	chanLog(rate, slots, chanLogInterval)

	var wg sync.WaitGroup

	wg.Add(len(channels))

	for _, channel := range channels {
		c := channel
		log.Println("target channel:", c.Name)

		go func() {
			defer wg.Done()

			err := retry.Do(func() error {
				rate <- struct{}{}

				go func() {
					time.Sleep(rateLimitWindowTime)
					<-rate
				}()

				slots <- struct{}{}
				defer func() { <-slots }()

				log.Println("start:", c.Name)

				err := report.SendChannelReport(c, cfg, ts, api, api)
				if err != nil {
					return fmt.Errorf("SendChannelReport fail: %s, %w", c.Name, err)
				}

				log.Println("done:", c.Name)

				return nil
			},
				retry.Attempts(maxRetry),
				retry.OnRetry(func(u uint, err error) {
					log.Printf("retry %s: attempt=%d, err=%s", c.Name, u, err)
				}),
				retry.RetryIf(retryIf),
				retry.Delay(retryDelay),
			)
			if err != nil {
				log.Println("retry.Do error:", c.Name, err)
			}
		}()
	}

	wg.Wait()

	return 0
}

func retryIf(err error) bool {
	var rateLimitedErr *slack.RateLimitedError
	isRateLimitErr := errors.As(err, &rateLimitedErr)

	return isRateLimitErr
}

func chanLog(r, s chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				log.Printf("rate:%d, slot:%d\n", len(r), len(s))
			}
		}
	}()
}

func main() {
	os.Exit(_main())
}
