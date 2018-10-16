package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mattn/go-mastodon"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	c := mastodon.NewClient(&mastodon.Config{
		Server:       "https://mstdn.jp",
		ClientID:     os.Getenv("MSTDN_CLIENT_ID"),
		ClientSecret: os.Getenv("MSTDN_CLIENT_SECRET"),
	})
	err = c.Authenticate(context.Background(), os.Getenv("MSTDN_USER_EMAIL"), os.Getenv("MSTDN_USER_PASSWORD"))
	if err != nil {
		log.Fatal(err)
	}
	eventChan, err := c.StreamingPublic(context.Background(), false)
	if err != nil {
		log.Fatal(err)
	}
	for e := range eventChan {
		fmt.Println(extractContent(e))
	}
}

func extractContent(ev mastodon.Event) string {
	switch e := ev.(type) {
	case *mastodon.UpdateEvent:
		status := e.Status
		return fmt.Sprintf("[%s]: %s", status.Account.Username, status.Content)
	default:
		return "other events"
	}
}
