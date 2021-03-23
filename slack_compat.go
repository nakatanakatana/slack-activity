package slackActivity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

var token string

func init() {
	token = os.Getenv("SLACK_TOKEN")
}

type PostFileResponse struct {
	Ok   bool        `json:"ok"`
	File PartialFile `json:"file"`
}

type PartialFile struct {
	ID        string `json:"id"`
	Permalink string `json:"permalink"`
}

func PostFile(api *slack.Client, channelID string, filePath string) (permalink string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	values := map[string]io.Reader{
		"file":     f,
		"channels": strings.NewReader(channelID),
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return "", err
		}
	}
	w.Close()
	req, err := http.NewRequest("POST", "https://slack.com/api/files.upload", &b)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var postFileResponse PostFileResponse
	err = json.Unmarshal(body, &postFileResponse)
	if err != nil {
		return "", err
	}
	fmt.Println(postFileResponse)

	return postFileResponse.File.Permalink, nil
}
