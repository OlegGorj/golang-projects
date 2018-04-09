package main

import (
	"fmt"
	"log"
	_ "encoding/json"
	"os"
	_ "strings"

	"github.com/gocql/gocql"
)

type tweet_struct struct {
	timeline string `json:"timeline"`
	id  gocql.UUID  `json:"id"`
	text string     `json:"text"`
}

func (s tweet_struct) isEmpty() bool {
    return s.id == (gocql.UUID{})
}

func (tw tweet_struct) Println() int {
	fmt.Printf("Tweet>> %+v, %+s, %+s \n", tw.id, tw.text, tw.timeline)
	return 0
}

func main() {

	arguments := os.Args
	var s_username, s_password, s_host string;
	for i:=1;len(arguments) > i;i++ {
		switch arguments[i] {
		case "-u":
				s_username = arguments[i+1]
		case "-p":
				s_password = arguments[i+1]
		case "-h":
				s_host = arguments[i+1]
		}
	}

  const c_consistency gocql.Consistency = gocql.One
  var id gocql.UUID
	var text string
	tweets := make([]tweet_struct, 1)

	cluster := gocql.NewCluster(s_host)
  cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: s_username,
		Password: s_password,
	}
	cluster.Keyspace = "example"
	cluster.Consistency = c_consistency
	session, err := cluster.CreateSession()

	err = session.Query(`INSERT INTO tweet (timeline, id, text) VALUES (?, ?, ?)`, "me", gocql.TimeUUID(), "tweet created by simple cassandra client").Exec()
  if err != nil { log.Fatalf("Authentication error: %s", err)  }

	err = session.Query(`SELECT id, text FROM tweet WHERE timeline = ? LIMIT 1`, "me").Consistency(c_consistency).Scan(&id, &text)
  if err != nil {  log.Fatal(err)  }

	iter := session.Query(`SELECT id, text FROM tweet WHERE timeline = ?`, "me").Iter()
	for i := 0;iter.Scan(&id, &text);i++ {
    tweets = append(tweets, tweet_struct{"me", id, text})
	}
	if err := iter.Close(); err != nil { log.Fatal(err) }

	for i:=0;i<len(tweets);i++ {
		tweets[i].Println()
	}

  session.Close()
}
