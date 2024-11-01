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
	tp := threadpool.NewThreadPool(1, 10, 5*time.Second)
	tp.Run()

	for _, v := range s.target.clients {
		tp.Add(NewSimulationTask(s.name+" "+s.target.id.String(), func() {
			v.Post()
		}))
	}

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
