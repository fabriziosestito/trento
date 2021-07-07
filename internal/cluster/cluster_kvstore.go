package cluster

import (
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/trento-project/trento/internal/consul"
)

func (c *Cluster) getKVPath() string {
	return path.Join(consul.KvClustersPath, c.Id)
}

func (c *Cluster) Store(client consul.Client) error {

	if !c.IsDc() {
		return nil
	}

	l, err := client.AcquireLockKey(consul.KvClustersPath)
	if err != nil {
		return errors.Wrap(err, "could not lock the kv for clusters")
	}
	defer l.Unlock()

	kvPath := c.getKVPath()
	// Clean the current data before storing the new values
	_, err = client.KV().DeleteTree(kvPath, nil)
	if err != nil {
		return errors.Wrap(err, "Error deleting cluster content")
	}

	clusterMap := make(map[string]interface{})
	mapstructure.Decode(c, &clusterMap)

	err = client.KV().PutMap(kvPath, clusterMap)
	if err != nil {
		return errors.Wrap(err, "Error storing cluster content")
	}

	return nil
}

func Load(client consul.Client) (map[string]*Cluster, error) {
	var clusters = map[string]*Cluster{}

	err := client.WaitLock(consul.KvClustersPath)
	if err != nil {
		return nil, errors.Wrap(err, "error waiting for the lock for clusters")
	}

	entries, err := client.KV().ListMap(consul.KvClustersPath, consul.KvClustersPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not query Consul for cluster KV values")
	}

	for entry, value := range entries {
		cluster := &Cluster{}
		err = mapstructure.Decode(value, &cluster)

		if err != nil {
			return nil, errors.Wrap(err, "could not decode cluster mapstructure")
		}

		clusters[entry] = cluster
	}

	return clusters, nil
}

func Filter(clusters map[string]*Cluster, name string) map[string]*Cluster {
	filteredClusters := make(map[string]*Cluster)
	for clusterId, cluster := range clusters {
		nameFilter := name == "" || strings.Contains(cluster.Name, name)
		// healthFilter := len(health) == 0 || internal.Contains(health, populatedHost.Health())

		if nameFilter {
			filteredClusters[clusterId] = cluster
		}
	}
	return filteredClusters
}
