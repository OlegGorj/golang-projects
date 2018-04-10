
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
	"net/http"
	"strings"
	_ "time"
	"io"
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
  //body, _ := ioutil.ReadAll(r.Body)

}
// Router for /user/ functions. Routing based on request method, i.e. GET, POST, PUT, DELETE.
func userHandler(w http.ResponseWriter, r *http.Request, session *gocql.Session) {
  //body, _ := ioutil.ReadAll(r.Body)
	switch {

	case r.Method == "GET":
		// GET request
		var username, password string
		// Get users list
		iter := session.Query("SELECT * from users ").Iter()
		//if err != nil {
		//	error_code := http.StatusInternalServerError
		//	http.Error(w, http.StatusText(error_code), error_code)
		//}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "Existing users are:\n")
		for i := 0;iter.Scan(&username, &password);i++ {
			io.WriteString(w, username + "\n")
		}
		if err := iter.Close(); err != nil { log.Fatal(err) }

	case r.Method == "POST":
		// POST method

	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
//------------------------------------------------------------------------------------------------
// MAIN
//------------------------------------------------------------------------------------------------
func main() {

	log.Println("API Service starting..")

	confFilePath := flag.String("conf", "config.json", "path to application config")
	flag.Parse()
	config, err := readConfig(*confFilePath)
	if err != nil {
		log.Fatal("Couldn't read config file ", err)
	}
	log.Println("Configs initialized.")

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
	log.Println("Session to backend created.")
	// create datastructures
	err = createDatastructure(session, config.Keyspace)
	if err != nil {
		log.Fatal("Get an error while creating datastructures: ", err)
	}
  session.Close()
	log.Println("Backend datastructures created.")

	cluster.Keyspace = config.Keyspace
	session, _ = cluster.CreateSession()
	defer session.Close()
	log.Println("Keyspace for backend is set.")

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

	log.Println("API Service shuting down..")

}
