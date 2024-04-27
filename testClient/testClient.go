package main

import (
	"context"
	"log"

	"nhooyr.io/websocket"
)

func main() {
	ctx := context.Background()
	ws, _, err := websocket.Dial(ctx, "ws://localhost:3000/ws", nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	defer ws.Close(websocket.StatusNormalClosure, "")

	err = ws.Write(ctx, websocket.MessageText, []byte("Hello Server!"))
	if err != nil {
		log.Fatal("Error sending message:", err)
	}

	for {
		_, msg, err := ws.Read(ctx)
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		log.Printf("Received Updated Currency Rates: %s", msg)
	}
}
