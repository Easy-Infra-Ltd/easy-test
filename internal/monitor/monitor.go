package monitor

import (
	"context"
	"encoding/json"
	"log/slog"
	"maps"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/api"
	"github.com/Easy-Infra-Ltd/easy-test/internal/assert"
	"github.com/Easy-Infra-Ltd/easy-test/internal/threadpool"
)

type MonitorTarget struct {
	client           *api.Client
	expectedResponse map[string]*json.RawMessage
}

func NewMonitorTarget(client *api.Client, expectedResponse map[string]*json.RawMessage) *MonitorTarget {
	assert.NotNil(client, "Client can not be nil when creating a MonitorTarget")

	return &MonitorTarget{
		client:           client,
		expectedResponse: expectedResponse,
	}
}

type Monitor struct {
	name    string
	targets []*MonitorTarget
	freq    time.Duration
	retries int
	ctx     context.Context
	cancel  context.CancelFunc
	logger  *slog.Logger
}

func NewMonitor(name string, targets []*MonitorTarget, freq time.Duration, retries int) *Monitor {
	assert.Assert(len(targets) > 0, "Can not have 0 clients to monitor")
	assert.Assert(freq > 0, "Frequency can not be 0")

	logger := slog.Default().With("area", "Monitor "+name)

	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		name:    name,
		targets: targets,
		freq:    freq,
		retries: retries,
		ctx:     ctx,
		cancel:  cancel,
		logger:  logger,
	}
}

func (m *Monitor) Start() {
	assert.Assert(len(m.targets) > 0, "When calling Start on Monitor must have more than 0 clients to monitor")
	assert.Assert(m.freq > 0, "When calling Start on Monitor freq must be greater than 0")

	tp := threadpool.NewThreadPool(1, 10, 5*time.Second)
	tp.Run()

	for _, v := range m.targets {
		task := NewMonitorTask(m.ctx, m.name, v, m.freq, m.retries)
		tp.Add(task)
	}

	tp.Wait()
}

type MonitorTask struct {
	name    string
	target  *MonitorTarget
	freq    time.Duration
	retries int
	ctx     context.Context
	logger  *slog.Logger
}

func NewMonitorTask(ctx context.Context, name string, target *MonitorTarget, freq time.Duration, retries int) *MonitorTask {
	assert.NotNil(target, "Target can not be nil when creating a MonitorTask")
	assert.Assert(freq > 0, "Freq must be greater than 0 when creating a MonitorTask")

	logger := slog.Default().With("area", "Monitor Task "+name)

	return &MonitorTask{
		name:    name,
		target:  target,
		freq:    freq,
		retries: 0,
		ctx:     ctx,
		logger:  logger,
	}
}

func (m *MonitorTask) GetName() string {
	return m.name
}

func (m *MonitorTask) Run() {
	assert.NotNil(m.target, "Target should not be nil when trying to run Monitor Task")
	assert.NotNil(m.target.client, "Client should not be nil on the target when trying to run the Monitor Task")

	for i := 0; i < m.retries; i++ {
		select {
		case <-m.ctx.Done():
			m.logger.Info("Monitor finished, exiting")
			return
		default:
			resp, _ := m.target.client.Get()
			defer resp.Body.Close()

			var v map[string]*json.RawMessage
			jsonErr := json.NewDecoder(resp.Body).Decode(v)
			assert.NoError(jsonErr, "Can not error when decoding json from monitored GET request")

			if maps.Equal(m.target.expectedResponse, v) {
				return
			}

			time.Sleep(m.freq)
		}
	}
}
