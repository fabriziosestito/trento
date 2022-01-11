package debug

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	dbCmd "github.com/trento-project/trento/cmd/db"
	"github.com/trento-project/trento/internal/db"
	"github.com/trento-project/trento/web"
	"github.com/trento-project/trento/web/datapipeline"
	"github.com/trento-project/trento/web/entities"
	"gorm.io/gorm"
)

func NewDebugCmd() *cobra.Command {
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Maintenance commands, USE WITH CAUTION.",
		PersistentPreRun: func(debugCmd *cobra.Command, _ []string) {
			debugCmd.Flags().VisitAll(func(f *pflag.Flag) {
				viper.BindPFlag(f.Name, f)
			})
			debugCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
				viper.BindPFlag(f.Name, f)
			})
		},
	}

	dbCmd.AddDBFlags(debugCmd)
	addPruneEventsCmd(debugCmd)
	addPruneChecksResultsCmd(debugCmd)
	addDBResetCmd(debugCmd)
	addDumpScenarioCmd(debugCmd)

	return debugCmd
}

func addPruneEventsCmd(debugCmd *cobra.Command) {
	var olderThan uint

	pruneCmd := &cobra.Command{
		Use:   "prune-events",
		Short: "Prune events older than",
		Run: func(*cobra.Command, []string) {
			db := initDB()
			olderThan := viper.GetUint("older-than")
			olderThanDuration := time.Duration(olderThan) * 24 * time.Hour

			pruneEvents(db, olderThanDuration)
		},
	}

	pruneCmd.Flags().UintVar(&olderThan, "older-than", 10, "Prune data discovery events older than <value> days.")

	debugCmd.AddCommand(pruneCmd)
}

func addPruneChecksResultsCmd(debugCmd *cobra.Command) {
	var olderThan uint

	pruneCmd := &cobra.Command{
		Use:   "prune-checks-results",
		Short: "Prune checks results older than",
		Run: func(*cobra.Command, []string) {
			db := initDB()
			olderThan := viper.GetUint("older-than")
			olderThanDuration := time.Duration(olderThan) * 24 * time.Hour

			pruneChecksResults(db, olderThanDuration)
		},
	}

	pruneCmd.Flags().UintVar(&olderThan, "older-than", 10, "Prune executed checks results data older than <value> days.")

	debugCmd.AddCommand(pruneCmd)
}

func addDBResetCmd(debugCmd *cobra.Command) {
	dbResetCmd := &cobra.Command{
		Use:   "db-reset",
		Short: "Reset the database",
		Run: func(*cobra.Command, []string) {
			db := initDB()

			dbReset(db)
		},
	}

	debugCmd.AddCommand(dbResetCmd)
}

func addDumpScenarioCmd(debugCmd *cobra.Command) {
	dumpScenarioCmd := &cobra.Command{
		Use:   "dump-scenario",
		Short: "Dump the current scenario",
		Run: func(*cobra.Command, []string) {
			db := initDB()
			exportPath := viper.GetString("path")
			scenarioName := viper.GetString("name")

			dumpScenario(db, exportPath, scenarioName)
		},
	}

	var exportPath, scenarioName string

	dumpScenarioCmd.Flags().StringVar(&exportPath, "path", ".", "The path where the scenario will be exported.")
	dumpScenarioCmd.Flags().StringVar(&scenarioName, "name", "", "The scenario name.")
	dumpScenarioCmd.MarkFlagRequired("name")

	debugCmd.AddCommand(dumpScenarioCmd)
}

func initDB() *gorm.DB {
	dbConfig := dbCmd.LoadConfig()
	db, err := db.InitDB(dbConfig)
	if err != nil {
		log.Fatal("Error while initializing the database: ", err)
	}

	return db
}

func pruneEvents(db *gorm.DB, olderThan time.Duration) {
	log.Infof("Pruning events older than %d days.", olderThan)

	result := db.Delete(datapipeline.DataCollectedEvent{}, "created_at < ?", time.Now().Add(-olderThan))
	log.Debugf("Pruned %d events", result.RowsAffected)

	if result.Error != nil {
		log.Fatalf("Error while pruning older events: %s", result.Error)
	}
	log.Infof("Events older than %d days pruned.", olderThan)
}

func pruneChecksResults(db *gorm.DB, olderThan time.Duration) {
	log.Infof("Pruning checks results older than %d days.", olderThan)

	result := db.Delete(entities.ChecksResult{}, "created_at < ?", time.Now().Add(-olderThan))
	log.Debugf("Pruned %d checks results", result.RowsAffected)

	if result.Error != nil {
		log.Fatalf("Error while pruning older events: %s", result.Error)
	}
	log.Infof("Checks results older than %d days pruned.", olderThan)
}

func dbReset(db *gorm.DB) {
	log.Info("Resetting database...")
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, t := range web.DBTables {
			stmt := &gorm.Statement{DB: db}
			stmt.Parse(t)
			tableName := stmt.Schema.Table

			err := tx.Raw("TRUNCATE TABLE ?", tableName).Error
			if err != nil {
				log.Fatalf("Error while truncating table %s: %s", tableName, err)
			}
			log.Infof("Table %s truncated.", tableName)
		}
		return nil
	})

	if err != nil {
		log.Fatal("Error while resetting the database: ", err)
	}

	log.Info("Database reset.")
}

func dumpScenario(db *gorm.DB, exportPath string, scenarioName string) {
	var events []datapipeline.DataCollectedEvent

	err := db.
		Joins("JOIN subscriptions ON subscriptions.last_projected_event_id = data_collected_events.id").
		Find(&events).Error
	if err != nil {
		log.Fatal("Error while exporting scenario from the database: ", err)
	}

	path := filepath.Join(exportPath, scenarioName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0700)
		if err != nil {
			log.Fatal("Error while creating directory: ", err)
		}
	}

	for _, event := range events {
		data, err := json.MarshalIndent(map[string]interface{}{
			"agent_id":       event.AgentID,
			"discovery_type": event.DiscoveryType,
			"payload":        event.Payload,
		}, "", " ")
		if err != nil {
			log.Fatal("Error while marshaling event: ", err)
		}

		filePath := filepath.Join(path, event.AgentID, event.DiscoveryType)
		err = ioutil.WriteFile(filePath, data, 0644)
		if err != nil {
			log.Fatal("Error while writing event: ", err)
		}
	}
}
