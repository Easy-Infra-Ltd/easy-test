package simulation_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/api"
	"github.com/Easy-Infra-Ltd/easy-test/internal/logger"
	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
)

type SimulationTestParam struct {
	name     string
	clients  []*api.Client
	attempts int
	cadence  time.Duration
}

func ServerHandlerTesting(http.ResponseWriter, *http.Request) {

}

func TestSimulation(t *testing.T) {
	logger := logger.CreateLoggerFromEnv(nil)
	logger = logger.With("area", "ThreadPool Test").With("process", "test")
	slog.SetDefault(logger)

	server := api.NewServer("Test Server", ":3333")
	server.AddRoute("POST /test", ServerHandlerTesting)

	go server.Start()

	clients := make([]*api.Client, 0, 10)
	for i := 0; i < 10; i++ {
		jsonData, _ := json.Marshal(`{"test": "value"}`)

		clients = append(clients, api.NewClient(api.NewClientParams("http://localhost:3333/test", "application/json", bytes.NewBuffer(jsonData))))
	}

	simulationTests := []SimulationTestParam{
		{
			name:     "3 attempts every 5 seconds",
			clients:  clients,
			attempts: 3,
			cadence:  5 * time.Second,
		},
	}

	t.Parallel()
	for _, v := range simulationTests {
		t.Run(v.name, func(t *testing.T) {
			target := simulation.NewSimulationTarget(v.clients)
			sim := simulation.NewSimulation("Test Simulation", target, v.attempts, v.cadence)

			sim.Start()
		})
	}
}
