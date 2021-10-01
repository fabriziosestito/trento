package services

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/trento/web/models"

	"gorm.io/gorm"
)

type TagsServiceTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func TestTagsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TagsServiceTestSuite))
}

func (suite *TagsServiceTestSuite) SetupSuite() {
	suite.db = setupTestDatabase()
	db := suite.db.Exec("TRUNCATE TABLE tags")

	if db.Error != nil {
		panic(db.Error)
	}

	loadTagsFixtures(suite.db)
}

func (suite *TagsServiceTestSuite) TearDownSuite() {
	suite.db.Exec("TRUNCATE TABLE tags")
}

func (suite *TagsServiceTestSuite) TestTagsService_GetAll() {
	tagsService := NewTagsService(suite.db)
	tags, _ := tagsService.GetAll()

	suite.ElementsMatch([]string{"tag1", "tag2", "tag3"}, tags)
}

func TestTagsService_GetAllByResource(t *testing.T) {
}

func (suite *TagsServiceTestSuite) TestTagsService_Create() {
	tx := suite.db.Begin()

	defer func() {
		tx.Rollback()
	}()

	tagsService := NewTagsService(tx)

	tagsService.Create("bananas", "pajamas", "123")
}

func (suite *TagsServiceTestSuite) TestTagsService_Delete() {
	tx := suite.db.Begin()

	defer func() {
		tx.Rollback()
	}()

	tagsService := NewTagsService(tx)

	tagsService.Delete("tag1", models.TagSAPSystemResourceType, "HA1")

	tags, _ := tagsService.GetAll()

	suite.Equal(2, len(tags))
}

// func setupTagsServiceTestDatabase(db *gorm.DB) {
// 	db.AutoMigrate(models.Tag{})
// }

func loadTagsFixtures(db *gorm.DB) {
	db.Create(&models.Tag{
		ResourceType: models.TagSAPSystemResourceType,
		ResourceId:   "HA1",
		Value:        "tag1",
	})
	db.Create(&models.Tag{
		ResourceType: models.TagClusterResourceType,
		ResourceId:   "cluster_id",
		Value:        "tag2",
	})
	db.Create(&models.Tag{
		ResourceType: models.TagHostResourceType,
		ResourceId:   "suse",
		Value:        "tag3",
	})
}
