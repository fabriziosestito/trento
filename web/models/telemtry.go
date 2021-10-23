package models

type Telemetry struct {
	AgentID       string `gorm:"primaryKey"`
	CloudProvider string
}
