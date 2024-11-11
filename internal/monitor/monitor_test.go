package monitor_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/api"
	"github.com/Easy-Infra-Ltd/easy-test/internal/logger"
	"github.com/Easy-Infra-Ltd/easy-test/internal/monitor"
)

type MonitorTestParams struct {
	name     string
	cliCount int
	freq     time.Duration
	retries  int
}

func handleGetTest(res http.ResponseWriter, req *http.Request) {

}

func TestMonitor(t *testing.T) {
	logger := logger.CreateLoggerFromEnv(nil)
	logger = logger.With("area", "Monitor Test").With("process", "test")
	slog.SetDefault(logger)

	s := api.NewServer("TestServer", ":3333")
	s.AddRoute("GET /test", handleGetTest)

	go s.Start()

	t.Parallel()

	tests := []MonitorTestParams{
		{
			name:     "3 Clients every 3 seconds 3 retries",
			cliCount: 3,
			freq:     3,
			retries:  3,
		},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			targets := make([]*monitor.MonitorTarget, 0, v.cliCount)
			for i := 0; i < v.cliCount; i++ {
				cli := api.NewClient(api.NewClientParams("https://localhost:3333/test", "application/json", nil))

				var expectedResponse map[string]*json.RawMessage
				targets = append(targets, monitor.NewMonitorTarget(cli, expectedResponse))
			}

			m := monitor.NewMonitor("Monitor Site", targets, v.freq, v.retries)
			m.Start()
		})
	}
}
