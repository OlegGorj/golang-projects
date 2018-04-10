
package main

import (
	_ "crypto/rand"
	_ "crypto/sha256"
	_ "encoding/base64"
	_ "encoding/hex"
	"encoding/json"
	_ "errors"
	"flag"
	"github.com/gocql/gocql"
	"io/ioutil"
	"log"
	_ "net/http"
	"strings"
	_ "time"
)

// Struct to represent configuration file params
type configFile struct {
	Port        string
	Serverslist string
	Keyspace    string
	Username    string
	Password    string
}

// Simple structure to represent login and add user response
type newUserRequest struct {
	Username string
	Password string
}

//------------------------------------------------------------------------------------------------
// CONFIG section
//------------------------------------------------------------------------------------------------
func readConfig(confFilePath string) (configFile, error) {
	var config configFile
	confFile, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return config, err
	}
	json.Unmarshal(confFile, &config)
	return config, nil
}
//------------------------------------------------------------------------------------------------
// datastructures section
//------------------------------------------------------------------------------------------------
func createDatastructure(session *gocql.Session, keyspace string) error {
	err := session.Query("CREATE KEYSPACE IF NOT EXISTS " + keyspace +
		" WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }").Exec()
	if err != nil {
		return err
	}

	err = session.Query("CREATE TABLE IF NOT EXISTS " + keyspace + ".users (" +
		"username varchar," +
		"password varchar," +
		"PRIMARY KEY(username))").Exec()
	if err != nil {
		return err
	}

	err = session.Query("CREATE TABLE IF NOT EXISTS " + keyspace + ".sessions (" +
		"session_id varchar PRIMARY KEY," +
		"username varchar)").Exec()
	return err
}
//------------------------------------------------------------------------------------------------
// Handlers section
//------------------------------------------------------------------------------------------------
// Router for /session/ functions. Routing based on request method, i.e. GET, POST, PUT, DELETE.
func sessionHandler(w http.ResponseWriter, r *http.Request, session *gocql.Session) {

}
func userHandler(w http.ResponseWriter, r *http.Request, session *gocql.Session) {

}
//------------------------------------------------------------------------------------------------
// MAIN
//------------------------------------------------------------------------------------------------
func main() {

	confFilePath := flag.String("conf", "config.json", "path to application config")
	flag.Parse()
	config, err := readConfig(*confFilePath)
	if err != nil {
		log.Fatal("Couldn't read config file ", err)
	}
	// Initialize Cassandra cluster
	cluster := gocql.NewCluster(strings.Split(config.Serverslist, ",")...)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Username,
		Password: config.Password,
	}
	// Establish connection to Cassandra
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	// create datastructures
	err = createDatastructure(session, config.Keyspace)
	if err != nil {
		log.Fatal("Get an error while creating datastructures: ", err)
	}
  session.Close()

	cluster.Keyspace = config.Keyspace
	session, _ = cluster.CreateSession()
	defer session.Close()

	// HTTP section starts here...
	// If someone ask root, reply 404
	http.HandleFunc("/", http.NotFound)
  // handle /users endpoint
	http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		userHandler(w, r, session)
	})
  // handle /session endpoint
	http.HandleFunc("/session/", func(w http.ResponseWriter, r *http.Request) {
		sessionHandler(w, r, session)
	})

	err = http.ListenAndServe(":"+config.Port, nil)
	if err != nil {
		log.Fatal("Error on creating listener: ", err)
	}

}
