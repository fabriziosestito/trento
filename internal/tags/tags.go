package tags

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/trento-project/trento/internal/consul"
)

const KvTagsPath string = "trento/v0/tags/%s/%s/"

type Tags struct {
	ResourceType string              `mapstructure:"resource"`
	ID           string              `mapstructure:"id"`
	Values       map[string]struct{} `mapstructure:"values"`
}

// type Tag struct {
// 	Value string `mapstructure:"value"`
// }

func getKvTagsPath(resource string, id string) string {
	return fmt.Sprintf(KvTagsPath, resource, id)
}

func Load(resource string, id string, client consul.Client) (*Tags, error) {
	path := getKvTagsPath(resource, id)
	err := client.WaitLock(path)

	if err != nil {
		return nil, errors.Wrap(err, "error waiting for the lock for tags")
	}

	entries, err := client.KV().ListMap(path, path)
	if err != nil {
		return nil, errors.Wrap(err, "could not query Consul for tags KV values")
	}
	tags := &Tags{}
	mapstructure.Decode(entries, &tags)

	return tags, nil
}

func (t *Tags) Store(client consul.Client) error {
	kvPath := getKvTagsPath(t.ResourceType, t.ID)

	tagsMap := make(map[string]interface{})
	mapstructure.Decode(t, &tagsMap)

	err := client.KV().PutMap(kvPath, tagsMap)
	if err != nil {
		return errors.Wrap(err, "Error storing a host tags")
	}

	return nil
}

func (t *Tags) Delete(value string, client consul.Client) error {
	delete(t.Values, value)
	kvPath := getKvTagsPath(t.ResourceType, t.ID)

	tagPath := kvPath + "values/" + value + "/"
	fmt.Println(tagPath)

	_, err := client.KV().DeleteTree(tagPath, nil)
	if err != nil {
		return err
	}
	return nil
}
