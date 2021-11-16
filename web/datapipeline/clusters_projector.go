package datapipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/trento/internal/cluster"
	"github.com/trento-project/trento/internal/cluster/cib"
	"github.com/trento-project/trento/web/entities"
	"github.com/trento-project/trento/web/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewClustersProjector(db *gorm.DB) *projector {
	clusterProjector := NewProjector("clusters", db)
	clusterProjector.AddHandler(ClusterDiscovery, clustersProjector_ClusterDiscoveryHandler)

	return clusterProjector
}

// TODO: this is a temporary solution, this code needs to be abstracted in the projector.Project() method
func clustersProjector_ClusterDiscoveryHandler(event *DataCollectedEvent, db *gorm.DB) error {
	data, _ := event.Payload.MarshalJSON()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var cluster cluster.Cluster
	if err := dec.Decode(&cluster); err != nil {
		log.Errorf("can't decode data: %s", err)
		return err
	}

	clusterListReadModel, err := transformClusterData(&cluster)
	if err != nil {
		log.Errorf("can't transform data: %s", err)
		return err
	}

	return db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(clusterListReadModel).Error
}

// transformClusterData transforms the cluster data into the read model
func transformClusterData(cluster *cluster.Cluster) (*entities.Cluster, error) {
	if cluster.Id == "" {
		return nil, fmt.Errorf("no cluster ID found")
	}

	clusterDetail, _ := parseClusterDetail(cluster)
	fmt.Println(clusterDetail)
	return &entities.Cluster{
		ID:          cluster.Id,
		Name:        cluster.Name,
		ClusterType: detectClusterType(cluster),
		// TODO: Cost-optimized has multiple SIDs, we will need to implement this in the future
		SIDs:            getClusterSIDs(cluster),
		ResourcesNumber: cluster.Crmmon.Summary.Resources.Number,
		HostsNumber:     cluster.Crmmon.Summary.Nodes.Number,
		Detail:          (datatypes.JSON)(clusterDetail),
	}, nil
}

// detectClusterType returns the cluster type based on the cluster resources
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

// getClusterSIDs returns the SIDs of the cluster
// TODO: HANA scale-out has multiple SIDs, we will need to implement this in the future
func getClusterSIDs(c *cluster.Cluster) []string {
	var sids []string
	for _, r := range c.Cib.Configuration.Resources.Clones {
		if r.Primitive.Type == "SAPHanaTopology" {
			for _, a := range r.Primitive.InstanceAttributes {
				if a.Name == "SID" && a.Value != "" {
					sids = append(sids, a.Value)
				}
			}
		}
	}

	return sids
}

func parseClusterDetail(c *cluster.Cluster) (json.RawMessage, error) {
	switch detectClusterType(c) {
	case models.ClusterTypeScaleUp:
		fallthrough
	case models.ClusterTypeScaleOut:
		return parseClusterDetailHANA(c)
	default:
		return json.RawMessage{}, nil
	}
}

func parseClusterDetailHANA(c *cluster.Cluster) (json.RawMessage, error) {
	nodes := parseClusterNodes(c)

	var sid string
	var systemReplicationMode, systemReplicationOperationMode, secondarySyncState, srHealthState string

	sids := getClusterSIDs(c)
	if len(sids) > 0 {
		sid = sids[0]
	}

	if len(nodes) > 0 {
		systemReplicationMode, _ = getHANAAttribute(nodes[0], "srmode", sid)
		systemReplicationOperationMode, _ = getHANAAttribute(nodes[0], "op_mode", sid)
		secondarySyncState = getHANASecondarySyncState(nodes, sid)
		srHealthState = getHANAHealthState(nodes[0], sid)
	}

	dateLayout := "Mon Jan 2 15:04:05 2006"
	cibLastWritten, _ := time.Parse(dateLayout, c.Crmmon.Summary.LastChange.Time)

	clusterDetail := &entities.ClusterDetailHANA{
		SystemReplicationMode:          systemReplicationMode,
		SecondarySyncState:             secondarySyncState,
		SystemReplicationOperationMode: systemReplicationOperationMode,
		SRHealthState:                  srHealthState,
		CIBLastWritten:                 cibLastWritten,
		StonithType:                    getClusterFencingType(c),
		StoppedResources:               getClusterStoppedResources(c),
		Nodes:                          nodes,
	}
	fmt.Println(clusterDetail)
	return json.Marshal(clusterDetail)
}

func parseClusterNodes(c *cluster.Cluster) []*entities.ClusterNode {
	var nodes []*entities.ClusterNode

	// TODO: remove plain resources grouping as in the future we'll need to distinguish between Cloned and Groups
	resources := c.Crmmon.Resources
	for _, g := range c.Crmmon.Groups {
		resources = append(resources, g.Resources...)
	}

	for _, c := range c.Crmmon.Clones {
		resources = append(resources, c.Resources...)
	}

	for _, n := range c.Crmmon.NodeAttributes.Nodes {
		node := &entities.ClusterNode{
			Name:       n.Name,
			Attributes: make(map[string]string),
		}

		for _, a := range n.Attributes {
			node.Attributes[a.Name] = a.Value
		}

		for _, r := range resources {
			if r.Node.Name == n.Name {
				resource := &entities.ClusterNodeResource{
					Id:   r.Id,
					Type: r.Agent,
					Role: r.Role,
				}

				switch {
				case r.Active:
					resource.Status = "active"
				case r.Blocked:
					resource.Status = "blocked"
				case r.Failed:
					resource.Status = "failed"
				case r.FailureIgnored:
					resource.Status = "failure_ignored"
				case r.Orphaned:
					resource.Status = "orphaned"
				}

				var primitives []cib.Primitive
				primitives = append(primitives, c.Cib.Configuration.Resources.Primitives...)

				for _, g := range c.Cib.Configuration.Resources.Groups {
					primitives = append(primitives, g.Primitives...)
				}

				if r.Agent == "ocf::heartbeat:IPaddr2" {
					for _, p := range primitives {
						if r.Id == p.Id {
							if len(p.InstanceAttributes) > 0 {
								node.VirtualIPs = append(node.VirtualIPs, p.InstanceAttributes[0].Value)
								break
							}
						}
					}
				}

				for _, nh := range c.Crmmon.NodeHistory.Nodes {
					if nh.Name == n.Name {
						for _, rh := range nh.ResourceHistory {
							if rh.Name == resource.Id {
								resource.FailCount = rh.FailCount
								break
							}
						}
					}
				}

				node.Resources = append(node.Resources, resource)
			}
		}
		nodes = append(nodes, node)
	}

	return nodes
}

func getHANAAttribute(node *entities.ClusterNode, attributeName string, sid string) (string, bool) {
	hanaAttributeName := fmt.Sprintf("hana_%s_%s", strings.ToLower(sid), attributeName)
	value, ok := node.Attributes[hanaAttributeName]

	return value, ok
}

func getHANASecondarySyncState(nodes []*entities.ClusterNode, sid string) string {
	for _, n := range nodes {
		if getHANAStatus(n, sid) == "Secondary" {
			if s, ok := getHANAAttribute(n, "sync_state", sid); ok {
				return s
			}
		}
	}
	return ""
}

// HANAStatus parses the hana_<SID>_roles string and returns the SAPHanaSR Health state
// Possible values: Primary, Secondary
// e.g. 4:P:master1:master:worker:master returns Primary (second element)
func getHANAStatus(node *entities.ClusterNode, sid string) string {
	if r, ok := getHANAAttribute(node, "roles", sid); ok {
		status := strings.SplitN(r, ":", 3)[1]

		switch status {
		case "P":
			return "Primary"
		case "S":
			return "Secondary"
		}
	}
	return ""
}

func getHANAHealthState(node *entities.ClusterNode, sid string) string {
	if r, ok := getHANAAttribute(node, "roles", sid); ok {
		healthState := strings.SplitN(r, ":", 2)[0]
		return healthState
	}
	return ""
}

func getClusterFencingType(c *cluster.Cluster) string {
	const stonithAgent string = "stonith:"
	const stonithResourceMissing string = "notconfigured"

	for _, resource := range c.Crmmon.Resources {
		if strings.HasPrefix(resource.Agent, stonithAgent) {
			return strings.Split(resource.Agent, ":")[1]
		}
	}
	return stonithResourceMissing
}

func getClusterStoppedResources(c *cluster.Cluster) []*entities.ClusterNodeResource {
	var stoppedResources []*entities.ClusterNodeResource

	for _, r := range c.Crmmon.Resources {
		if r.NodesRunningOn == 0 && !r.Active {
			resource := &entities.ClusterNodeResource{
				Id: r.Id,
			}
			stoppedResources = append(stoppedResources, resource)
		}
	}

	return stoppedResources
}
