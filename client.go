package main

import (
	"nhooyr.io/websocket"
)

type Client struct {
	Id   string
	Conn *websocket.Conn
	Send chan []byte
}
