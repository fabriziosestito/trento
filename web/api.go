package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trento-project/trento/internal/cluster"
	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/internal/tags"
)

func ApiPingHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

type JSONTag struct {
	Tag string `json:"tag" binding:"required"`
}

// ApiHostCreateTagHandler godoc
// @Summary Add tag to host
// @Accept json
// @Produce json
// @Param name path string true "Host name"
// @Param Body body JSONTag true "The tag to create"
// @Success 201 {object} JSONTag
// @Failure 404 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/hosts/{name}/tags [post]
func ApiHostCreateTagHandler(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		catalogNode, _, err := client.Catalog().Node(name, nil)
		if err != nil {
			_ = c.Error(err)
			return
		}

		if catalogNode == nil {
			_ = c.Error(NotFoundError("could not find host"))
			return
		}

		var r JSONTag

		err = c.BindJSON(&r)
		if err != nil {
			_ = c.Error(UnprocessableEntityError("unable to parse JSON body"))
			return
		}

		t := tags.NewTags(client, "hosts", name)
		t.Create(r.Tag)

		c.JSON(http.StatusCreated, &r)
	}
}

// ApiHostListTagHandler godoc
// @Summary Get all tags that belong to a host
// @Accept json
// @Produce json
// @Param name path string true "Host name"
// @Success 200 {object} []JSONTag
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/hosts/{name}/tags [get]
func ApiHostListTagHandler(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		catalogNode, _, err := client.Catalog().Node(name, nil)
		if err != nil {
			_ = c.Error(err)
			return
		}

		if catalogNode == nil {
			_ = c.Error(NotFoundError("could not find host"))
			return
		}

		t := tags.NewTags(client, "hosts", name)
		tagsList, err := t.GetAll()
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, tagsList)
	}
}

// ApiHostDeleteTagHandler godoc
// @Summary Delete a specific tag that belongs to a host
// @Accept json
// @Produce json
// @Param name path string true "Host name"
// @Param tag path string true "Tag"
// @Success 204 {object} map[string]interface{}
// @Router /api/hosts/{name}/tags/{tag} [delete]
func ApiHostDeleteTagHandler(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		tag := c.Param("tag")

		catalogNode, _, err := client.Catalog().Node(name, nil)
		if err != nil {
			_ = c.Error(err)
			return
		}

		if catalogNode == nil {
			_ = c.Error(NotFoundError("could not find host"))
			return
		}

		t := tags.NewTags(client, "hosts", name)
		err = t.Delete(tag)

		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}

// ApiClusterCreateTagHandler godoc
// @Summary Add tag to Cluster
// @Accept json
// @Produce json
// @Param id path string true "Cluster id"
// @Param Body body JSONTag true "The tag to create"
// @Success 201 {object} JSONTag
// @Failure 404 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/clusters/{id}/tags [post]
func ApiClusterCreateTagHandler(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		clusters, err := cluster.Load(client)
		if err != nil {
			_ = c.Error(err)
			return
		}

		if _, ok := clusters[id]; !ok {
			_ = c.Error(NotFoundError("could not find cluster"))
			return
		}

		var r JSONTag

		err = c.BindJSON(&r)
		if err != nil {
			_ = c.Error(UnprocessableEntityError("unable to parse JSON body"))
			return
		}

		t := tags.NewTags(client, "clusters", id)
		t.Create(r.Tag)

		c.JSON(http.StatusCreated, &r)
	}
}

// ApiClusterListTagHandler godoc
// @Summary Get all tags that belong to a cluster
// @Accept json
// @Produce json
// @Param id path string true "Cluster id"
// @Success 200 {object} []JSONTag
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/clusters/{name}/tags [get]
func ApiClusterListTagHandler(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		clusters, err := cluster.Load(client)
		if err != nil {
			_ = c.Error(err)
			return
		}

		if _, ok := clusters[id]; !ok {
			_ = c.Error(NotFoundError("could not find cluster"))
			return
		}

		t := tags.NewTags(client, "clusters", id)
		tagsList, err := t.GetAll()
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, tagsList)
	}
}

// ApiClusterDeleteTagHandler godoc
// @Summary Delete a specific tag that belongs to a cluster
// @Accept json
// @Produce json
// @Param cluster path string true "Cluster id"
// @Param tag path string true "Tag"
// @Success 204 {object} map[string]interface{}
// @Router /api/clusters/{name}/tags/{tag} [delete]
func ApiClusterDeleteTagHandler(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		tag := c.Param("tag")

		clusters, err := cluster.Load(client)
		if err != nil {
			_ = c.Error(err)
			return
		}

		if _, ok := clusters[id]; !ok {
			_ = c.Error(NotFoundError("could not find cluster"))
			return
		}

		t := tags.NewTags(client, "clusters", id)
		err = t.Delete(tag)

		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}
