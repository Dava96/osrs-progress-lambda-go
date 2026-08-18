package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type Client struct {
	webhookUrl *url.URL
	httpClient *http.Client
}

type WebhookBody struct {
	Content string  `json:"content"`
	Embeds  []Embed `json:"embeds"`
}

func NewClient(webhookUrl string) (*Client, error) {
	if webhookUrl == "" {
		return nil, errors.New("webhookUrl is empty")
	}

	parsedWebhookUrl, err := url.Parse(webhookUrl)

	if err != nil {
		return nil, err
	}

	return &Client{
		webhookUrl: parsedWebhookUrl,
		httpClient: http.DefaultClient,
	}, nil
}

func (c *Client) Send(message Embed) (int, error) {
	body := WebhookBody{
		Embeds: []Embed{message},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	response, err := c.httpClient.Post(c.webhookUrl.String(), "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		return 0, errors.New(response.Status)
	}

	return response.StatusCode, nil
}
