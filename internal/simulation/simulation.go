package simulation

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/api"
	"github.com/Easy-Infra-Ltd/easy-test/internal/assert"
	"github.com/Easy-Infra-Ltd/easy-test/internal/threadpool"
	"github.com/google/uuid"
)

type SimulationTarget struct {
	id      uuid.UUID
	clients []*api.Client
}

func NewSimulationTarget(clients []*api.Client) *SimulationTarget {
	assert.Assert(len(clients) > 0, "Simulation Target can not have 0 or less clients")

	return &SimulationTarget{
		id:      uuid.New(),
		clients: clients,
	}
}

type Simulation struct {
	name     string
	target   *SimulationTarget
	attempts int
	cadence  time.Duration
	logger   *slog.Logger
}

func NewSimulation(name string, target *SimulationTarget, attempts int, cadence time.Duration) *Simulation {
	assert.NotNil(target, "Simulation target can not be nil when creating a Simulation")
	assert.Assert(attempts > 0, "Must have at least 1 attempts")

	logger := slog.Default().With("area", fmt.Sprintf("Simulation %s", name))

	logger.Info("Creating new simulation")

	return &Simulation{
		name:     name,
		target:   target,
		attempts: attempts,
		cadence:  cadence,
		logger:   logger,
	}
}

func (s *Simulation) Start() []*api.Client {
	s.logger.Info("Starting Simulation")
	tp := threadpool.NewThreadPool(1, 10, 5*time.Second)
	tp.Run()

	s.logger.Info("ThreadPool Initialised, executing attempts")
	for i := 0; i < s.attempts; i++ {
		for _, v := range s.target.clients {
			s.logger.Info("Adding new simulation task to ThreadPool")
			tp.Add(NewSimulationTask(s.name+" "+s.target.id.String(), func() {
				v.Post()
			}))
		}

		time.Sleep(s.cadence)
	}

	tp.Wait()

	return s.target.clients
}

type SimulationTask struct {
	name string
	task func()
}

func NewSimulationTask(name string, task func()) *SimulationTask {
	return &SimulationTask{
		name: name,
		task: task,
	}
}

func (t *SimulationTask) GetName() string {
	return t.name
}

func (t *SimulationTask) Run() {
	t.task()
}
