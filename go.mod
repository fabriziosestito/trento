module github.com/trento-project/trento

go 1.16

require (
	github.com/ClusterLabs/ha_cluster_exporter v0.0.0-20210420075709-eb4566acab09
	github.com/aquasecurity/bench-common v0.4.4
	github.com/cloudquery/sqlite v1.0.1
	github.com/dustinkirkland/golang-petname v0.0.0-20191129215211-8e5a1ed0cff0
	github.com/gin-gonic/gin v1.6.3
	github.com/gomarkdown/markdown v0.0.0-20210514010506-3b9f47219fe7
	github.com/hashicorp/consul-template v0.25.2
	github.com/hashicorp/consul/api v1.4.0
	github.com/hooklift/gowsdl v0.5.0
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.1.2
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/tdewolff/minify/v2 v2.9.16
	github.com/vektra/mockery/v2 v2.9.0
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/tools v0.1.5 // indirect
	gorm.io/gorm v1.21.12
	modernc.org/ccgo/v3 v3.9.6 // indirect
	modernc.org/memory v1.0.5 // indirect
	modernc.org/sqlite v1.11.2 // indirect
)

replace github.com/trento-project/trento => ./
