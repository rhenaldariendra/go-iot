package main

import (
	"Websocket_Service/data/model"
	"Websocket_Service/data/request"
	"Websocket_Service/data/webresponse"
	"Websocket_Service/helper"
	"encoding/json"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
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
		processedMessage, isClient := processMessage(message, DB)

		if isClient == 0 {
			// Broadcast the processed message to all other clients
			hub.broadcast <- processedMessage
		} else if isClient == 1 {
			// send message back to the sender
			err = c.conn.WriteMessage(websocket.TextMessage, processedMessage)
		}
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

func addBooking(tx *gorm.DB, data model.BookingData) error {
	err := tx.Omit("id").Save(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func updateStatusParking(tx *gorm.DB, data model.ParkingSlotData, omitMessage string) error {
	err := tx.Omit(omitMessage).Save(data).Error
	if err != nil {
		return err
	}
	return nil
}

func processMessage(message []byte, DB *gorm.DB) ([]byte, int) {
	var socketMessage request.SocketRequest
	isClient := 0

	err := helper.ReadJSONFromByte(message, &socketMessage)
	if err != nil {
		log.Printf("Error Message Not Valid: %v", err)
		return []byte("Error Message Not Valid"), 0
	}
	nameParking := "PARKING_DEMO"
	nameID := 1

	switch socketMessage.Type {
	case "status_update":
		omitMsg := "gate_in"
		a1, _ := helper.StringToInt(socketMessage.Slots.A1)
		if a1 == 404 {
			omitMsg += ",a1"
			a1 = 0
		}

		a2, _ := helper.StringToInt(socketMessage.Slots.A2)
		if a2 == 404 {
			omitMsg += ",a2"
			a2 = 0
		}

		a3, _ := helper.StringToInt(socketMessage.Slots.A3)
		if a3 == 404 {
			omitMsg += ",a3"
			a3 = 0
		}

		a4, _ := helper.StringToInt(socketMessage.Slots.A4)
		if a4 == 404 {
			omitMsg += ",a4"
			a4 = 0
		}

		data := model.ParkingSlotData{
			ID:   int64(nameID),
			Name: nameParking,
			A1:   a1,
			A2:   a2,
			A3:   a3,
			A4:   a4,
		}

		err = updateStatusParking(DB, data, omitMsg)

		jsonData, _ := json.Marshal(data)

		message = jsonData
	case "book_request":
		// Handle booking request
		bookData := model.BookingData{
			SlotID: socketMessage.Slot,
			UserID: socketMessage.User,
		}

		err = addBooking(DB, bookData)

		omitMsg := "gate_in"
		data := model.ParkingSlotData{}
		bookingRes := webresponse.BookingResponse{}
		switch socketMessage.Slot {
		case "A1":
			omitMsg += ",a2,a3,a4"
			data = model.ParkingSlotData{
				ID:     int64(nameID),
				Name:   nameParking,
				A1:     2,
				A2:     0,
				A3:     0,
				A4:     0,
				GateIn: false,
			}
			bookingRes = webresponse.BookingResponse{
				Type:   "command",
				Action: "book_slot",
				Slot:   "A1",
			}
		case "A2":
			omitMsg += ",a1,a3,a4"
			data = model.ParkingSlotData{
				ID:     int64(nameID),
				Name:   nameParking,
				A1:     0,
				A2:     2,
				A3:     0,
				A4:     0,
				GateIn: false,
			}
			bookingRes = webresponse.BookingResponse{
				Type:   "command",
				Action: "book_slot",
				Slot:   "A2",
			}
		case "A3":
			omitMsg += ",a1,a2,a4"
			data = model.ParkingSlotData{
				ID:     int64(nameID),
				Name:   nameParking,
				A1:     0,
				A2:     0,
				A3:     2,
				A4:     0,
				GateIn: false,
			}
			bookingRes = webresponse.BookingResponse{
				Type:   "command",
				Action: "book_slot",
				Slot:   "A3",
			}

		case "A4":
			omitMsg += ",a1,a2,a3"
			data = model.ParkingSlotData{
				ID:     int64(nameID),
				Name:   nameParking,
				A1:     0,
				A2:     0,
				A3:     0,
				A4:     2,
				GateIn: false,
			}
			bookingRes = webresponse.BookingResponse{
				Type:   "command",
				Action: "book_slot",
				Slot:   "A4",
			}

		}
		err = updateStatusParking(DB, data, omitMsg)
		jsonData, _ := json.Marshal(bookingRes)
		message = jsonData
	case "command":
	case "init":
		// Handle initialization
		parkingData := model.ParkingSlotData{}
		err = DB.Table("parking_slot").Where("name = ?", nameParking).First(&parkingData).Error
		isClient = 1

		jsonData, _ := json.Marshal(parkingData)
		message = jsonData
	case "gate_status":
	case "image_upload":
		// Handle image upload
		// Convert base64 string to byte array and save to file
		data, err := helper.ConvertBase64ToBytes(socketMessage.Image)
		filePath := "assets/" + socketMessage.User + time.Now().Format("150405") + ".jpg"
		err = helper.SaveBytesToFile(data, filePath)
		if err != nil {
			log.Printf("Error saving file: %v", err)
			return []byte("Error saving file"), 0
		}
		isClient = 404
	case "ping":
		// Handle ping
		jsonData, _ := json.Marshal(map[string]interface{}{
			"nama":      "tio novriadi putra",
			"status":    "single",
			"mantan":    "naura",
			"isBalikan": "yes",
		})
		isClient = 1
		message = jsonData

	}

	return message, isClient
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

	r := chi.NewRouter()
	r.Get("/iot/socket/channel", app.handleWebSocket)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Println("Server started on :8080")
	log.Fatal(srv.ListenAndServe())

}
