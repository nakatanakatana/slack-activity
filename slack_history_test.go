package slackactivity_test

import (
	"strconv"
	"testing"

	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/slack-go/slack"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

type conversationHistoryMock struct {
	messages []slack.Message
}

var _ slackactivity.SlackChannelHistoryClient = conversationHistoryMock{}

func (m conversationHistoryMock) GetConversationHistory(
	params *slack.GetConversationHistoryParameters,
) (*slack.GetConversationHistoryResponse, error) {
	return &slack.GetConversationHistoryResponse{
		Messages: m.messages,
	}, nil
}

func newConversationHistoryMock(length int) conversationHistoryMock {
	m := make([]slack.Message, length)

	for i := range m {
		m[i] = slack.Message{Msg: slack.Msg{Text: strconv.Itoa(i)}}
	}

	return conversationHistoryMock{
		messages: m,
	}
}

func TestGetChannelHistory(t *testing.T) {
	t.Parallel()

	const mockLen = 1000
	mock := newConversationHistoryMock(mockLen)

	t.Run("GetConversationsHistoryの結果内のMessagesを返す", func(t *testing.T) {
		result, err := slackactivity.GetChannelHistory(mock, "dummy")
		assert.NilError(t, err)
		assert.Assert(t, is.Len(result, mockLen))
	})
}

func TestFilterMessage(t *testing.T) {
	t.Parallel()

	const mockLen = 1000
	m := newConversationHistoryMock(mockLen)
	messages := m.messages

	t.Run("fnがすべてtrueを返すとき元のメッセージと同じものを返す", func(t *testing.T) {
		t.Parallel()

		fn := func(slack.Message) bool { return true }
		result := slackactivity.FilterMessage(messages, fn)

		assert.Assert(t, is.Len(result, mockLen))
		assert.DeepEqual(t, result, messages)
	})

	t.Run("fnがすべてfalseを返すとき空のメッセージを返す", func(t *testing.T) {
		t.Parallel()

		fn := func(slack.Message) bool { return false }
		result := slackactivity.FilterMessage(messages, fn)

		assert.Assert(t, is.Len(result, 0))
		assert.DeepEqual(t, result, []slack.Message{})
	})
}
