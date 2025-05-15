package model

type BookingData struct {
	ID     int64  `gorm:"column:id" json:"id"`
	SlotID string `gorm:"column:slot_id" json:"slot_id"`
	UserID string `gorm:"column:user_id" json:"user_id"`
}

func (*BookingData) TableName() string {
	return "booking_table"
}
