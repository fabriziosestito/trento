package agent

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/fabriziosestito/phxgoclient"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/trento-project/trento/agent/discovery"
	"github.com/trento-project/trento/agent/discovery/collector"
	"github.com/trento-project/trento/internal"
)

const trentoAgentCheckId = "trentoAgent"

type Agent struct {
	config          *Config
	collectorClient collector.Client
	discoveries     []discovery.Discovery
	ctx             context.Context
	ctxCancel       context.CancelFunc
}

type Config struct {
	InstanceName    string
	SSHAddress      string
	DiscoveryPeriod time.Duration
	CollectorConfig *collector.Config
}

// NewAgent returns a new instance of Agent with the given configuration
func NewAgent(config *Config) (*Agent, error) {
	collectorClient, err := collector.NewCollectorClient(config.CollectorConfig)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a collector client")
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	agent := &Agent{
		config:          config,
		collectorClient: collectorClient,
		ctx:             ctx,
		ctxCancel:       ctxCancel,
		discoveries: []discovery.Discovery{
			discovery.NewClusterDiscovery(collectorClient),
			discovery.NewSAPSystemsDiscovery(collectorClient),
			discovery.NewCloudDiscovery(collectorClient),
			discovery.NewSubscriptionDiscovery(collectorClient),
			discovery.NewHostDiscovery(config.SSHAddress, collectorClient),
		},
	}
	return agent, nil
}

// Start the Agent. This will start the discovery ticker and the heartbeat ticker
func (a *Agent) Start() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Info("Starting Discover loop...")
		defer wg.Done()
		a.startDiscoverTicker()
		log.Info("Discover loop stopped.")
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Info("Starting heartbeat loop...")
		defer wg.Done()
		a.startHeartbeatTicker()
		log.Info("heartbeat loop stopped.")
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Info("Starting checks handler...")
		defer wg.Done()
		a.startChecksHandler()
		log.Info("checks loop stopped.")
	}(&wg)

	wg.Wait()

	return nil
}

func (a *Agent) Stop() {
	a.ctxCancel()
}

// Start a Ticker loop that will iterate over the hardcoded list of Discovery backends and execute them.
func (a *Agent) startDiscoverTicker() {
	tick := func() {
		var output []string
		for _, d := range a.discoveries {
			result, err := d.Discover()
			if err != nil {
				result = fmt.Sprintf("Error while running discovery '%s': %s", d.GetId(), err)

				log.Errorln(result)
			}
			output = append(output, result)
		}
		log.Infof("Discovery tick output: %s", strings.Join(output, "\n\n"))
	}

	interval := a.config.DiscoveryPeriod
	internal.Repeat("agent.discovery", tick, interval, a.ctx)
}

func (a *Agent) startHeartbeatTicker() {
	tick := func() {
		err := a.collectorClient.Heartbeat()
		if err != nil {
			log.Errorf("Error while sending the heartbeat to the server: %s", err)
		}
	}

	internal.Repeat("agent.heartbeat", tick, internal.HeartbeatInterval, a.ctx)
}

const inventoryTemplate = `
[{{.ClusterID}}]
{{.HostID}} ansible_connection=local ansible_host=localhost cluster_selected_checks=[{{ range .Checks}}"{{.}}",{{end}}]
`

func (a *Agent) startChecksHandler() {
	createAnsibleFiles("/tmp/trento")
	machineIDBytes, _ := afero.ReadFile(afero.NewOsFs(), "/etc/machine-id")

	machineID := strings.TrimSpace(string(machineIDBytes))
	agentID := uuid.NewSHA1(internal.TrentoNamespace, []byte(machineID))
	fmt.Println(agentID)
	url := fmt.Sprintf("%s:%d", a.config.CollectorConfig.CollectorHost, a.config.CollectorConfig.CollectorPort)
	socket := phxgoclient.NewPheonixWebsocket(url, "/socket", "ws", false)
	socket.Listen()

	topic := fmt.Sprintf("monitoring:agent_%s", agentID)
	channel, err := socket.OpenChannel(topic)
	if err != nil {
		log.Fatalf("Error while opening the channel: %s", err)
	}

	socket.JoinChannel(topic, nil)
	channel.Register("checks_execution_requested", func(response phxgoclient.MessageResponse) (data interface{}, err error) {
		template := template.Must(template.New("").Parse(inventoryTemplate))
		f, err := os.Create("/tmp/trento/ansible/inventory")
		if err != nil {
			panic(err)
		}
		err = template.Execute(f, map[string]interface{}{
			"ClusterID": response.Payload.Response.(map[string]interface{})["cluster_id"].(string),
			"HostID":    response.Payload.Response.(map[string]interface{})["host_id"].(string),
			"Checks":    response.Payload.Response.(map[string]interface{})["checks"],
		})
		if err != nil {
			panic(err)
		}
		f.Close()

		cmd := exec.Command("ansible-playbook", "/tmp/trento/ansible/check.yml", "-i", "/tmp/trento/ansible/inventory", "--check")
		cmd.Env = os.Environ()
		cmd.Env = append(
			cmd.Env, fmt.Sprintf("TRENTO_WEB_API_HOST=%s", a.config.CollectorConfig.CollectorHost))

		cmd.Env = append(
			cmd.Env, fmt.Sprintf("TRENTO_WEB_API_PORT=%d", a.config.CollectorConfig.CollectorPort))
		out, err := cmd.Output()

		if err != nil {
			log.Errorf("An error occurred while running ansible: %s", err)

			return nil, err
		}
		fmt.Printf("%s\n", out)

		return response, nil
	})

	for {
		select {
		case <-a.ctx.Done():
			return
		default:
			channel.Observe()
		}
	}
}

//go:embed ansible
var ansibleFS embed.FS

func createAnsibleFiles(folder string) error {
	log.Infof("Creating the ansible file structure in %s", folder)
	// Clean the folder if it stores old files
	ansibleFolder := path.Join(folder, "ansible")
	err := os.RemoveAll(ansibleFolder)
	if err != nil {
		log.Error(err)
		return err
	}

	err = os.MkdirAll(ansibleFolder, 0755)
	if err != nil {
		log.Error(err)
		return err
	}

	// Create the ansible file structure from the FS
	err = fs.WalkDir(ansibleFS, "ansible", func(fileName string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !dir.IsDir() {
			content, err := ansibleFS.ReadFile(fileName)
			if err != nil {
				log.Errorf("Error reading file %s", fileName)
				return err
			}
			f, err := os.Create(path.Join(folder, fileName))
			if err != nil {
				log.Errorf("Error creating file %s", fileName)
				return err
			}
			fmt.Fprintf(f, "%s", content)
		} else {
			os.Mkdir(path.Join(folder, fileName), 0755)
		}
		return nil
	})

	if err != nil {
		log.Errorf("An error ocurred during the ansible file structure creation: %s", err)
		return err
	}

	log.Info("Ansible file structure successfully created")

	return nil
}
