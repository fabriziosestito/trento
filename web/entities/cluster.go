package entities

import (
	"time"

	"github.com/lib/pq"
	"github.com/trento-project/trento/web/models"
	"gorm.io/datatypes"
)

type Cluster struct {
	ID              string `gorm:"primaryKey"`
	Name            string
	ClusterType     string
	SIDs            pq.StringArray `gorm:"column:sids; type:text[]"`
	ResourcesNumber int
	HostsNumber     int
	Tags            []*models.Tag `gorm:"polymorphic:Resource;polymorphicValue:clusters"`
	UpdatedAt       time.Time
	Hosts           []*Host        `gorm:"foreignkey:cluster_id"`
	Detail          datatypes.JSON `json:"payload" binding:"required"`
}

type ClusterDetailHANA struct {
	SystemReplicationMode          string                 `json:"system_replication_mode"`
	SystemReplicationOperationMode string                 `json:"system_replication_operation_mode"`
	SecondarySyncState             string                 `json:"secondary_sync_state"`
	SRHealthState                  string                 `json:"sr_health_state"`
	CIBLastWritten                 time.Time              `json:"cib_last_written"`
	StonithType                    string                 `json:"stonith_type"`
	StoppedResources               []*ClusterNodeResource `json:"stopped_resources"`
	Nodes                          []*ClusterNode         `json:"nodes"`
}

type ClusterNode struct {
	Name       string                 `json:"name"`
	Attributes map[string]string      `json:"attributes"`
	Resources  []*ClusterNodeResource `json:"resources"`
	VirtualIPs []string               `json:"virtual_ips"`
}

type ClusterNodeResource struct {
	Id        string `json:"id"`
	Type      string `json:"type"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	FailCount int    `json:"fail_count"`
}

func (h *Cluster) ToModel() *models.Cluster {
	// TODO: move to Tags entity when we will have it
	var tags []string
	for _, tag := range h.Tags {
		tags = append(tags, tag.Value)
	}

	return &models.Cluster{
		ID:              h.ID,
		Name:            h.Name,
		ClusterType:     h.ClusterType,
		SIDs:            h.SIDs,
		ResourcesNumber: h.ResourcesNumber,
		HostsNumber:     h.HostsNumber,
		Tags:            tags,
	}
}
