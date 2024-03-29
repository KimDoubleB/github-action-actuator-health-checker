package main

import (
	"encoding/json"
	"fmt"
	"github.com/slack-go/slack"
	"io"
	"net/http"
	"os"
	"strconv"
)

type HealthResponse struct {
	StatusMsg string `json:"status"`
}

func main() {
	// Print input arguments
	//fmt.Println(strings.Join(os.Args[1:], " "))

	targetUrl := os.Getenv("INPUT_HEALTH_CHECK_URL")

	healthResponse, responseStatus, err := sendHTTPRequest(targetUrl)
	if responseStatus == 0 && err != nil {
		fmt.Println("sendHTTPRequest function Error:", err)
		os.Exit(1)
	}

	if responseStatus != 200 || healthResponse.StatusMsg != "UP" {
		sendSlackMessage(healthResponse, responseStatus, targetUrl, err)
	}
}

func sendSlackMessage(healthResponse *HealthResponse, responseStatus int, targetUrl string, err error) (string, string) {
	slackToken := os.Getenv("INPUT_SLACK_TOKEN")
	slackChannel := os.Getenv("INPUT_SLACK_CHANNEL")

	slackApi := slack.New(slackToken)
	attachment := slack.Attachment{
		Pretext: ":scream: 서버가 죽었습니다 :scream:",
		Fields: []slack.AttachmentField{
			{
				Title: "서버 Health check URL",
				Value: targetUrl,
				Short: false,
			},
			{
				Title: "서버 health response status",
				Value: strconv.Itoa(responseStatus),
				Short: true,
			},
			{
				Title: "서버 health status",
				Value: healthResponse.StatusMsg,
				Short: true,
			},
		},
	}

	channelId, timestamp, err := slackApi.PostMessage(slackChannel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		fmt.Println("Slack send error:", err)
		os.Exit(1)
	}
	return channelId, timestamp
}

func sendHTTPRequest(url string) (*HealthResponse, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		downResponse := &HealthResponse{
			StatusMsg: "DOWN",
		}
		return downResponse, resp.StatusCode, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	var responseJSON HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&responseJSON)
	if err != nil {
		return nil, 0, err
	}

	return &responseJSON, resp.StatusCode, nil
}
