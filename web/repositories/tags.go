package repositories

import (
	"github.com/trento-project/trento/web/models"
	"gorm.io/gorm"
)

type TagsRepository interface {
	GetAll(resourceTypeFilter ...string) ([]string, error)
	GetAllByResource(resourceType string, resourceId string) ([]string, error)
	Create(value string, resourceType string, resourceId string) error
	Delete(value string, resourceType string, resourceId string) error
}

type tagsRepository struct {
	db *gorm.DB
}

func NewTagsRepository(db *gorm.DB) *tagsRepository {
	return &tagsRepository{db: db}
}

func (r *tagsRepository) GetAll(resourceTypeFilter ...string) ([]string, error) {
	for _, f := range resourceTypeFilter {
		r.db.Where("resource_type", f)
	}

	return getTags(r.db)
}

func (r *tagsRepository) GetAllByResource(resourceType string, resourceId string) ([]string, error) {
	r.db.Where("resource_type", resourceType)
	r.db.Where("resource_id", resourceId)

	return getTags(r.db)
}

func (r *tagsRepository) Create(value string, resourceType string, resourceId string) error {
	tag := models.Tag{
		Value:        value,
		ResourceType: resourceType,
		ResourceId:   resourceId,
	}

	result := r.db.Create(&tag)

	return result.Error
}

func (r *tagsRepository) Delete(value string, resourceType string, resourceId string) error {
	tag := models.Tag{
		Value:        value,
		ResourceType: resourceType,
		ResourceId:   resourceId,
	}

	result := r.db.Delete(&tag)

	return result.Error
}

func getTags(db *gorm.DB) ([]string, error) {
	var tags []models.Tag
	result := db.Find(&tags)

	if result.Error != nil {
		return nil, result.Error
	}

	var tagStrings []string
	for _, t := range tags {
		tagStrings = append(tagStrings, t.Value)
	}

	return tagStrings, nil
}
