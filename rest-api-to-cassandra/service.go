
package main

import (
	_ "crypto/rand"
	"crypto/sha512"
	_ "encoding/base64"
  "encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"github.com/gocql/gocql"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	_ "time"
	"io"
	_ "fmt"
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
type userStruct struct {
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
// Users Functions section
//------------------------------------------------------------------------------------------------
func createUser(body *[]byte, session *gocql.Session) (int, error) {
	var request userStruct
	var count int

	err := json.Unmarshal(*body, &request)
	if err != nil { return http.StatusBadRequest, err }

	// Here should be call of function to extended validation, but nothing was in requirements
	if request.Password == "" || request.Username == "" { return http.StatusBadRequest, errors.New("User or password is empty") }

	// Check if such user already existing
	err = session.Query("SELECT COUNT(*) from users where username = '" + request.Username + "'").Scan(&count)
	if err != nil { return http.StatusInternalServerError, err }

	if count > 0 { return http.StatusConflict, errors.New("Such user alredy existing") }

	// Prepare password hash to write it to DB
	hash := sha512.New()
	hash.Write([]byte(request.Password))
  err = session.Query("INSERT INTO users (username, password) VALUES (?,?)", request.Username, hex.EncodeToString(hash.Sum(nil))).Exec()
	if err != nil { return http.StatusInternalServerError, err }

	return http.StatusCreated, nil
}

//------------------------------------------------------------------------------------------------
func deleteUser(body *[]byte, session *gocql.Session) (int, error) {
	var request userStruct
	var count int

	err := json.Unmarshal(*body, &request)
	if err != nil { return http.StatusBadRequest, err }
	// Here should be call of function to extended validation, but nothing was in requirements
	if request.Username == "" { return http.StatusBadRequest, errors.New("User name is empty") }

	// Check if such user existing in db
	err = session.Query("SELECT COUNT(*) from users where username = '" + request.Username + "'").Scan(&count)
	if err != nil { return http.StatusInternalServerError, err }
	if count > 0 {
		var applied bool
		err = session.Query("UPDATE testapp.users SET active = true WHERE username = '" + request.Username + "'").Scan(&applied)
		if applied == true {
				return http.StatusOK, err
			}else{
				return http.StatusInternalServerError, errors.New("Error deleting user.")
			}
	}else{
		return http.StatusBadRequest, errors.New("No such user")
	}
  return http.StatusOK, err
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
  body, _ := ioutil.ReadAll(r.Body)
	switch {

	  case r.Method == "GET":
	  	// GET request
	  	var username string
	  	// Get users list
	  	iter := session.Query("SELECT username from users ").Iter()
	  	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	  	w.WriteHeader(http.StatusOK)
	  	io.WriteString(w, "Existing users are:\n")
	  	for i := 0;iter.Scan(&username);i++ {
	  		io.WriteString(w, username + "\n")
	  	}
	  	if err := iter.Close(); err != nil { log.Fatal(err) }

	  case r.Method == "POST":
	  	error_code, err := createUser(&body, session)
	  	if err != nil {
	  		log.Println("Error on creating user: ", err, "\nClient: ", r.RemoteAddr, " Request: ", string(body))
	  	}
	  	http.Error(w, http.StatusText(error_code), error_code)

	  case r.Method == "DELETE":
			error_code, err := deleteUser(&body, session)
	  	if err != nil {
	  		log.Println("Error on deleting user: ", err, "\nClient: ", r.RemoteAddr, " Request: ", string(body))
	  	}
	  	http.Error(w, http.StatusText(error_code), error_code)

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
