package datapipeline

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func StartProjectorsWorkerPool(workersNumber int, db *gorm.DB) chan *DataCollectedEvent {
	// find a way to make the projectons non blocking
	ch := make(chan *DataCollectedEvent, workersNumber*1000)

	for i := 0; i < workersNumber; i++ {
		go Worker(ch, db)
	}

	return ch
}

func Worker(ch chan *DataCollectedEvent, db *gorm.DB) {
	for event := range ch {
		switch event.DiscoveryType {
		case ClusterDiscovery:
			Project(event, db, ClusterListHandler)
		default:
			log.Errorf("unknown discovery type: %s", event.DiscoveryType)
		}
		fmt.Println("Received event:", event.DiscoveryType)
	}
}

func Project(event *DataCollectedEvent, db *gorm.DB, handler func(*DataCollectedEvent, *gorm.DB) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&Subscription{
			DiscoveryType:  event.DiscoveryType,
			AgentID:        event.AgentID,
			LastSeenDataID: event.ID,
		})

		err := handler(event, tx)
		if err != nil {
			return err
		}

		return nil
	})
}
