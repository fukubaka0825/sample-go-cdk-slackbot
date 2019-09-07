package main

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/gommon/log"
	"net/http"
)

type SlackRequestBody struct {
	Channel    string            `json:"channel"`
	Username   string            `json:"username"`
	Text       string            `json:"text"`
	IconEmoji  *string           `json:"icon_emoji,omitempty"`
	Attachment []SlackAttachment `json:"attachments"`
}

type SlackAttachment struct {
	Fallback string        `json:"fallback"`
	Pretext  string        `json:"pretext"`
	Color    *NOTIFY_COLOR `json:"color,omitempty"`
	Fields   []SlackField  `json:"fields"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type NOTIFY_COLOR string

const NOTIFY_COLOR_RED NOTIFY_COLOR = `#F00000`
const NOTIFY_COLOR_GREEN NOTIFY_COLOR = `#00F000`
const NOTIFY_COLOR_WHITE NOTIFY_COLOR = `#FFFFFF`

type SLACK_MESSAGE_LEVEL int

const (
	SLACK_MESSAGE_LEVEL_ALART SLACK_MESSAGE_LEVEL = iota
	SLACK_MESSAGE_LEVEL_OK
	SLACK_MESSAGE_LEVEL_NOTIFY
)

// <@channel>
/*
	lambdaの場合はこんなです。
    const slackMessage = {
        username: `OK: VKのアラート`,
        channel: slackChannel,
        icon_emoji: `:nick:`,
        "attachments": [{
            "fallback": alarmName,
            "pretext": ":parrotdad: *アラート復帰* :parrotdad:",
            "color": "#00F000",
            "fields": [{
                "title": alarmName,
                "value": `アラーム: ${alarmName} \n 理由:${reason}`
            }]
        }]
    }

*/

func ReportToLaboonSlack(webhockURL, channel, name, icon, title, text string, level SLACK_MESSAGE_LEVEL) error {

	if icon == "" {
		icon = func() string {
			switch level {
			case SLACK_MESSAGE_LEVEL_ALART:
				return `:heavy_exclamation_mark:`
			case SLACK_MESSAGE_LEVEL_OK:
				return `:ok_hand::skin-tone-2:`
			}
			return `:vk:`
		}()
	}

	preText := func() string {
		switch level {
		case SLACK_MESSAGE_LEVEL_ALART:
			return `*アラート*`
		case SLACK_MESSAGE_LEVEL_OK:
			return `*OK*`
		}
		return `*お知らせ*`

	}()

	color := func() NOTIFY_COLOR {
		switch level {
		case SLACK_MESSAGE_LEVEL_ALART:
			return NOTIFY_COLOR_RED
		case SLACK_MESSAGE_LEVEL_OK:
			return NOTIFY_COLOR_GREEN
		}
		return NOTIFY_COLOR_WHITE

	}()

	reqBody := SlackRequestBody{
		Channel:   channel,
		Username:  name,
		IconEmoji: &icon,
		//Text:      text,
		Attachment: []SlackAttachment{
			SlackAttachment{
				Fallback: name,
				Pretext:  preText,
				Color:    &color,
				Fields: []SlackField{
					SlackField{
						Title: title,
						Value: text,
					},
				},
			},
		},
	}

	return ReportToSlack(webhockURL, reqBody)
}

func ReportToSlack(webhookUrl string, reqBody SlackRequestBody) error {

	jsonAsByte, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(
		"POST",
		webhookUrl,
		bytes.NewReader(jsonAsByte),
	)

	if err != nil {
		log.Error(err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}

	defer resp.Body.Close()

	return nil
}
