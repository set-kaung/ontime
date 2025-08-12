package internal

import (
	"os"

	"github.com/pusher/pusher-http-go/v5"
)

var PusherClient = NewPusherClient()

func NewPusherClient() *pusher.Client {
	return &pusher.Client{
		AppID:   os.Getenv("PUSHER_APP_ID"),
		Key:     os.Getenv("PUSHER_KEY"),
		Secret:  os.Getenv("PUSHER_SECRET"),
		Cluster: os.Getenv("PUSHER_CLUSTER"), // e.g., "us2"
		Secure:  true,
	}
}
