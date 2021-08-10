package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"

	"github.com/trento-project/trento/internal/cloud"
	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/internal/hosts"
	"github.com/trento-project/trento/internal/sapsystem"
)

func NewHostsHealthContainer(hostList hosts.HostList) *HealthContainer {
	h := &HealthContainer{}
	for _, host := range hostList {
		switch host.Health() {
		case consulApi.HealthPassing:
			h.PassingCount += 1
		case consulApi.HealthWarning:
			h.WarningCount += 1
		case consulApi.HealthCritical:
			h.CriticalCount += 1
		}
	}
	return h
}

func NewHostListHandler(client consul.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Request.URL.Query()
		queryFilter := hosts.CreateFilterMetaQuery(query)
		healthFilter := query["health"]

		hostList, err := hosts.Load(client, queryFilter, healthFilter)
		if err != nil {
			_ = c.Error(err)
			return
		}

		hContainer := NewHostsHealthContainer(hostList)
		hContainer.Layout = "horizontal"

		page := c.DefaultQuery("page", "1")
		perPage := c.DefaultQuery("per_page", "10")
		pagination := NewPaginationWithStrings(len(hostList), page, perPage)
		firstElem, lastElem := pagination.GetSliceNumbers()

		c.HTML(http.StatusOK, "hosts.html.tmpl", gin.H{
			"Hosts":           hostList[firstElem:lastElem],
			"SIDs":            getAllSIDs(hostList),
			"AppliedFilters":  query,
			"HealthContainer": hContainer,
			"Pagination":      pagination,
		})
	}
}

func getAllSIDs(hostList hosts.HostList) []string {
	var sids []string
	set := make(map[string]struct{})

	for _, host := range hostList {
		for _, s := range strings.Split(host.TrentoMeta()["trento-sap-systems"], ",") {
			if s == "" {
				continue
			}

			_, ok := set[s]
			if !ok {
				sids = append(sids, s)
				set[s] = struct{}{}
			}
		}
	}

	return sids
}

func loadHealthChecks(client consul.Client, node string) ([]*consulApi.HealthCheck, error) {

	checks, _, err := client.Health().Node(node, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not query Consul for health checks")
	}

	return checks, nil
}

func NewHostHandler(client consul.Client) gin.HandlerFunc {
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

		checks, err := loadHealthChecks(client, name)
		if err != nil {
			_ = c.Error(err)
			return
		}

		systems, err := sapsystem.Load(client, name)
		if err != nil {
			_ = c.Error(err)
			return
		}

		cloudData, err := cloud.Load(client, name)
		if err != nil {
			_ = c.Error(err)
			return
		}

		host := hosts.NewHost(*catalogNode.Node, client)
		c.HTML(http.StatusOK, "host.html.tmpl", gin.H{
			"Host":         &host,
			"HealthChecks": checks,
			"SAPSystems":   systems,
			"CloudData":    cloudData,
		})
	}
}

func NewHAChecksHandler(client consul.Client) gin.HandlerFunc {
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

		host := hosts.NewHost(*catalogNode.Node, client)
		c.HTML(http.StatusOK, "ha_checks.html.tmpl", gin.H{
			"Hostname": host.Name(),
			"HAChecks": host.HAChecks(),
		})
	}
}
