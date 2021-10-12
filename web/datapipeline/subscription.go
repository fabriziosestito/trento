package datapipeline

import "time"

type Subscription struct {
	LastSeenDataCollectedEventID int64
	AgentID                      string `gorm:"primaryKey"`
	ProjectorID                  string `gorm:"primaryKey"`
	UpdatedAt                    time.Time
}
