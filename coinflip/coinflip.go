package coinflip

import (
	"github.com/gorilla/websocket"
)

const (
	COMMISSION float64 = 0.025 
	RED        = "red"
	BLUE       = "blue"
)

var clients = make(map[*websocket.Conn]bool) // ws connected clients
var broadcast = make(chan interface{}, 20)   // ws broadcast channel
