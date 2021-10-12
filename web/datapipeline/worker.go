package datapipeline

import (
	"gorm.io/gorm"
)

// TODO: tune bufferSize and workersNumber
var workersNumber int = 10
var bufferSize int = workersNumber * 1000

func initProjectorRegistry(db *gorm.DB) []*Projector {
	clusterListProjector := NewProjector("cluster_list", db)
	clusterListProjector.AddHandler(ClusterDiscovery, ClusterListHandler)

	return []*Projector{
		clusterListProjector,
	}
}

func StartProjectorsWorkerPool(db *gorm.DB) chan *DataCollectedEvent {
	ch := make(chan *DataCollectedEvent, bufferSize)
	projectorRegistry := initProjectorRegistry(db)

	for i := 0; i < workersNumber; i++ {
		go Worker(ch, projectorRegistry)
	}

	return ch
}

func Worker(ch chan *DataCollectedEvent, projectors []*Projector) {
	for event := range ch {
		for _, projector := range projectors {
			projector.Project(event)
		}
	}
}
