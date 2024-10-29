package threadpool

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/assert"
	"github.com/google/uuid"
)

const (
	MIN_WORKERS      = 0
	MAX_WORKERS      = 64
	WORKER_IDLE_TIME = 5 * time.Second
)

type Task interface {
	Run()
}

type Worker struct {
	id         uuid.UUID
	taskQueue  chan Task
	workerPool chan *Worker
	threadPool *ThreadPool
	idleTime   time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *slog.Logger
}

func NewWorker(wp chan *Worker, tp *ThreadPool, idleTime time.Duration) *Worker {
	assert.NotNil(tp, "ThreadPool should not be nil when creating a worker")
	assert.Assert(idleTime > time.Second, "Idle time should be greater than a second")

	id := uuid.New()
	ctx, cancel := context.WithCancel(tp.ctx)
	logger := slog.Default().With("area", fmt.Sprintf("Worker %s", id.String()))
	return &Worker{
		id:         id,
		taskQueue:  make(chan Task),
		workerPool: wp,
		threadPool: tp,
		idleTime:   idleTime,
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
	}
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	assert.NotNil(wg, "WaitGroup should not be nil when starting a worker")

	ctxDoneStr := fmt.Sprintf("Stopping worker %s as context has resolved", w.id.String())
	go func() {
		defer func() {
			w.logger.Info(fmt.Sprintf("Remove worker %s", w.id.String()))
			w.threadPool.RemoveWorker()
		}()

		for {
			select {
			case w.workerPool <- w:
				select {
				case task := <-w.taskQueue:
					task.Run()
					wg.Done()
				case <-time.After(w.idleTime):
					w.logger.Info(fmt.Sprintf("Stopping worker %s as idle for %d", w.id.String(), w.idleTime))
					return
				case <-w.ctx.Done():
					w.logger.Info(ctxDoneStr)
					return
				}
			case <-w.ctx.Done():
				w.logger.Info(ctxDoneStr)
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.cancel()
}

type ThreadPool struct {
	wg          *sync.WaitGroup
	taskQueue   chan Task
	workerPool  chan *Worker
	activeCount int
	maxWorkers  int
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *slog.Logger
	mutex       sync.Mutex
}

func NewThreadPool(maxWorkers int) *ThreadPool {
	assert.Assert(maxWorkers > MIN_WORKERS, fmt.Sprintf("you should never have a thread pool with %d or less workers", MIN_WORKERS))
	assert.Assert(maxWorkers <= MAX_WORKERS, fmt.Sprintf("thread pool max workers should never exceed %d", MAX_WORKERS))

	logger := slog.Default().With("area", "ThreadPool")
	logger.Info(fmt.Sprintf("Creating new runner with %d max workers", maxWorkers))

	ctx, cancel := context.WithCancel(context.Background())

	return &ThreadPool{
		wg:          &sync.WaitGroup{},
		taskQueue:   make(chan Task),
		workerPool:  make(chan *Worker, maxWorkers),
		activeCount: 0,
		maxWorkers:  maxWorkers,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
	}
}

func (tp *ThreadPool) Run() {
	go tp.dispatch()
}

func (tp *ThreadPool) dispatch() {
	for {
		select {
		case task := <-tp.taskQueue:
			select {
			case worker := <-tp.workerPool:
				tp.logger.Info(fmt.Sprintf("Adding task to %s", worker.id.String()))
				worker.taskQueue <- task
			default:
				tp.mutex.Lock()
				if tp.activeCount < tp.maxWorkers {
					tp.activeCount++

					tp.logger.Info(fmt.Sprintf("Creating new worker as %d is less than max workers %d", tp.activeCount, tp.maxWorkers))
					tp.mutex.Unlock()

					worker := NewWorker(tp.workerPool, tp, WORKER_IDLE_TIME)
					worker.Start(tp.wg)
					worker.taskQueue <- task
				} else {
					tp.mutex.Unlock()

					worker := <-tp.workerPool
					worker.taskQueue <- task
				}
			}
		case <-tp.ctx.Done():
			tp.logger.Info("Context resolved, stopping threadpool")
			return
		}
	}
}

func (tp *ThreadPool) RemoveWorker() {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	tp.activeCount--
}

func (tp *ThreadPool) Add(task Task) {
	assert.NotNil(task, "Task can not be nil when added to the thread pool")

	tp.wg.Add(1)
	tp.taskQueue <- task
}

func (tp *ThreadPool) Wait() {
	tp.wg.Wait()
}

func (tp *ThreadPool) Stop() {
	tp.cancel()
}
