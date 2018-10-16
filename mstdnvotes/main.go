package main

import (
	"syscall"
	"os/signal"
	"context"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mattn/go-mastodon"
	"github.com/bitly/go-nsq"
	"gopkg.in/mgo.v2"
)

var db *mgo.Session

// dialdb connect to local MongoDB 
func dialdb() error {
	var err error
	log.Println("dial to MongoDB: localhost")
	db, err = mgo.Dial("localhost")
	return err
}

// closedb close local MongoDB connection
func closedb() {
	db.Close()
	log.Println("DB connection has been closed")
}

type poll struct {
	Options []string
}

// loadOptions load poll options from "ballots" collection 
func loadOptions() ([]string, error) {
	var options []string
	// Find(nil) means no filtering
	iter := db.DB("ballots").C("polls").Find(nil).Iter()
	
	var p poll
	for iter.Next(&p) {
		options = append(options, p.Options...)
	}
	iter.Close()
	return options, iter.Err()
}

var mstdnClient *mastodon.Client

// startMstdnStream start subscription to mstdn event stream
func startMstdnStream(ctx context.Context, votes chan<- string) <-chan struct{} {
	stoppedchan := make(chan struct{}, 1)
	go func() {
		defer func() {
			stoppedchan <- struct{}{}
		}()
		options, err := loadOptions()
		if err != nil {
			log.Println("failed to load options:", err)
			return
		}
		eventchan, err := mstdnClient.StreamingPublic(ctx, false)
		for {
			select {
			case <-ctx.Done():
				log.Println("stopping subscription to mstdn stream...")
				return
			case ev := <-eventchan:
				upd, ok := ev.(*mastodon.UpdateEvent)
				if ok {
					for _, option := range options {
						content := upd.Status.Content
						if strings.Contains(strings.ToLower(content), strings.ToLower(option)) {
							log.Println("vote: ", option)
							votes <- option
						}
					}
				}
			}
		}
	}()
	return stoppedchan
}

// publishVotes publish votes from mstdn to NSQ
func publishVotes(votes <-chan string) <-chan struct{} {
	stoppedchan := make(chan struct{}, 1)
	pub, _ := nsq.NewProducer("localhost:4150", nsq.NewConfig())
	go func() {
		defer func() {
			stoppedchan <- struct{}{}
		}()
		for vote := range votes {
			pub.Publish("votes", []byte(vote))
		}
		log.Println("Publisher: stopping")
		pub.Stop()
		log.Println("Publisher: stopped")
	}()
	return stoppedchan
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	if err := dialdb(); err != nil {
		log.Fatalf("failed to dial MongoDB: %v", err)
	}
	defer closedb()

	mstdnClient = mastodon.NewClient(&mastodon.Config{
		Server:       "https://mstdn.jp",
		ClientID:     os.Getenv("MSTDN_CLIENT_ID"),
		ClientSecret: os.Getenv("MSTDN_CLIENT_SECRET"),
	})
	err = mstdnClient.Authenticate(context.Background(), os.Getenv("MSTDN_USER_EMAIL"), os.Getenv("MSTDN_USER_PASSWORD"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	signalchan := make(chan os.Signal, 1)
	go func() {
		<-signalchan
		log.Println("stopping...")
		cancel()
	}()
	signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)

	votes := make(chan string)
	publisherStoppedchan := publishVotes(votes)
	mstdnStoppedchan := startMstdnStream(ctx, votes)
	<-mstdnStoppedchan
	close(votes)
	<-publisherStoppedchan
}

