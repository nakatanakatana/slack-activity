package main_test

import (
	"log"
	"testing"

	main "github.com/nakatanakatana/slack-activity/cmd/eventServer"
)

func TestGetChannelIDs(t *testing.T) {
	t.Parallel()

	type testcase struct {
		name   string
		input  string
		expect []string
	}

	for _, tt := range []testcase{
		{"empty string", "", []string{}},
		{"not contains channel id", "dummy string", []string{}},
		{"invalid channelID string", "dummy <#hhhhhh", []string{}},
		{"invalid channelID string2", "dummy <hogefuga>", []string{}},
		{"contain channelID", "dummy <#CABCDEF>", []string{"CABCDEF"}},
		{"contain multiple channelID ", "dummy <#CABCDEF> hoge <#C123456>", []string{"CABCDEF", "C123456"}},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := main.GetChannelIDs(tt.input)
			log.Println("result", result)
			if len(tt.expect) != len(result) {
				t.Fail()
			}

			for i := 0; i < len(tt.expect); i++ {
				if tt.expect[i] != result[i] {
					t.Fail()
				}
			}
		})
	}
}
