package models

import (
	"time"

	"github.com/lib/pq"
)

type Host struct {
	AgentID       string `gorm:"primaryKey"`
	Name          string
	IPAddresses   pq.StringArray `gorm:"type:text[]"`
	CloudProvider string
	ClusterID     string
	ClusterName   string
	SIDs          pq.StringArray `gorm:"column:sids; type:text[]"`
	AgentVersion  string
	UpdatedAt     time.Time
}
