package main

import (
	"fmt"
	"log"
	_ "encoding/json"
	"os"
	_ "strings"

	"github.com/gocql/gocql"
)

type tweetStruct struct {
	timeline string `json:"timeline"`
	id  gocql.UUID  `json:"id"`
	text string     `json:"text"`
}

func (tw tweetStruct) isEmpty() bool {
    return tw.id == (gocql.UUID{})
}

func (tw tweetStruct) Println() int {
	fmt.Printf("Tweet>> %+v, %+s, %+s \n", tw.id, tw.text, tw.timeline)
	return 0
}

func main() {

	arguments := os.Args
	var sUsername, sPassword, sHost string;
	for i:=1;len(arguments) > i;i++ {
		switch arguments[i] {
		case "-u":
				sUsername = arguments[i+1]
		case "-p":
				sPassword = arguments[i+1]
		case "-h":
				sHost = arguments[i+1]
		}
	}

  const cConsistency gocql.Consistency = gocql.One
  var id gocql.UUID
	var text string
	tweets := make([]tweetStruct, 1)

	cluster := gocql.NewCluster(sHost)
  cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: sUsername,
		Password: sPassword,
	}
	cluster.Keyspace = "example"
	cluster.Consistency = cConsistency
	session, err := cluster.CreateSession()

	err = session.Query(`INSERT INTO tweet (timeline, id, text) VALUES (?, ?, ?)`, "me", gocql.TimeUUID(), "tweet created by simple cassandra client").Exec()
  if err != nil { log.Fatalf("Authentication error: %s", err)  }

	err = session.Query(`SELECT id, text FROM tweet WHERE timeline = ? LIMIT 1`, "me").Consistency(cConsistency).Scan(&id, &text)
  if err != nil {  log.Fatal(err)  }

	iter := session.Query(`SELECT id, text FROM tweet WHERE timeline = ?`, "me").Iter()
	for i := 0;iter.Scan(&id, &text);i++ {
    tweets = append(tweets, tweetStruct{"me", id, text})
	}
	if err := iter.Close(); err != nil { log.Fatal(err) }

	for i:=0;i<len(tweets);i++ {
		tweets[i].Println()
	}

  session.Close()
}
