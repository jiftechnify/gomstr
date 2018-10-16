package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mattn/go-mastodon"
)

func main() {
	c := mastodon.NewClient(&mastodon.Config{
		Server:       "https://mstdn.jp",
		ClientID:     "CLIENT_ID",
		ClientSecret: "CLIENT_SECRET",
	})
	err := c.Authenticate(context.Background(), "email", "pass")
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
