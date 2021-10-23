package datapipeline

import (
	"bytes"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/trento/internal/cloud"
	"github.com/trento-project/trento/web/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TelemetryCloudHandler(event *DataCollectedEvent, db *gorm.DB) error {
	data, _ := event.Payload.MarshalJSON()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var cloud cloud.CloudInstance
	if err := dec.Decode(&cloud); err != nil {
		log.Errorf("can't decode data: %s", err)
		return err
	}

	return db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&models.Telemetry{
		AgentID:       event.AgentID,
		CloudProvider: cloud.Provider,
	}).Error
}
