package request

type SocketRequest struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Status string `json:"status"`
	User   string `json:"user"`
	// For Frontend
	Slot string `json:"slot"`
	// For ESP32
	Slots  SlotRequest     `json:"slots"`
	Slotss []SlotRequestV2 `json:"slotss"`
	Image  string          `json:"image"`
}
