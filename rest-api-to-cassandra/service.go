
package main

import (
	"crypto/rand"
	"crypto/sha256"
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
)

// Struct to represent configuration file params
type configFile struct {
	Port        string
	Serverslist string
	Keyspace    string
}

// Simple structure to represent login and add user response
type newUserRequest struct {
	Username string
	Password string
}
