package datapipeline

import (
	"encoding/json"
	"testing"

	"github.com/trento-project/trento/agent/discovery/mocks"

	"github.com/stretchr/testify/suite"
	_ "github.com/trento-project/trento/test"
	"github.com/trento-project/trento/test/helpers"
	"github.com/trento-project/trento/web/models"
	"gorm.io/gorm"
)

type TelemetryProjectorTestSuite struct {
	suite.Suite
	db *gorm.DB
	tx *gorm.DB
}

func TestTelemetryProjectorTestSuite(t *testing.T) {
	suite.Run(t, new(TelemetryProjectorTestSuite))
}

func (suite *TelemetryProjectorTestSuite) SetupSuite() {
	suite.db = helpers.SetupTestDatabase(suite.T())

	suite.db.AutoMigrate(&Subscription{}, &models.Telemetry{})
}

func (suite *TelemetryProjectorTestSuite) TearDownSuite() {
	suite.db.Migrator().DropTable(Subscription{}, models.Telemetry{})
}

func (suite *TelemetryProjectorTestSuite) SetupTest() {
	suite.tx = suite.db.Begin()
}

func (suite *TelemetryProjectorTestSuite) TearDownTest() {
	suite.tx.Rollback()
}

// Test_HostDiscoveryHandler tests the HostDiscoveryHandler function execution on a HostDiscovery published by an agent
func (s *TelemetryProjectorTestSuite) Test_HostDiscoveryHandler() {
	discoveredHostMock := mocks.NewDiscoveredHostMock()

	requestBody, _ := json.Marshal(discoveredHostMock)

	telemetryProjector_HostDiscoveryHandler(&DataCollectedEvent{
		ID:            1,
		AgentID:       "agent_id",
		DiscoveryType: HostDiscovery,
		Payload:       requestBody,
	}, s.tx)

	var projectedTelemetry models.Telemetry
	s.tx.First(&projectedTelemetry)

	s.Equal(discoveredHostMock.HostName, projectedTelemetry.HostName)
	s.Equal(discoveredHostMock.CPUCount, projectedTelemetry.CPUCount)
	s.Equal(discoveredHostMock.SocketCount, projectedTelemetry.SocketCount)
	s.Equal(discoveredHostMock.TotalMemoryMB, projectedTelemetry.TotalMemoryMB)
	s.Equal("", projectedTelemetry.SLESVersion)
	s.Equal("", projectedTelemetry.CloudProvider)
}

// Test_CloudDiscoveryHandler tests the loudDiscoveryHandler function execution on a CloudDiscovery published by an agent
func (s *TelemetryProjectorTestSuite) Test_CloudDiscoveryHandler() {
	discoveredCloudMock := mocks.NewDiscoveredCloudMock()

	requestBody, _ := json.Marshal(discoveredCloudMock)

	telemetryProjector_CloudDiscoveryHandler(&DataCollectedEvent{
		ID:            1,
		AgentID:       "agent_id",
		DiscoveryType: CloudDiscovery,
		Payload:       requestBody,
	}, s.tx)

	var projectedTelemetry models.Telemetry
	s.tx.First(&projectedTelemetry)

	s.Equal("", projectedTelemetry.HostName)
	s.Equal(0, projectedTelemetry.CPUCount)
	s.Equal(0, projectedTelemetry.SocketCount)
	s.Equal(0, projectedTelemetry.TotalMemoryMB)
	s.Equal("", projectedTelemetry.SLESVersion)

	expectedCloudProvider := discoveredCloudMock.Provider
	s.Equal(expectedCloudProvider, projectedTelemetry.CloudProvider)
}

// Test_SubscriptionDiscoveryHandler tests the SubscriptionDiscoveryHandler function execution on a SubscriptionDiscovery published by an agent
func (s *TelemetryProjectorTestSuite) Test_SubscriptionDiscoveryHandler() {
	discoveredsubscriptionsMock := mocks.NewDiscoveredSubscriptionsMock()

	requestBody, _ := json.Marshal(discoveredsubscriptionsMock)

	telemetryProjector_SubscriptionDiscoveryHandler(&DataCollectedEvent{
		ID:            1,
		AgentID:       "agent_id",
		DiscoveryType: SubscriptionDiscovery,
		Payload:       requestBody,
	}, s.tx)

	var projectedTelemetry models.Telemetry
	s.tx.First(&projectedTelemetry)

	s.Equal("", projectedTelemetry.HostName)
	s.Equal(0, projectedTelemetry.CPUCount)
	s.Equal(0, projectedTelemetry.SocketCount)
	s.Equal(0, projectedTelemetry.TotalMemoryMB)
	s.Equal("", projectedTelemetry.CloudProvider)

	expectedsubscriptionVersion := discoveredsubscriptionsMock[0].Version
	s.Equal(expectedsubscriptionVersion, projectedTelemetry.SLESVersion)
}

// Test_TelemetryProjector tests the TelemetryProjector projects all of the discoveries it is interested in, resulting in a single telemetry readmodel
func (s *TelemetryProjectorTestSuite) Test_TelemetryProjector() {
	telemetryProjector := NewTelemetryProjector(s.tx)

	discoveredCloudMock := mocks.NewDiscoveredCloudMock()
	discoveredHostMock := mocks.NewDiscoveredHostMock()
	discoveredsubscriptionsMock := mocks.NewDiscoveredSubscriptionsMock()

	agentDiscoveries := make(map[string]interface{})
	agentDiscoveries[CloudDiscovery] = discoveredCloudMock
	agentDiscoveries[HostDiscovery] = discoveredHostMock
	agentDiscoveries[SubscriptionDiscovery] = discoveredsubscriptionsMock

	evtID := int64(1)

	for discoveryType, discoveredData := range agentDiscoveries {
		requestBody, _ := json.Marshal(discoveredData)

		telemetryProjector.Project(&DataCollectedEvent{
			ID:            evtID,
			AgentID:       "agent_id",
			DiscoveryType: discoveryType,
			Payload:       requestBody,
		})
		evtID++
	}

	var projectedTelemetry models.Telemetry
	s.tx.First(&projectedTelemetry)

	s.Equal(discoveredHostMock.HostName, projectedTelemetry.HostName)
	s.Equal(discoveredHostMock.CPUCount, projectedTelemetry.CPUCount)
	s.Equal(discoveredHostMock.SocketCount, projectedTelemetry.SocketCount)
	s.Equal(discoveredHostMock.TotalMemoryMB, projectedTelemetry.TotalMemoryMB)

	expectedsubscriptionVersion := discoveredsubscriptionsMock[0].Version
	s.Equal(expectedsubscriptionVersion, projectedTelemetry.SLESVersion)

	expectedCloudProvider := discoveredCloudMock.Provider
	s.Equal(expectedCloudProvider, projectedTelemetry.CloudProvider)
}
