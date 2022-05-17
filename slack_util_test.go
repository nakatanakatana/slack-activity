package slackactivity_test

import (
	"testing"
	"time"

	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/slack-go/slack"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func createTestMessages() []slack.Message {
	//nolint:exhaustruct
	return []slack.Message{
		// oldest is 2022/05/10
		{Msg: slack.Msg{Timestamp: "1652693435.024849"}},
		{Msg: slack.Msg{Timestamp: "1652607035.024849"}},
		{Msg: slack.Msg{Timestamp: "1652520635.024849"}},
		{Msg: slack.Msg{Timestamp: "1652175035.024849"}},
	}
}

func createMessageCount(key string, count int) slackactivity.MessageCount {
	return slackactivity.MessageCount{
		Key:   key,
		Count: count,
	}
}

func TestCountMessage(t *testing.T) {
	t.Run("check count result", func(t *testing.T) {
		t.Parallel()
		testCase := createTestMessages()
		latest := time.Date(2022, time.May, 17, 1, 2, 3, 4, time.UTC)
		expect := make([]slackactivity.MessageCount, 8)
		expect[0] = createMessageCount("2022/05/10", 1)
		expect[1] = createMessageCount("2022/05/11", 0)
		expect[2] = createMessageCount("2022/05/12", 0)
		expect[3] = createMessageCount("2022/05/13", 0)
		expect[4] = createMessageCount("2022/05/14", 1)
		expect[5] = createMessageCount("2022/05/15", 1)
		expect[6] = createMessageCount("2022/05/16", 1)
		expect[7] = createMessageCount("2022/05/17", 0)

		result, err := slackactivity.CountMessage(testCase, latest)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(result, 8))
		assert.DeepEqual(t, result, expect)
	})
}

func TestSlackTimestampToTime(t *testing.T) {
	t.Run("parse success", func(t *testing.T) {
		t.Parallel()
		testCase := "1652693435.024849"
		expect := time.Date(2022, time.May, 16, 9, 30, 35, 24849000, time.UTC)

		result, err := slackactivity.SlackTimestampToTime(testCase)
		assert.NilError(t, err)
		assert.DeepEqual(t, expect, result)
	})

	t.Run("parse sec failed", func(t *testing.T) {
		t.Parallel()
		testCase := "hoge.024849"
		expectErr := "timestamp sec parse failed"

		_, err := slackactivity.SlackTimestampToTime(testCase)
		assert.ErrorContains(t, err, expectErr)
	})

	t.Run("parse millisec failed", func(t *testing.T) {
		t.Parallel()
		testCase := "1652693435.abc"
		expectErr := "timestamp millisec parse failed"

		_, err := slackactivity.SlackTimestampToTime(testCase)
		assert.ErrorContains(t, err, expectErr)
	})
}
