package request

type SlotRequest struct {
	SlotName string `json:"slot_name"`
	Status   int    `json:"status"`
}
