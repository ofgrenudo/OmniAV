package models

// Group, Equipment, and Request are pointers so they're omitted from JSON entirely when not
// Preloaded, instead of serializing as a misleading zero-value record.
type RequestedEquipment struct {
	ID uint `json:"id"`
	// GroupID records what the requester actually asked for (e.g. the "Cow Cart" group);
	// EquipmentID records the specific unit (e.g. "Cow Cart B") the assignment logic picked.
	GroupID     uint            `json:"groupId" gorm:"not null"`
	Group       *EquipmentGroup `json:"group,omitempty" gorm:"foreignKey:GroupID;references:ID"`
	EquipmentID uint            `json:"equipmentId" gorm:"not null"`
	Equipment   *Equipment      `json:"equipment,omitempty" gorm:"foreignKey:EquipmentID;references:ID"`
	RequestID   uint            `json:"requestId" gorm:"not null"`
	Request     *Request        `json:"request,omitempty" gorm:"foreignKey:RequestID;references:ID"`
}
