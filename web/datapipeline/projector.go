package datapipeline

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProjectorHandler func(dataCollectedEvent *DataCollectedEvent, db *gorm.DB) error

type Projector struct {
	ID       string
	db       *gorm.DB
	handlers map[string]ProjectorHandler
}

func NewProjector(ID string, db *gorm.DB) *Projector {
	return &Projector{
		ID:       ID,
		db:       db,
		handlers: make(map[string]ProjectorHandler),
	}
}

func (p *Projector) AddHandler(discoveryType string, handler ProjectorHandler) {
	p.handlers[discoveryType] = handler
}

func (p *Projector) Project(dataCollectedEvent *DataCollectedEvent) error {
	handler, ok := p.handlers[dataCollectedEvent.DiscoveryType]

	if !ok {
		log.Infof("Projector: %s is not interested in %s", p.ID, dataCollectedEvent.DiscoveryType)
		return nil
	}

	return p.db.Transaction(func(tx *gorm.DB) error {
		tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&Subscription{
			ProjectorID:                  p.ID,
			AgentID:                      dataCollectedEvent.AgentID,
			LastSeenDataCollectedEventID: dataCollectedEvent.ID,
		})

		err := handler(dataCollectedEvent, tx)
		if err != nil {
			return err
		}

		return nil
	})

}
