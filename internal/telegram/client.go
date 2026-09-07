package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultBaseURL = "https://api.telegram.org"

// Client sends notifications through the Telegram Bot API.
type Client struct {
	Token      string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a Telegram Bot API client.
func NewClient(token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{Token: token, BaseURL: defaultBaseURL, HTTPClient: httpClient}
}

type sendMessageRequest struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

type apiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

// SendMessage sends a plain-text notification to a Telegram chat.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	if c.Token == "" {
		return fmt.Errorf("telegram bot token is required")
	}
	if text == "" {
		return fmt.Errorf("telegram message is required")
	}

	payload, err := json.Marshal(sendMessageRequest{ChatID: chatID, Text: text})
	if err != nil {
		return fmt.Errorf("encode Telegram message: %w", err)
	}

	endpoint := c.BaseURL + "/bot" + url.PathEscape(c.Token) + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create Telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request Telegram API: %w", err)
	}
	defer resp.Body.Close()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode Telegram response: %w", err)
	}
	if resp.StatusCode != http.StatusOK || !result.OK {
		if result.Description == "" {
			result.Description = resp.Status
		}
		return fmt.Errorf("Telegram API failed: %s", result.Description)
	}
	return nil
}

// ChatIDFromString is useful when chat IDs are supplied through environment
// variables or configuration files.
func ChatIDFromString(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid Telegram chat ID %q: %w", value, err)
	}
	return id, nil
}
