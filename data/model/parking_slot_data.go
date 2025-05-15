package model

type ParkingSlotData struct {
	ID     int64  `gorm:"column:id" json:"-"`
	Name   string `gorm:"column:name" json:"-"`
	A1     int    `gorm:"column:a1" json:"A1"`
	A2     int    `gorm:"column:a2" json:"A2"`
	A3     int    `gorm:"column:a3" json:"A3"`
	A4     int    `gorm:"column:a4" json:"A4"`
	GateIn bool   `gorm:"column:gate_in" json:"GateIn"`
}

func (*ParkingSlotData) TableName() string {
	return "parking_slot"
}
