package datapipeline

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	discoveryModels "github.com/trento-project/trento/agent/discovery/models"
	"github.com/trento-project/trento/internal/cloud"
	"github.com/trento-project/trento/internal/cluster"
	"github.com/trento-project/trento/internal/sapsystem"
	"github.com/trento-project/trento/web/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewHostsProjector(db *gorm.DB) *projector {
	hostsProjector := NewProjector("hosts", db)

	hostsProjector.AddHandler(HostDiscovery, hostsProjector_HostDiscoveryHandler)
	hostsProjector.AddHandler(CloudDiscovery, hostsProjector_CloudDiscoveryHandler)
	hostsProjector.AddHandler(ClusterDiscovery, hostsProjector_ClusterDiscoveryHandler)
	hostsProjector.AddHandler(SAPsystemDiscovery, hostsProjector_SAPSystemDiscoveryHandler)

	return hostsProjector
}

func hostsProjector_HostDiscoveryHandler(dataCollectedEvent *DataCollectedEvent, db *gorm.DB) error {
	decoder := payloadDecoder(dataCollectedEvent.Payload)

	var discoveredHost discoveryModels.DiscoveredHost
	if err := decoder.Decode(&discoveredHost); err != nil {
		log.Errorf("can't decode data: %s", err)
		return err
	}

	hostReadModel := models.Host{
		AgentID:      dataCollectedEvent.AgentID,
		Name:         discoveredHost.HostName,
		IPAddresses:  discoveredHost.HostIpAddresses,
		AgentVersion: discoveredHost.AgentVersion,
	}

	return storeHost(db, hostReadModel,
		"name",
		"ip_addresses",
		"agent_version",
	)
}

func hostsProjector_CloudDiscoveryHandler(dataCollectedEvent *DataCollectedEvent, db *gorm.DB) error {
	decoder := payloadDecoder(dataCollectedEvent.Payload)

	var discoveredCloud cloud.CloudInstance
	if err := decoder.Decode(&discoveredCloud); err != nil {
		log.Errorf("can't decode data: %s", err)
		return err
	}

	hostReadModel := models.Host{
		AgentID:       dataCollectedEvent.AgentID,
		CloudProvider: discoveredCloud.Provider,
	}

	return storeHost(db, hostReadModel, "cloud_provider")
}

func hostsProjector_ClusterDiscoveryHandler(dataCollectedEvent *DataCollectedEvent, db *gorm.DB) error {
	decoder := payloadDecoder(dataCollectedEvent.Payload)

	var discoveredCluster cluster.Cluster
	if err := decoder.Decode(&discoveredCluster); err != nil {
		log.Errorf("can't decode data: %s", err)
		return err
	}

	hostReadModel := models.Host{
		AgentID:     dataCollectedEvent.AgentID,
		ClusterID:   discoveredCluster.Id,
		ClusterName: discoveredCluster.Name,
	}

	return storeHost(db, hostReadModel, "cluster_id", "cluster_name")
}

func hostsProjector_SAPSystemDiscoveryHandler(dataCollectedEvent *DataCollectedEvent, db *gorm.DB) error {
	decoder := payloadDecoder(dataCollectedEvent.Payload)

	var discoveredSAPSystems sapsystem.SAPSystemsList
	if err := decoder.Decode(&discoveredSAPSystems); err != nil {
		log.Errorf("can't decode data: %s", err)
		return err
	}

	var sids []string
	for _, s := range discoveredSAPSystems {
		fmt.Println(s.SID)
		sids = append(sids, s.SID)
	}

	hostReadModel := models.Host{
		AgentID: dataCollectedEvent.AgentID,
		SIDs:    sids,
	}

	return storeHost(db, hostReadModel, "sids")
}

func storeHost(db *gorm.DB, hostReadModel models.Host, updateProperties ...string) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "agent_id"},
		},
		DoUpdates: clause.AssignmentColumns(updateProperties),
	}).Create(&hostReadModel).Error
}
