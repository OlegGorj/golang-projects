
package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
  "encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"github.com/gocql/gocql"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"io"
	_ "fmt"
)

// Struct to represent configuration file params
type configFile struct {
	Port          string
	Serverslist   string
	Keyspace      string
	Username      string
	Password      string
	keyspace      string
	userstable    string
	sessionstable string
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
	err := session.Query("CREATE KEYSPACE IF NOT EXISTS " + keyspace + " WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }").Exec()
	if err != nil { return err }

	err = session.Query("CREATE TABLE IF NOT EXISTS " + keyspace + ".users (id UUID, username varchar, password varchar, active boolean, ts timestamp, PRIMARY KEY (id) )").Exec()
	if err != nil { return err }

	err = session.Query("CREATE INDEX IF NOT EXISTS ON testapp.users (username)").Exec()
	if err != nil { return err }

	err = session.Query("CREATE INDEX IF NOT EXISTS ON testapp.users (active)").Exec()
	if err != nil { return err }

	err = session.Query("CREATE INDEX IF NOT EXISTS ON testapp.users (ts)").Exec()
	if err != nil { return err }

	err = session.Query("CREATE TABLE IF NOT EXISTS " +keyspace + ".sessions (sessionID varchar PRIMARY KEY, username varchar)").Exec()

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

	if count > 0 { return http.StatusConflict, errors.New("User with same name already exists") }

	// Prepare password hash to write it to DB
	hash := sha512.New()
	hash.Write([]byte(request.Password))
  err = session.Query("INSERT INTO users (id, username, password, active, ts) VALUES (?, ?, ?, true, toTimestamp(now()) )",
		gocql.TimeUUID(), request.Username, hex.EncodeToString(hash.Sum(nil)) ).Exec()
	if err != nil { return http.StatusInternalServerError, err }

	return http.StatusCreated, nil
}

//------------------------------------------------------------------------------------------------
func deleteUser(body *[]byte, session *gocql.Session) (int, error) {
	var request userStruct
	var id string

	err := json.Unmarshal(*body, &request)
	if err != nil { return http.StatusBadRequest, err }

	// Here should be call of function to extended validation, but nothing was in requirements
	if request.Username == "" { return http.StatusBadRequest, errors.New("User name is empty") }

	// Check if such user existing in db
	err = session.Query("SELECT id from users where username = '" + request.Username + "'").Scan(&id)
	if err != nil { return http.StatusInternalServerError, err }
	if id == "" {
		return http.StatusBadRequest, errors.New("No such user (" + request.Username + ")")
	}
	err = session.Query("UPDATE testapp.users SET active = false WHERE id = " + id).Exec()
	if err != nil { return http.StatusInternalServerError, err }
  return http.StatusOK, err
}

//------------------------------------------------------------------------------------------------
// Session Functions section
//------------------------------------------------------------------------------------------------
// Generation random session ID and verifiyng that it is unique
func generateSessionID(session *gocql.Session) (string, error) {
	var sessionID string
	count := 2
	size := 32
	rb := make([]byte, size)

	// generating sessionID while it will be uniq(actually in most cases it will be uniq in a first time)
	for count != 0 {
		rand.Read(rb)
		sessionID = base64.URLEncoding.EncodeToString(rb)
		err := session.Query("SELECT COUNT(*) from sessions where sessionID = '" + sessionID + "'").Scan(&count)
		if err != nil {
			return sessionID, err
		}
	}

	return sessionID, nil
}
//------------------------------------------------------------------------------------------------
func deleteSession(session *gocql.Session, sessionID string) (int, error) {
	// fast path to don't use DB when sessionID cookie not indicated at all
	if sessionID == "" {
		return http.StatusUnauthorized, nil
	}

	// removing session for DB
	err := session.Query("DELETE FROM sessions WHERE sessionID = '" + sessionID + "'").Exec()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil

}
//------------------------------------------------------------------------------------------------
// Function handling creating new session
func createSession(body *[]byte, session *gocql.Session) (string, int, error) {
	var request userStruct
	var sessionID string
	var password, userid string

	err := json.Unmarshal(*body, &request)
	if err != nil {
		return sessionID, http.StatusBadRequest, err
	}

	// Here should be call of function to extended validation, but nothing was in requirements
	if request.Password == "" || request.Username == "" { return sessionID, http.StatusBadRequest, errors.New("User or password is empty") }

	// Prepare password hash
	hash := sha512.New()
	hash.Write([]byte(request.Password))

	// Check if user and password is valid
// this line won't work! cuz cassandra datamodel wasn't designed to support select query by username+password
//	err = session.Query("SELECT COUNT(*) from users where username = '" + request.Username + "' and password ='" + hex.EncodeToString(hash.Sum(nil)) + "'").Scan(&count)
// hence, solution is to get password hash and compare with hash of provided password
	err = session.Query("SELECT id, password from users where username = '" + request.Username + "' ").Scan(&userid, &password)
	if err != nil { return sessionID, http.StatusInternalServerError, err }
	if userid == "" {return sessionID, http.StatusUnauthorized, errors.New("User does not exist")}
	if password != hex.EncodeToString(hash.Sum(nil)) { return sessionID, http.StatusUnauthorized, errors.New("User password does not match") }
	// prepare session ID for a new session
	sessionID, err = generateSessionID(session)
	if err != nil {
		return sessionID, http.StatusInternalServerError, err
	}

	// set TTL to a one year to expire in same time with cookie
	err = session.Query("INSERT INTO sessions (sessionID,username) VALUES (?,?) USING TTL 31536000", sessionID, request.Username).Exec()
	if err != nil { return sessionID, http.StatusInternalServerError, err  }

	return sessionID, http.StatusCreated, nil
}
//------------------------------------------------------------------------------------------------
// Function checking if provided cookie matching to active sessions
func checkSession(session *gocql.Session, sessionID string) (int, error) {
	var count int
	// fast path to don't use DB when sessionID cookie not indicated at all
	if sessionID == "" { return http.StatusUnauthorized, nil }

	// Check if such session exist
	err := session.Query("SELECT COUNT(*) from sessions where sessionID = '" + sessionID + "'").Scan(&count)
	if err != nil { return http.StatusInternalServerError, err }

	if count == 0 {
		return http.StatusUnauthorized, nil
	}
	return http.StatusOK, nil

}

//------------------------------------------------------------------------------------------------
// Handlers section
//------------------------------------------------------------------------------------------------
// Router for /session/ functions. Routing based on request method, i.e. GET, POST, PUT, DELETE.
func sessionHandler(w http.ResponseWriter, r *http.Request, session *gocql.Session) {
	body, _ := ioutil.ReadAll(r.Body)

	switch {
		case r.Method == "POST":
				sessionID, errorCode, err := createSession(&body, session)
				if err != nil { log.Println("Error on creating session: ", err, "\nClient: ", r.RemoteAddr, " Request: ", string(body))  }

				// Set expire for a one year, same as in sessions table
				if sessionID != "" {
					expire := time.Now().AddDate(1, 0, 0)
					authCookie := &http.Cookie{
						Name:    "sessionID",
						Expires: expire,
						Value:   sessionID,
					}
					http.SetCookie(w, authCookie)
				}
				http.Error(w, http.StatusText(errorCode), errorCode)

		case r.Method == "GET":
				sessionID, err := r.Cookie("sessionID")
				//  this is critial - when cookie doesnot exist, sessionID returned as nil, hence next call (checkSession) will fail
				if err != nil { log.Println("Cookie returned empty sessionID: ", err); return }

				errorCode, err := checkSession(session, sessionID.Value)
				if err != nil { log.Println("Error on checking authorization: ", err) }
				http.Error(w, http.StatusText(errorCode), errorCode)

		case r.Method == "DELETE":
				sessionID, _ := r.Cookie("sessionID")
				errorCode, err := deleteSession(session, sessionID.Value)
				if err != nil {
					log.Println("Error on checking authorization: ", err)
				}

				// Rewrite sessionID cookie with empty sting and set expiration now
				expire := time.Now()

				authCookie := &http.Cookie{
					Name:    "sessionID",
					Expires: expire,
					Value:   "",
				}

				http.SetCookie(w, authCookie)

				http.Error(w, http.StatusText(errorCode), errorCode)

		default:
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

}
// Router for /user/ functions. Routing based on request method, i.e. GET, POST, PUT, DELETE.
func userHandler(w http.ResponseWriter, r *http.Request, session *gocql.Session) {
  body, _ := ioutil.ReadAll(r.Body)
	switch {

	  case r.Method == "GET":
	  	// GET request
	  	var username string
	  	// Get users list
	  	iter := session.Query("SELECT username FROM users WHERE active = true").Iter()
	  	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	  	w.WriteHeader(http.StatusOK)
	  	io.WriteString(w, "Existing users are:\n")
	  	for i := 0;iter.Scan(&username);i++ {
	  		io.WriteString(w, username + "\n")
	  	}
	  	if err := iter.Close(); err != nil { log.Fatal(err) }

	  case r.Method == "POST":
	  	errorcode, err := createUser(&body, session)
	  	if err != nil {
	  		log.Println("Error on creating user: ", err, "\nClient: ", r.RemoteAddr, " Request: ", string(body) )
	  	}
	  	http.Error(w, http.StatusText(errorcode), errorcode)

	  case r.Method == "DELETE":
			errorcode, err := deleteUser(&body, session)
	  	if err != nil {
	  		log.Println("Error on deleting user: ", err, "\nClient: ", r.RemoteAddr, " Request: ", string(body) )
	  	}
	  	http.Error(w, http.StatusText(errorcode), errorcode)

	  default:
	  	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

	}
}
//------------------------------------------------------------------------------------------------
// MAIN
//------------------------------------------------------------------------------------------------
func main() {
  var err error
	var configuration configFile

	log.Println("API Service starting..")

	confFilePath := flag.String("conf", "config.json", "path to application config")
	flag.Parse()
	configuration, err = readConfig(*confFilePath)
	if err != nil {
		log.Fatal("Couldn't read config file ", err)
	}
	log.Println("Configs initialized.")

	// Initialize Cassandra cluster
	cluster := gocql.NewCluster(strings.Split(configuration.Serverslist, ",")...)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: configuration.Username,
		Password: configuration.Password,
	}
	// Establish connection to Cassandra
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Session to backend created.")
	// create datastructures
	err = createDatastructure(session, configuration.Keyspace)
	if err != nil {
		log.Fatal("Get an error while creating datastructures: ", err)
	}
  session.Close()
	log.Println("Backend datastructures created.")

	cluster.Keyspace = configuration.Keyspace
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

	err = http.ListenAndServe(":" + configuration.Port, nil)
	if err != nil {
		log.Fatal("Error on creating listener: ", err)
	}

	log.Println("API Service shuting down..")

}
