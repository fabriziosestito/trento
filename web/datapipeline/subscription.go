package datapipeline

import "time"

type Subscription struct {
	LastSeenDataID int64
	AgentID        string `gorm:"primaryKey"`
	DiscoveryType  string `gorm:"primaryKey"`
	SeenAt         time.Time
}
