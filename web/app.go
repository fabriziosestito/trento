package web

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/internal/tags"
)

//go:embed frontend/assets
var assetsFS embed.FS

//go:embed templates
var templatesFS embed.FS

type App struct {
	host string
	port int
	Dependencies
}

type Dependencies struct {
	consul consul.Client
	engine *gin.Engine
}

func DefaultDependencies() Dependencies {
	consulClient, _ := consul.DefaultClient()
	engine := gin.Default()

	return Dependencies{consulClient, engine}
}

// shortcut to use default dependencies
func NewApp(host string, port int) (*App, error) {
	return NewAppWithDeps(host, port, DefaultDependencies())
}

func NewAppWithDeps(host string, port int, deps Dependencies) (*App, error) {
	app := &App{
		Dependencies: deps,
		host:         host,
		port:         port,
	}

	engine := deps.engine
	engine.HTMLRender = NewLayoutRender(templatesFS, "templates/*.tmpl")
	engine.Use(ErrorHandler)
	engine.StaticFS("/static", http.FS(assetsFS))
	engine.GET("/", HomeHandler)
	engine.GET("/hosts", NewHostListHandler(deps.consul))
	engine.GET("/hosts/:name", NewHostHandler(deps.consul))
	engine.GET("/hosts/:name/ha-checks", NewHAChecksHandler(deps.consul))
	engine.GET("/clusters", NewClusterListHandler(deps.consul))
	engine.GET("/clusters/:id", NewClusterHandler(deps.consul))
	engine.GET("/sapsystems", NewSAPSystemListHandler(deps.consul))
	engine.GET("/sapsystems/:sid", NewSAPSystemHandler(deps.consul))

	engine.GET("/write-tags/", writeTags(deps.consul))
	engine.GET("/get-tags/", getTags(deps.consul))
	engine.GET("/delete-tag/", deleteTags(deps.consul))

	apiGroup := engine.Group("/api")
	{
		apiGroup.GET("/ping", ApiPingHandler)
	}

	return app, nil
}

func writeTags(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tags := &tags.Tags{
			ResourceType: "banana",
			ID:           "yeah",
			Values: map[string]struct{}{
				"tag1": struct{}{},
				"tag2": struct{}{},
			},
		}

		tags.Store(client)
	}
}

func getTags(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, _ := tags.Load("banana", "yeah", client)

		for k, _ := range tags.Values {
			fmt.Println(k)
		}
	}
}

func deleteTags(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, _ := tags.Load("banana", "yeah", client)
		tags.Delete("tag2", client)

		fmt.Println(tags)
	}
}

func (a *App) Start() error {
	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", a.host, a.port),
		Handler:        a,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s.ListenAndServe()
}

func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.engine.ServeHTTP(w, req)
}
