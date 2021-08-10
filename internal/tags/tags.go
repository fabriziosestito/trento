package tags

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/trento-project/trento/internal/consul"
)

const kvTagsPath string = "trento/v0/tags/%s/%s/"

type Tags struct {
	client   consul.Client
	resource string
	id       string
}

func NewTags(client consul.Client, resource string, id string) *Tags {
	return &Tags{
		client:   client,
		resource: resource,
		id:       id,
	}
}

func (t *Tags) getKvTagsPath() string {
	return fmt.Sprintf(kvTagsPath, t.resource, t.id)
}

func (t *Tags) GetAll() ([]string, error) {
	path := t.getKvTagsPath()

	tagsMap, err := t.client.KV().ListMap(path, path)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving tags")
	}

	var tags []string
	for tag := range tagsMap {
		tags = append(tags, tag)
	}

	return tags, nil
}

func (t *Tags) Create(tag string) error {
	path := t.getKvTagsPath() + tag + "/"

	if err := t.client.KV().PutMap(path, nil); err != nil {
		return errors.Wrap(err, "Error storing a host tags")
	}

	return nil
}

func (t *Tags) Delete(tag string) error {
	path := t.getKvTagsPath() + tag + "/"

	_, err := t.client.KV().DeleteTree(path, nil)
	if err != nil {
		return err
	}
	return nil
}
