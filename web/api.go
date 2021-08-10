package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/internal/tags"
)

func ApiPingHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func getTags(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := tags.NewTags(client, "banana", "yeah")
		ttt, _ := t.GetAll()

		c.JSON(http.StatusOK, ttt)
	}
}

func deleteTags(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := tags.NewTags(client, "banana", "yeah")
		t.Delete("moar-bananas")

		c.JSON(http.StatusOK, "pong")
	}
}
