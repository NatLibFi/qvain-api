package main

import (
	"log"
	"strings"
	"errors"
	"time"
	"net/http"
	
	"github.com/gorilla/websocket"
	"github.com/patrickmn/go-cache"
)


const BEARER_PREFIX = "Bearer "

var (
	errMissingWSToken = errors.New("Missing token")
	errInvalidWSToken = errors.New("Invalid token")
)

var OpenCache = cache.New(30*time.Second, 1*time.Minute)

func init() {
}


var upgrader = websocket.Upgrader{}
/*
{
	CheckOrigin: func(r *http.Request) bool {
		log.Println("Origin:", r.Header.Get("Origin"))
		if r.Header.Get("Origin") != "http://"+r.Host {
			http.Error(w, "Origin not allowed", 403)
			return
		}
		return true
	},
}
*/
/*
if req.Header.Get("Origin") != "http://"+req.Host {
	http.Error(w, "Origin not allowed", 403)
	return
}
*/

// If opening a websocket from a local html file:
//   Origin: file://

func parseToken(token string) (string, error) {
	if len(token) < 1 {
		return "", errInvalidWSToken
	}
	log.Println("parseToken:", token)
	return "wouter", nil
}


func echo(w http.ResponseWriter, r *http.Request) {
	log.Println("Origin:", r.Header.Get("Origin"), r.Host)
	log.Println("Subprotocols:", websocket.Subprotocols(r))
	_ = strings.HasPrefix("lwkejr", "lw")
	
	var user string
	var haveToken bool
	for _, proto := range websocket.Subprotocols(r) {
		if !strings.HasPrefix(proto, BEARER_PREFIX) {
			continue
		}
		
		token, err := parseToken(proto[len(BEARER_PREFIX):])
		if err != nil {
			//http.Error(w, errInvalidToken.Error(), http.StatusForbidden)
			//return
		}
		// pass
		user = token
		haveToken = true
	}
	if !haveToken {
		//http.Error(w, errMissingToken.Error(), http.StatusUnauthorized)
		//return
	}
	if user == "" {
		//http.Error(w, errInvalidToken.Error(), http.StatusUnauthorized)
		//return
	}
	
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		
		if strings.HasPrefix(string(message), "open:") {
			log.Printf("opening...")
			
			/*
			err := OpenCache.Add(string(message[5:]), "uid", cache.DefaultExpiration)
			if err != nil {
				log.Printf("already open")
				message = []byte("error: already open")
			}
			*/
			if val, e := OpenCache.Get(string(message[5:])); e {
				log.Printf("already open")
				message = []byte("error: already opened by " + val.(string))
			} else {
				err := OpenCache.Add(string(message[5:]), "uid", cache.DefaultExpiration)
				if err != nil {
					log.Printf("already open (race)")
					message = []byte("error: already open (race)")
				}
			}
		}
		
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
