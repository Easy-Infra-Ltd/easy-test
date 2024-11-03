package monitor

import (
	"context"
	"log/slog"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/api"
	"github.com/Easy-Infra-Ltd/easy-test/internal/assert"
	"github.com/Easy-Infra-Ltd/easy-test/internal/threadpool"
)

type MonitorTarget struct {
	clients []*api.Client
}

type Monitor struct {
	name   string
	target *MonitorTarget
	freq   time.Duration
	ctx    context.Context
	cancel context.CancelFunc
	logger *slog.Logger
}

func NewMonitor(name string, freq time.Duration) *Monitor {
	assert.Assert(freq > 0, "Frequency can not be 0")

	logger := slog.Default().With("area", "Monitor "+name)

	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		name:   name,
		freq:   freq,
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}
}

func (m *Monitor) Start() {
	tp := threadpool.NewThreadPool(1, 10, 5*time.Second)
	t := time.NewTicker(m.freq)
	for {
		select {
		case <-t.C:
			task := NewMonitorTask(m.name)
			tp.Add(task)
		case <-m.ctx.Done():
			t.Stop()
			m.logger.Info("Monitor finished, exiting")
			return
		default:
			assert.Never("Should never reach monitor default")
		}
	}
}

type MonitorTask struct {
	name    string
	clients []*api.Client
	logger  *slog.Logger
}

func NewMonitorTask(name string, clients []*api.Client) *MonitorTask {
	assert.Assert(len(clients) > 0, "Can not pass an empty clients array to monitor task")

	logger := slog.Default().With("area", "Monitor Task "+name)

	return &MonitorTask{
		name:    name,
		clients: clients,
		logger:  logger,
	}
}

func (m *MonitorTask) GetName() string {
	return m.name
}

func (m *MonitorTask) Run() {
	for _, v := range m.clients {
		v.Get()
	}
}
