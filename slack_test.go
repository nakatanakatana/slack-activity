package slackActivity

import (
	"fmt"
	"os"
	"testing"

	"github.com/slack-go/slack"
)

func TestGetAllChannels(t *testing.T) {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	result, err := GetAllUnarchivedChannels(api)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	for _, r := range result {
		fmt.Println(r.IsChannel, r.Name, r.ID, r.IsArchived, r.IsMember)
	}
	fmt.Println(len(result))
}

func TestGetChannelHistory(t *testing.T) {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	result, err := GetChannelHistory(api, "C48S2UXPC")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	for _, r := range result {
		fmt.Println(r.Timestamp, r.Msg.Text, r.Type, r.SubType)
	}
	fmt.Println(len(result))
}

func TestFilterMessage(t *testing.T) {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	messages, err := GetChannelHistory(api, "C48S2UXPC")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	result := FilterMessage(messages, IgnoreMessageSubType)
	for _, r := range result {
		fmt.Println(r.Timestamp, r.Msg.Text, r.Type, r.SubType)
	}
	fmt.Println(len(result))
}

func TestCountMessage(t *testing.T) {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	messages, err := GetChannelHistory(api, "C48S2UXPC")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	result, err := CountMessage(FilterMessage(messages, IgnoreMessageSubType))
	fmt.Println(result, err, len(result))
}

func TestJoinChannel(t *testing.T) {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	result, s, sa, err := api.JoinConversation("C48S2UXPC")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(result, s, sa)
}

func TestPostMessage(t *testing.T) {
	t.Skip()
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	result, ts, err := api.PostMessage(
		"C48S2UXPC",
		slack.MsgOptionText("test message", false),
	)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(result, ts)
	result2, ts, err := api.PostMessage(
		"C48S2UXPC",
		slack.MsgOptionTS(ts),
		slack.MsgOptionText("inThread", false),
	)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(result2, ts)
}

func TestPostFile(t *testing.T) {
	// t.Skip()
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	result, err := PostFile(api, "C47QX63GS", "./tmp/random.png")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(result)
	result2, ts, err := api.PostMessage("C48S2UXPC",
		slack.MsgOptionAttachments(slack.Attachment{
			Title: "title",
		}),
	)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(result2, ts)
}

func TestGetUsers(t *testing.T) {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	user, _ := api.GetUserInfo("U0LBT5WE6")
	fmt.Println(user.Name)
}
