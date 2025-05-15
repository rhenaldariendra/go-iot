package webresponse

type BookingResponse struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Slot   string `json:"slot"`
}
