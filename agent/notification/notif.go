package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SendNotification sends a message via a Telegram or Bale bot
func SendNotification(url, token, text string, chatID int64) error {
	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", url, token)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API responded with status code %d", resp.StatusCode)
	}

	return nil
}
