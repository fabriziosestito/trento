package models

import "time"

type Telemetry struct {
	AgentID       string `gorm:"primaryKey"`
	HostName      string
	SLESVersion   string
	CPUCount      int
	SocketCount   int
	TotalMemoryMB int
	CloudProvider string
	UpdatedAt     time.Time
}

func (Telemetry) TableName() string {
	return "telemetry"
}
