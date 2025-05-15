package model

type BookingData struct {
	ID     int64  `gorm:"column:id" json:"id"`
	SlotID string `gorm:"column:slotID" json:"slot_id"`
	UserID string `gorm:"column:userID" json:"user_id"`
}

func (*BookingData) TableName() string {
	return "booking_table"
}
