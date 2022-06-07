package slackactivity_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	slackactivity "github.com/nakatanakatana/slack-activity"
	"github.com/slack-go/slack"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

type conversationsMock struct {
	channels []slack.Channel
}

var _ slackactivity.SlackGetChannelsClient = conversationsMock{}

var (
	ErrMock                        = errors.New("mock error")
	ErrGetConversationInfoNotFound = errors.New("GetConversationInfo:not found")
)

func (m conversationsMock) GetConversations(
	params *slack.GetConversationsParameters,
) ([]slack.Channel, string, error) {
	var cur int

	cur, _ = strconv.Atoi(params.Cursor)
	if params.Cursor == "" {
		cur = 0
	}

	if cur > len(m.channels) {
		return []slack.Channel{}, "", fmt.Errorf("cur over max: %w", ErrMock)
	}

	restLength := len(m.channels[cur:])

	switch {
	case restLength >= params.Limit+1:
		return m.channels[cur : cur+params.Limit], strconv.Itoa(cur + params.Limit), nil
	case restLength > 0:
		return m.channels[cur:], "", nil
	default:
		return []slack.Channel{}, "", fmt.Errorf("unknown: %w", ErrMock)
	}
}

func (m conversationsMock) GetConversationInfo(channelID string, includeLocale bool) (*slack.Channel, error) {
	if len(m.channels) == 0 {
		return nil, ErrGetConversationInfoNotFound
	}

	return &m.channels[0], nil
}

func newConversationsMock(length int) conversationsMock {
	c := make([]slack.Channel, length)

	for i := range c {
		c[i] = slack.Channel{
			GroupConversation: slack.GroupConversation{
				Name: strconv.Itoa(i),
			},
			IsChannel: true,
			IsGeneral: false,
			IsMember:  true,
			Locale:    "",
		}
	}

	return conversationsMock{
		channels: c,
	}
}

func TestConversationsMock(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		cur         string
		limit       int
		expectLen   int
		expectStart string
		expectEnd   string
		expectNext  string
		expectErr   string
	}

	for _, tt := range []testCase{
		{"limitの件数結果を返す", "", 5, 5, "0", "4", "5", ""},
		{"limitが全件を超える場合は全件を返す", "", 20, 10, "0", "9", "", ""},
		{"curを指定した場合はそこからの結果を返す", "1", 10, 9, "1", "9", "", ""},
		{"curが範囲外を指定した場合はエラー返す", "11", 0, 0, "", "", "", "cur over max"},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := newConversationsMock(10)

			result, next, err := mock.GetConversations(&slack.GetConversationsParameters{
				Cursor: tt.cur,
				Limit:  tt.limit,
			})
			if tt.expectErr == "" {
				assert.NilError(t, err)
				assert.Assert(t, is.Len(result, tt.expectLen))
				assert.Equal(t, result[0].GroupConversation.Name, tt.expectStart)
				assert.Equal(t, result[len(result)-1].GroupConversation.Name, tt.expectEnd)
				assert.Equal(t, next, tt.expectNext)
			} else {
				assert.ErrorContains(t, err, tt.expectErr)
			}
		})
	}
}

func TestGetAllUnarchivedChannels(t *testing.T) {
	t.Parallel()

	t.Run("1000件以内のときにページングをせずに全件のリストを返す", func(t *testing.T) {
		t.Parallel()
		const channelsLen = 200
		mock := newConversationsMock(channelsLen)
		result, err := slackactivity.GetAllUnarchivedChannels(mock)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(result, channelsLen))
	})

	t.Run("チャネルが5500件のときにページングをして全件のリストを返す", func(t *testing.T) {
		t.Parallel()
		const channelsLen = 5500
		mock := newConversationsMock(channelsLen)
		result, err := slackactivity.GetAllUnarchivedChannels(mock)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(result, channelsLen))
	})

	t.Run("チャネルが1000*30件以上のときにページングをして30000件のリストを返す(途中で中断する)", func(t *testing.T) {
		t.Parallel()
		const channelsLen = 35000
		mock := newConversationsMock(channelsLen)
		result, err := slackactivity.GetAllUnarchivedChannels(mock)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(result, 1000*30))
	})
}
