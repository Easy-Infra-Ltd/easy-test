package threadpool_test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/Easy-Infra-Ltd/easy-test/internal/threadpool"
)

type TestTask struct {
	id     int
	logger *slog.Logger
}

func NewTeskTask(id int) *TestTask {
	logger := slog.Default().With("area", fmt.Sprintf("TestTask %d", id))
	return &TestTask{
		id:     id,
		logger: logger,
	}
}

func (t *TestTask) Run() {
	t.logger.Info(fmt.Sprintf("Test task executing %d", t.id))
}

type ThreadPoolTest struct {
	name       string
	maxWorkers int
	tasks      int
}

func TestThreadPool(t *testing.T) {
	tests := []ThreadPoolTest{
		{
			name:       "More tasks than workers",
			maxWorkers: 5,
			tasks:      10,
		},
		{
			name:       "Less tasks than workers",
			maxWorkers: 5,
			tasks:      3,
		},
		{
			name:       "Equal taks and workers",
			maxWorkers: 5,
			tasks:      5,
		},
	}

	t.Parallel()
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			tp := threadpool.NewThreadPool(v.maxWorkers)

			tp.Run()

			for i := 0; i < v.tasks; i++ {
				task := NewTeskTask(i)
				t.Logf("Creating tasks %d", i)
				tp.Add(task)
			}

			tp.Wait()
		})
	}
}
