package main

import (
	"encoding/json"
	"fmt"
	"github.com/slack-go/slack"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type HealthResponse struct {
	StatusMsg string `json:"status"`
}

func main() {
	// Print input arguments
	//fmt.Println(strings.Join(os.Args[1:], " "))

	const targetUrlKey = "TARGET_URL"
	targetUrl := os.Getenv(targetUrlKey)

	healthResponse, responseStatus, err := sendHTTPRequest(targetUrl)
	if responseStatus == 0 && err != nil {
		fmt.Println("sendHTTPRequest function Error:", err)
		os.Exit(1)
	}

	if responseStatus != 200 || healthResponse.StatusMsg != "UP" {
		sendSlackMessage(healthResponse, responseStatus, err)
	}
}

func sendSlackMessage(healthResponse *HealthResponse, responseStatus int, err error) (string, string) {
	slackToken := os.Getenv("SLACK_TOKEN")
	slackChannel := os.Getenv("SERVER_HEALTH_CHECKER_SLACK_CHANNEL")
	slackApi := slack.New(slackToken)

	now := time.Now().Format("2006-01-02 15:04:05")
	attachment := slack.Attachment{
		Pretext: ":scream: 서버가 죽었습니다 :scream:",
		Fields: []slack.AttachmentField{
			{
				Title: "현재 시각",
				Value: now,
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
		return nil, resp.StatusCode, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	var responseJSON HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&responseJSON)
	if err != nil {
		return nil, 0, err
	}

	return &responseJSON, resp.StatusCode, nil
}
