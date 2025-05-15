package main

import (
	"Websocket_Service/data/request"
	"Websocket_Service/helper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type MainApp struct {
	DB *gorm.DB
}

// Client represents a connected WebSocket client
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub maintains the set of active clients
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

var hub = &Hub{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Println("Client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Println("Client disconnected")

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (m *MainApp) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump(m.DB)
	go client.readPump(m.DB)
}

func (c *Client) readPump(DB *gorm.DB) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Do some processing with the message
		processedMessage := processMessage(message)

		// Broadcast the processed message to all other clients
		hub.broadcast <- processedMessage
	}
}

func (c *Client) writePump(DB *gorm.DB) {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func processMessage(message []byte) []byte {

	var socketMessage request.SocketRequest

	err := helper.ReadJSONFromByte(message, &socketMessage)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		return []byte("Error processing message")
	}

	switch socketMessage.Type {

	}

	return append([]byte("Processed: "), message...)
}

func main() {
	go hub.run()

	dbConf := helper.ReadConfigDB()
	dsn := "host=" + dbConf.DBHost +
		" user=" + dbConf.DBUser +
		" dbname=" + dbConf.DBName +
		" password=" + dbConf.DBPassword +
		" port=" + dbConf.DBPort +
		" sslmode=disable TimeZone=UTC"

	var err error
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	app := &MainApp{
		DB: DB,
	}

	log.Println("Database connection established")

	http.HandleFunc("/iot/socket/channel", app.handleWebSocket)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
