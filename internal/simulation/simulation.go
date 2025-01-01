package simulation

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/api"
	"github.com/Easy-Infra-Ltd/easy-test/internal/assert"
	"github.com/Easy-Infra-Ltd/easy-test/internal/monitor"
	"github.com/Easy-Infra-Ltd/easy-test/internal/threadpool"
	"github.com/google/uuid"
)

type SimulationMonitorConfig struct {
	name           string
	monitorTargets []*monitor.MonitorTarget
}

type SimulationTarget struct {
	id      uuid.UUID
	clients []*api.Client
	monitor *SimulationMonitorConfig
}

func NewSimulationTarget(clients []*api.Client, monitor *SimulationMonitorConfig) *SimulationTarget {
	assert.Assert(len(clients) > 0, "Simulation Target can not have 0 or less clients")

	return &SimulationTarget{
		id:      uuid.New(),
		clients: clients,
		monitor: monitor,
	}
}

type SimulationTargetConfig struct {
	Count   int                    `json:"count"`
	Client  *api.ClientConfig      `json:"client"`
	Monitor *monitor.MonitorConfig `json:"monitor"`
}

type SimulationConfig struct {
	Name     string                 `json:"name"`
	Target   SimulationTargetConfig `json:"target"`
	Cadence  time.Duration          `json:"cadence"`
	Attempts int                    `json:"attempts"`
}

func NewSimulationFromConfig(simConfig *SimulationConfig, dry bool) *Simulation {
	// TODO: Allow for simulation to be created from config without Monitors
	monitorTargets := monitor.CreateMonitorTargetsFromConfig(simConfig.Target.Monitor.MonitorTargets)
	monitorConfig := &SimulationMonitorConfig{
		name:           simConfig.Target.Monitor.Name,
		monitorTargets: monitorTargets,
	}

	clients := make([]*api.Client, 0, simConfig.Target.Count)
	for i := 0; i < simConfig.Target.Count; i++ {
		client := api.NewClient(api.NewClientParams(simConfig.Target.Client.Url, simConfig.Target.Client.ContentType, nil))

		clients = append(clients, client)
	}

	simTarget := NewSimulationTarget(clients, monitorConfig)
	return NewSimulation(simConfig.Name, simTarget, simConfig.Attempts, simConfig.Cadence*time.Second, dry)
}

type Simulation struct {
	id       uuid.UUID
	name     string
	target   *SimulationTarget
	attempts int
	cadence  time.Duration
	logger   *slog.Logger
	dry      bool
}

func NewSimulation(name string, target *SimulationTarget, attempts int, cadence time.Duration, dry bool) *Simulation {
	assert.NotNil(target, "Simulation target can not be nil when creating a Simulation")
	assert.Assert(attempts > 0, "Must have at least 1 attempts")

	id := uuid.New()
	logger := slog.Default().With("area", fmt.Sprintf("Simulation %s %s", id.String(), name))

	logger.Info("Creating new simulation")

	return &Simulation{
		id:       id,
		name:     name,
		target:   target,
		attempts: attempts,
		cadence:  cadence,
		logger:   logger,
		dry:      dry,
	}
}

func (s *Simulation) Start() []*api.Client {
	assert.NotNil(s, "Simulation can not be nil when calling start on it")
	assert.NotNil(s.target, "SimulationTarget can not be nill when calling start on a Simulation")
	assert.Assert(len(s.target.clients) > 0, "When calling Simulation Start the target for the Simulation must have at least one client")

	s.logger.Info("Starting Simulation")
	tp := threadpool.NewThreadPool(1, 10, 5*time.Second)
	tp.Run()

	s.logger.Info("ThreadPool Initialised, executing attempts")
	for i := 0; i < s.attempts; i++ {
		for _, v := range s.target.clients {
			s.logger.Info("Adding new simulation task to ThreadPool")
			tp.Add(NewSimulationTask(s.name+" "+s.id.String(), func() string {
				// TODO: Make this execute some Lua Script
				resp, err := v.Post()
				if err != nil {
					s.logger.Error(err.Error())
					return ""
				}

				assert.NotNil(resp, "Response from Post can not be nil")
				assert.NotNil(resp.Body, "Response Body can not be nil")

				var id string

				// TODO: Extract the ID
				return id
			}, s.target.monitor))
		}

		time.Sleep(s.cadence)
	}

	tp.Wait()

	return s.target.clients
}

type SimulationTaskFunc func() string

type SimulationTask struct {
	name    string
	task    SimulationTaskFunc
	monitor *SimulationMonitorConfig
	logger  *slog.Logger
}

func NewSimulationTask(name string, task SimulationTaskFunc, monitor *SimulationMonitorConfig) *SimulationTask {
	logger := slog.Default().With("area", "SimulationTask "+name)
	return &SimulationTask{
		name:    name,
		task:    task,
		monitor: monitor,
		logger:  logger,
	}
}

func (t SimulationTask) GetName() string {
	return t.name
}

func (t *SimulationTask) Run() {
	id := t.task()
	monitor := t.CreateMonitor(id)

	monitor.Start()
}

func (t *SimulationTask) CreateMonitor(id string) *monitor.Monitor {
	if t.monitor == nil {
		t.logger.Info("No monitors configured for SimulationTask")
		return nil
	}

	return monitor.NewMonitor(t.monitor.name, t.monitor.monitorTargets)
}
