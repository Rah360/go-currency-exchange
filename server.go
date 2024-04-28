package main

import (
	"context"
	"currency-exchange-app/storage"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

/*
ServerManager keeps list of all clients
and register and unregister based on their connection status
Storage interface that will be used to store currency rate
*/
type ServerManager struct {
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Lock       sync.Mutex
	Storage    storage.CurrencyRateStorage
}

func NewServerManager(storage storage.CurrencyRateStorage) *ServerManager {
	return &ServerManager{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Storage:    storage,
	}
}

// adds and delete new clients
func (manager *ServerManager) run() {
	for {
		select {
		case client := <-manager.Register:
			manager.Clients[client.Id] = client
			fmt.Println("new client connected", client.Id)
		case client := <-manager.Unregister:
			_, ok := manager.Clients[client.Id]
			if ok {
				delete(manager.Clients, client.Id)
				fmt.Println("client disconnected", client.Id)
			}
		}
	}
}

// handles incoming connections
func (manager *ServerManager) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		Conn: ws,
		Id:   uuid.New().String(),
		Send: make(chan []byte, 256),
	}
	manager.Register <- client
}

// broadcasting messages
func (manager *ServerManager) sendToAll(message []byte) {
	manager.Lock.Lock()
	defer manager.Lock.Unlock()
	for _, client := range manager.Clients {
		err := client.Conn.Write(context.Background(), websocket.MessageText, message)
		if err != nil {
			manager.Unregister <- client
		}
	}
}

// updates in storage and publishes new values
func (manager *ServerManager) UpdateAndBroadcastRates(currency string, rate string) {
	err := manager.Storage.Update(currency, rate)
	if err != nil {
		log.Printf("Failed to update rate for %s: %v", currency, err)
		return
	}

	rates, err := manager.Storage.Get()
	if err != nil {
		log.Printf("Failed to Get rate for %s: %v", currency, err)
		return
	}

	err = manager.Storage.Publish(rates)
	if err != nil {
		log.Printf("Failed to publish rate for %s: %v", currency, err)
		return
	}
}

// endlessly will listen to redis key and then broadcast updates
func (manager *ServerManager) handleWebSocketConnections() {
	ratesChannel, err := manager.Storage.Subscribe()
	if err != nil {
		return
	}
	for ratesJson := range ratesChannel {
		manager.sendToAll([]byte(ratesJson))
	}
}
