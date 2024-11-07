package threadpool_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/logger"
	"github.com/Easy-Infra-Ltd/easy-test/internal/threadpool"
)

type TestTask struct {
	id     int
	name   string
	logger *slog.Logger
}

func NewTeskTask(id int, name string) *TestTask {
	logger := slog.Default().With("area", fmt.Sprintf("TestTask %d for test %s", id, name))
	return &TestTask{
		id:     id,
		name:   name,
		logger: logger,
	}
}

func (t *TestTask) GetName() string {
	return fmt.Sprintf("Task %d for %s", t.id, t.name)
}

func (t *TestTask) Run() {
	t.logger.Info(fmt.Sprintf("Test task executing %d", t.id))
	time.Sleep(2 * time.Second)
}

type ThreadPoolTest struct {
	name       string
	maxWorkers int
	tasks      int
	attempts   int
	idleTime   time.Duration
	pause      time.Duration
}

func TestThreadPool(t *testing.T) {
	logger := logger.CreateLoggerFromEnv(os.Stdout)
	logger = logger.With("area", "ThreadPool Test").With("process", "test")
	slog.SetDefault(logger)

	tests := []ThreadPoolTest{
		{
			name:       "More tasks than workers",
			maxWorkers: 5,
			tasks:      10,
			attempts:   1,
			idleTime:   5 * time.Second,
			pause:      0,
		},
		{
			name:       "Less tasks than workers",
			maxWorkers: 5,
			tasks:      3,
			attempts:   1,
			idleTime:   5 * time.Second,
			pause:      0,
		},
		{
			name:       "Equal tasks and workers",
			maxWorkers: 5,
			tasks:      5,
			attempts:   1,
			idleTime:   5 * time.Second,
			pause:      0,
		},
		{
			name:       "Equal tasks and workers with 5 attempts every 3 seconds",
			maxWorkers: 5,
			tasks:      5,
			attempts:   5,
			idleTime:   5 * time.Second,
			pause:      3 * time.Second,
		},
		{
			name:       "Equal tasks and workers with 5 attempts and idle time is equal to pause",
			maxWorkers: 5,
			tasks:      5,
			attempts:   5,
			idleTime:   5 * time.Second,
			pause:      3 * time.Second,
		},
	}

	t.Parallel()
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			tp := threadpool.NewThreadPool(1, v.maxWorkers, v.idleTime)

			tp.Run()

			for j := 0; j < v.attempts; j++ {
				for i := 0; i < v.tasks; i++ {
					task := NewTeskTask(i, v.name)
					t.Logf("Creating tasks %d", i)
					err := tp.Add(task)
					if err != nil {
						t.Errorf("ThreadPool has been shutdown before task has been added")
					}
				}

				time.Sleep(v.pause)
			}

			tp.Wait()
			tp.Stop()
		})
	}
}
