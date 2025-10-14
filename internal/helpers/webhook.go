package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteToWebHook(message string, webhookURL string) error {
	payload := map[string]string{
		"content": fmt.Sprint(message),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	_, err = http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	return err
}
