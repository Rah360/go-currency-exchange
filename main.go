package main

import (
	"currency-exchange-app/storage"
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte("test_key") // This should be stored securely

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}
type CurrencyPayload struct {
	Currency string `json:"currency"`
	Rate     string `json:"rate"`
}

func isAdminAuthenticated(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := &Claims{}

		tknStr := r.Header.Get("Authorization")
		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if !tkn.Valid || claims.Role != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		endpoint(w, r)
	})
}

func main() {
	// using redis for storage and for pubsub
	redisStore := storage.NewRedisRateStore("localhost:6379")
	// setting up our websocket server
	// with use of repository pattern we can change data storage
	// so business logic wont be tightly coupled with data storage
	server := NewServerManager(redisStore)
	// runs a goroutine that will handle clients
	go server.run()
	// goroutine for redis subscriber
	go server.handleWebSocketConnections()

	// all new clients will be connected using this handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.handleConnections(w, r)
	})

	// this method will update currency rate in redis and then publish it to subscribers
	http.HandleFunc("/update", isAdminAuthenticated(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported", http.StatusMethodNotAllowed)
			return
		}

		var updates CurrencyPayload
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}
		server.UpdateAndBroadcastRates(updates.Currency, updates.Rate)
		defer r.Body.Close()
		w.WriteHeader(http.StatusOK)
	}).ServeHTTP)

	log.Fatal(http.ListenAndServe(":3000", nil))
}
