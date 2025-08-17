package internal

import (
	"log"
	"os"

	"github.com/pusher/pusher-http-go/v5"
)

var PusherClient *pusher.Client

func NewPusherClient() *pusher.Client {
	pusherAppID := os.Getenv("PUSHER_APP_ID")
	pusherKey := os.Getenv("PUSHER_KEY")
	pusherSecret := os.Getenv("PUSHER_SECRET")
	pusherCluster := os.Getenv("PUSHER_CLUSTER")

	if pusherAppID == "" || pusherKey == "" || pusherSecret == "" || pusherCluster == "" {
		log.Fatalln("failed to load pusher variables:")
	}

	return &pusher.Client{
		AppID:   pusherAppID,
		Key:     pusherKey,
		Secret:  pusherSecret,
		Cluster: pusherCluster,
		Secure:  true,
	}
}
