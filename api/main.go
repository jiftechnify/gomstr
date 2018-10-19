package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/stretchr/graceful"
	mgo "gopkg.in/mgo.v2"
)

func main() {
	var (
		addr  = flag.String("addr", ":8080", "the address of the endpoint")
		mongo = flag.String("mongo", "localhost", "the address of the MongoDB")
	)
	flag.Parse()

	log.Println("connecting to MongoDB...", *mongo)
	db, err := mgo.Dial(*mongo)
	if err != nil {
		log.Fatalln("failed to connect MongoDB:", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/polls/", withCORS(withVars(withData(db, withAPIKey(handlePolls)))))
	log.Println("starting Web server...:", *addr)
	graceful.Run(*addr, 1*time.Second, mux)
	log.Println("stopping...")
}

// validates API key before the main process. If API key is invalid, the main process is skipped.
func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isValidAPIKey(r.URL.Query().Get("key")) {
			respondErr(w, r, http.StatusUnauthorized, "invalid API key")
			return
		}
		fn(w, r)
	}
}

func isValidAPIKey(key string) bool {
	return key == "abc123"
}

// before the main process, opens connection to MongoDB and registers it to request scoped vars.
// after the main process, the connection will be automatically closed.
func withData(d *mgo.Session, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		thisDb := d.Copy()
		defer thisDb.Close()

		SetVar(r, "db", thisDb.DB("ballots"))
		fn(w, r)
	}
}

// handler that calls OpenVars before the main process, and calls CloseVars after all.
func withVars(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		OpenVars(r)
		defer CloseVars(r)
		fn(w, r)
	}
}

// handler that sets headers in order to utilize CORS.
func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Heanders", "Location")
		fn(w, r)
	}
}
