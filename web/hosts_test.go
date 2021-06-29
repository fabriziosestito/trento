package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/internal/consul/mocks"
)

func setupHostsTest() *mocks.Client {
	nodes := []*consulApi.Node{
		{
			Node:       "foo",
			Datacenter: "dc1",
			Address:    "192.168.1.1",
			Meta: map[string]string{
				"trento-sap-environment": "env1",
				"trento-sap-landscape":   "land1",
				"trento-sap-system":      "sys1",
			},
		},
		{
			Node:       "bar",
			Datacenter: "dc",
			Address:    "192.168.1.2",
			Meta: map[string]string{
				"trento-sap-environment": "env1",
				"trento-sap-landscape":   "land1",
				"trento-sap-system":      "sys1",
			},
		},
	}

	fooHealthChecks := consulApi.HealthChecks{
		&consulApi.HealthCheck{
			Status: consulApi.HealthPassing,
		},
	}

	barHealthChecks := consulApi.HealthChecks{
		&consulApi.HealthCheck{
			Status: consulApi.HealthCritical,
		},
	}

	filters := map[string]interface{}{
		"env1": map[string]interface{}{
			"name": "env1",
			"landscapes": map[string]interface{}{
				"land1": map[string]interface{}{
					"name": "land1",
					"sapsystems": map[string]interface{}{
						"sys1": map[string]interface{}{
							"name": "sys1",
						},
					},
				},
				"land2": map[string]interface{}{
					"name": "land2",
					"sapsystems": map[string]interface{}{
						"sys2": map[string]interface{}{
							"name": "sys2",
						},
					},
				},
			},
		},
		"env2": map[string]interface{}{
			"name": "env2",
			"landscapes": map[string]interface{}{
				"land3": map[string]interface{}{
					"name": "land3",
					"sapsystems": map[string]interface{}{
						"sys3": map[string]interface{}{
							"name": "sys3",
						},
					},
				},
			},
		},
	}

	consulInst := new(mocks.Client)
	catalog := new(mocks.Catalog)
	health := new(mocks.Health)
	kv := new(mocks.KV)

	consulInst.On("Catalog").Return(catalog)
	consulInst.On("Health").Return(health)
	consulInst.On("KV").Return(kv)

	kv.On("ListMap", consul.KvEnvironmentsPath, consul.KvEnvironmentsPath).Return(filters, nil)

	query := &consulApi.QueryOptions{Filter: ""}
	catalog.On("Nodes", (*consulApi.QueryOptions)(query)).Return(nodes, nil, nil)

	filterSys1 := &consulApi.QueryOptions{
		Filter: "(Meta[\"trento-sap-environment\"] == \"env1\") and (Meta[\"trento-sap-landscape\"] == \"land1\") and (Meta[\"trento-sap-system\"] == \"sys1\")"}
	catalog.On("Nodes", filterSys1).Return(nodes, nil, nil)

	filterSys2 := &consulApi.QueryOptions{
		Filter: "(Meta[\"trento-sap-environment\"] == \"env1\") and (Meta[\"trento-sap-landscape\"] == \"land2\") and (Meta[\"trento-sap-system\"] == \"sys2\")"}
	catalog.On("Nodes", filterSys2).Return(nodes, nil, nil)

	filterSys3 := &consulApi.QueryOptions{
		Filter: "(Meta[\"trento-sap-environment\"] == \"env2\") and (Meta[\"trento-sap-landscape\"] == \"land3\") and (Meta[\"trento-sap-system\"] == \"sys3\")"}
	catalog.On("Nodes", filterSys3).Return(nodes, nil, nil)

	health.On("Node", "foo", (*consulApi.QueryOptions)(nil)).Return(fooHealthChecks, nil, nil)
	health.On("Node", "bar", (*consulApi.QueryOptions)(nil)).Return(barHealthChecks, nil, nil)

	catalog.On("Node", "foo", (*consulApi.QueryOptions)(nil)).Return(&consulApi.CatalogNode{Node: nodes[0]}, nil, nil)
	consulInst.On("WaitLock", "trento/v0/hosts/foo/sapsystems/").Return(nil)
	kv.On("ListMap", "trento/v0/hosts/foo/sapsystems/", "trento/v0/hosts/foo/sapsystems/").Return(filters, nil)

	return consulInst
}

func TestHostsListHandler(t *testing.T) {
	consulInst := setupHostsTest()

	deps := DefaultDependencies()
	deps.consul = consulInst

	app, err := NewAppWithDeps("", 80, deps)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/hosts", nil)
	if err != nil {
		t.Fatal(err)
	}

	app.ServeHTTP(resp, req)

	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.Add("text/html", &html.Minifier{
		KeepDefaultAttrVals: true,
		KeepEndTags:         true,
	})
	minified, err := m.String("text/html", resp.Body.String())
	if err != nil {
		panic(err)
	}
	fmt.Println(minified)
	assert.Equal(t, 200, resp.Code)
	assert.Contains(t, minified, "Hosts")
	assert.Regexp(t, regexp.MustCompile("<select name=trento-sap-environment.*>.*env1.*env2.*</select>"), minified)
	assert.Regexp(t, regexp.MustCompile("<select name=trento-sap-landscape.*>.*land1.*land2.*land3.*</select>"), minified)
	assert.Regexp(t, regexp.MustCompile("<select name=trento-sap-system.*>.*sys1.*sys2.*sys3.*</select>"), minified)
	assert.Regexp(t, regexp.MustCompile("<td>foo</td><td>192.168.1.1</td>.*passing.*"), minified)
	//assert.Regexp(t, regexp.MustCompile("<td>bar</td><td>192.168.1.2</td><td>.*land2.*</td><td>.*critical.*</td>"), minified)

}

func TestHostHandler(t *testing.T) {
	consulInst := setupHostsTest()

	deps := DefaultDependencies()
	deps.consul = consulInst

	app, err := NewAppWithDeps("", 80, deps)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/hosts/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/html")

	app.ServeHTTP(resp, req)

	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.Add("text/html", &html.Minifier{
		KeepDefaultAttrVals: true,
		KeepEndTags:         true,
	})
	minified, err := m.String("text/html", resp.Body.String())
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 200, resp.Code)
	assert.Contains(t, minified, "Host details")
	assert.Regexp(t, regexp.MustCompile("<dd.*>foo</dd>"), minified)
	assert.Regexp(t, regexp.MustCompile("<a.*environments.*>env1</a>"), minified)
	assert.Regexp(t, regexp.MustCompile("<a.*landscapes.*>land1</a>"), minified)
	assert.Regexp(t, regexp.MustCompile("<a.*sapsystems.*>sys1</a>"), minified)
	assert.Regexp(t, regexp.MustCompile("<span.*>passing</span>"), minified)
}

func TestHostHandler404Error(t *testing.T) {
	consulInst := new(mocks.Client)
	catalog := new(mocks.Catalog)
	catalog.On("Node", "foobar", (*consulApi.QueryOptions)(nil)).Return((*consulApi.CatalogNode)(nil), nil, nil)
	consulInst.On("Catalog").Return(catalog)

	deps := DefaultDependencies()
	deps.consul = consulInst

	app, err := NewAppWithDeps("", 80, deps)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/hosts/foobar", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/html")

	app.ServeHTTP(resp, req)

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.Code)
	assert.Contains(t, resp.Body.String(), "Not Found")
}

func TestHAChecksHandler(t *testing.T) {
	consulInst := setupHostsTest()

	deps := DefaultDependencies()
	deps.consul = consulInst

	app, err := NewAppWithDeps("", 80, deps)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/hosts/foo/ha-checks", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/html")

	app.ServeHTTP(resp, req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Code)
	assert.Contains(t, resp.Body.String(), "HA Configuration Checker")
	assert.Contains(t, resp.Body.String(), "<a href=\"/hosts/foo\">foo</a>")
}

func TestHAChecksHandler404Error(t *testing.T) {
	consulInst := new(mocks.Client)
	catalog := new(mocks.Catalog)
	catalog.On("Node", "foobar", (*consulApi.QueryOptions)(nil)).Return((*consulApi.CatalogNode)(nil), nil, nil)
	consulInst.On("Catalog").Return(catalog)

	deps := DefaultDependencies()
	deps.consul = consulInst

	app, err := NewAppWithDeps("", 80, deps)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/hosts/foobar/ha-checks", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/html")

	app.ServeHTTP(resp, req)

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.Code)
	assert.Contains(t, resp.Body.String(), "Not Found")
}
