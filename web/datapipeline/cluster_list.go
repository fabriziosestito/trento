package datapipeline

import (
	"bytes"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/trento/internal/cluster"
	"github.com/trento-project/trento/web/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ClusterListHandler(event *DataCollectedEvent, db *gorm.DB) error {
	data, _ := event.Payload.MarshalJSON()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var cluster cluster.Cluster
	if err := dec.Decode(&cluster); err != nil {
		log.Errorf("can't decode data: %s", err)
		return err
	}

	clusterListReadModel, err := transformClusterListData(&cluster)
	if err != nil {
		log.Errorf("can't transform data: %s", err)
		return err
	}

	return db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(clusterListReadModel).Error
}

func transformClusterListData(cluster *cluster.Cluster) (*models.Cluster, error) {
	return &models.Cluster{
		ID:          cluster.Id,
		Name:        cluster.Name,
		ClusterType: detectClusterType(cluster),
		// TODO: Cost-optimized has multiple SIDs, we will need to implement this in the future
		SIDs:            []string{getHanaSID(cluster)},
		ResourcesNumber: cluster.Crmmon.Summary.Resources.Number,
		HostsNumber:     cluster.Crmmon.Summary.Nodes.Number,
	}, nil
}

func detectClusterType(cluster *cluster.Cluster) string {
	var hasSapHanaTopology, hasSAPHanaController, hasSAPHana bool

	for _, c := range cluster.Crmmon.Clones {
		for _, r := range c.Resources {
			switch r.Agent {
			case "ocf::suse:SAPHanaTopology":
				hasSapHanaTopology = true
			case "ocf::suse:SAPHana":
				hasSAPHana = true
			case "ocf::suse:SAPHanaController":
				hasSAPHanaController = true
			}
		}
	}

	switch {
	case hasSapHanaTopology && hasSAPHana:
		return models.ClusterTypeScaleUp
	case hasSapHanaTopology && hasSAPHanaController:
		return models.ClusterTypeScaleOut
	default:
		return models.ClusterTypeUnknown
	}
}

func getHanaSID(c *cluster.Cluster) string {
	for _, r := range c.Cib.Configuration.Resources.Clones {
		if r.Primitive.Type == "SAPHanaTopology" {
			for _, a := range r.Primitive.InstanceAttributes {
				if a.Name == "SID" {
					return a.Value
				}
			}
		}
	}

	return ""
}
