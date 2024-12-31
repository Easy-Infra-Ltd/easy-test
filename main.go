package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Easy-Infra-Ltd/easy-test/internal/assert"
	"github.com/Easy-Infra-Ltd/easy-test/internal/logger"
	"github.com/Easy-Infra-Ltd/easy-test/internal/simulation"
)

func main() {
	// TODO: Add ability to dry run simulation
	var path string
	flag.StringVar(&path, "path", "simulation.json", "Path to simulation configuration file")
	dry := flag.Bool("dry", false, "Allow you to run the simulation without making any requests externally to ensure you have it setup correctly")
	flag.Parse()

	logger := logger.CreateLoggerFromEnv(nil, "lightGreen")
	logger = logger.With("process", "Simulation").With("area", "Simulation CLI")
	slog.SetDefault(logger)

	file, err := os.Open(path)
	assert.NoError(err, "Failed to load configuration file")
	defer file.Close()

	var simConfig simulation.SimulationConfig
	decoder := json.NewDecoder(file)
	jErr := decoder.Decode(&simConfig)
	assert.NoError(jErr, "Failed to unmarshal json when loading configuration file")

	logger.Info(fmt.Sprintf("config: %+v\n", simConfig))

	sim := simulation.NewSimulationFromConfig(&simConfig, *dry)
	sim.Start()
}
