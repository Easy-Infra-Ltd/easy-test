package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/api"
	"github.com/Easy-Infra-Ltd/easy-test/internal/assert"
	"github.com/Easy-Infra-Ltd/easy-test/internal/threadpool"
)

type MonitorTargetConfig struct {
	Client           *api.ClientConfig `json:"client"`
	Freq             time.Duration     `json:"freq"`
	Retries          int               `json:"retries"`
	ExpectedResponse map[string]any    `json:"expectedResponse"`
}

type MonitorConfig struct {
	Name           string                 `json:"name"`
	MonitorTargets []*MonitorTargetConfig `json:"monitorTargets"`
}

func CreateMonitorTargetsFromConfig(monitorTargetConfig []*MonitorTargetConfig) []*MonitorTarget {
	monitorTargets := make([]*MonitorTarget, 0, len(monitorTargetConfig))
	for _, v := range monitorTargetConfig {
		client := api.NewClient(api.NewClientParams(v.Client.Url, v.Client.ContentType, nil))
		monitorTarget := NewMonitorTarget(client, v.ExpectedResponse, v.Freq*time.Second, v.Retries)

		monitorTargets = append(monitorTargets, monitorTarget)
	}

	return monitorTargets
}

type MonitorTarget struct {
	client           *api.Client
	freq             time.Duration
	retries          int
	expectedResponse map[string]any
}

func NewMonitorTarget(client *api.Client, expectedResponse map[string]any, freq time.Duration, retries int) *MonitorTarget {
	assert.NotNil(client, "Client can not be nil when creating a MonitorTarget")
	assert.Assert(freq > 0, "Frequency can not be 0")

	return &MonitorTarget{
		client:           client,
		freq:             freq,
		retries:          retries,
		expectedResponse: expectedResponse,
	}
}

type Monitor struct {
	name    string
	targets []*MonitorTarget
	ctx     context.Context
	cancel  context.CancelFunc
	logger  *slog.Logger
}

func NewMonitor(name string, targets []*MonitorTarget) *Monitor {
	assert.Assert(len(targets) > 0, "Can not have 0 clients to monitor")

	logger := slog.Default().With("area", "Monitor "+name)

	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		name:    name,
		targets: targets,
		ctx:     ctx,
		cancel:  cancel,
		logger:  logger,
	}
}

func (m *Monitor) Start() {
	assert.Assert(len(m.targets) > 0, "When calling Start on Monitor must have more than 0 clients to monitor")

	tp := threadpool.NewThreadPool(1, 10, 5*time.Second)
	tp.Run()

	m.logger.Info("Adding Monitor Tasks to thread pool")
	for _, v := range m.targets {
		assert.Assert(v.freq > 0, "When calling Start on Monitor freq must be greater than 0")

		task := NewMonitorTask(m.ctx, m.name, v, v.freq, v.retries)
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
		retries: retries,
		ctx:     ctx,
		logger:  logger,
	}
}

func (m MonitorTask) GetName() string {
	return m.name
}

func (m *MonitorTask) Run() {
	for i := 0; i < m.retries; i++ {
		assert.NotNil(m.target, "Target should not be nil when trying to run Monitor Task")
		assert.NotNil(m.target.client, "Client should not be nil on the target when trying to run the Monitor Task")

		select {
		case <-m.ctx.Done():
			m.logger.Info("Monitor finished, exiting")
			return
		default:
			resp, _ := m.target.client.Get()
			m.logger.Info(fmt.Sprintf("Response %+v", resp))

			if resp != nil && resp.Body != nil {
				m.logger.Info(fmt.Sprintf("Response Body %+v", resp.Body))

				var v map[string]any
				jsonErr := json.NewDecoder(resp.Body).Decode(&v)
				assert.NoError(jsonErr, "Can not error when decoding json from monitored GET request")

				resp.Body.Close()

				if reflect.DeepEqual(m.target.expectedResponse, v) {
					m.logger.Info("Successfully found response")
					return
				}
			}

			time.Sleep(m.freq)
		}
	}
}
