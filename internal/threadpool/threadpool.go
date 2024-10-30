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
	MIN_WORKERS   = 1
	MAX_WORKERS   = 64
	MIN_IDLE_TIME = 0
)

type Task interface {
	GetName() string
	Run()
}

type worker struct {
	id             uuid.UUID
	threadPool     *ThreadPool
	lastActiveTime time.Time
	logger         *slog.Logger
}

func NewWorker(tp *ThreadPool) *worker {
	assert.NotNil(tp, "ThreadPool should not be nil when creating a worker")

	id := uuid.New()
	logger := slog.Default().With("area", fmt.Sprintf("Worker %s", id.String()))

	logger.Info("Creating new worker")
	return &worker{
		id:             id,
		threadPool:     tp,
		lastActiveTime: time.Now(),
		logger:         logger,
	}
}

func (w *worker) start(ctx context.Context, wg *sync.WaitGroup) {
	assert.NotNil(wg, "WaitGroup should not be nil when starting a worker")

	for {
		select {
		case task := <-w.threadPool.taskQueue:
			w.lastActiveTime = time.Now()
			w.logger.Info("Worker executing task")
			task.Run()
			wg.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (w *worker) stop() {
	w.logger.Info("Stopping worker")
}

type ThreadPool struct {
	wg          *sync.WaitGroup
	taskQueue   chan Task
	workerPool  []*worker
	minWorkers  int
	maxWorkers  int
	idleTimeout time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *slog.Logger
	mutex       sync.Mutex
}

func NewThreadPool(minWorkers int, maxWorkers int, idleTimeout time.Duration) *ThreadPool {
	assert.Assert(minWorkers < maxWorkers, fmt.Sprintf("minWorkers of %d must be less than maxWorkers of %d", minWorkers, maxWorkers))
	assert.Assert(maxWorkers > MIN_WORKERS, fmt.Sprintf("you should never have a thread pool with %d or less workers", MIN_WORKERS))
	assert.Assert(maxWorkers <= MAX_WORKERS, fmt.Sprintf("thread pool max workers should never exceed %d", MAX_WORKERS))
	assert.Assert(idleTimeout > MIN_IDLE_TIME, fmt.Sprintf("Threadpool timeout must be greated than %d seconds", MIN_IDLE_TIME))

	logger := slog.Default().With("area", "ThreadPool")
	logger.Info(fmt.Sprintf("Creating new ThreadPool with %d max workers", maxWorkers))

	ctx, cancel := context.WithCancel(context.Background())

	return &ThreadPool{
		wg:          &sync.WaitGroup{},
		taskQueue:   make(chan Task),
		workerPool:  make([]*worker, 0, maxWorkers),
		minWorkers:  minWorkers,
		maxWorkers:  maxWorkers,
		idleTimeout: idleTimeout,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
	}
}

func (tp *ThreadPool) Run() {
	tp.mutex.Lock()
	for i := 0; i < tp.minWorkers; i++ {
		tp.addWorker()
	}
	tp.mutex.Unlock()

	go tp.monitorIdleWorkers()
}

func (tp *ThreadPool) Add(task Task) error {
	assert.NotNil(task, "Task can not be nil when added to the thread pool")

	tp.logger.Info(fmt.Sprintf("Adding Task to queue %s", task.GetName()))

	tp.wg.Add(1)
	select {
	case <-tp.ctx.Done():
		tp.wg.Done()
		return fmt.Errorf("Threadpool has been shutdown")
	case tp.taskQueue <- task:
		tp.scaleUp()
		return nil
	}
}

func (tp *ThreadPool) scaleUp() {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	tp.logger.Info("Attempting to scale up workers")
	if len(tp.workerPool) < tp.maxWorkers {
		tp.addWorker()
	}
}

func (tp *ThreadPool) addWorker() {
	w := NewWorker(tp)

	tp.logger.Info("Adding new worker to the ThreadPool")
	tp.workerPool = append(tp.workerPool, w)
	go w.start(tp.ctx, tp.wg)
}

func (tp *ThreadPool) monitorIdleWorkers() {
	ticker := time.NewTicker(tp.idleTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tp.cleanupIdleWorkers()
		case <-tp.ctx.Done():
			tp.logger.Info("Context resolved, stopping threadpool monitorIdleWorkers")
			return
		}
	}
}

func (tp *ThreadPool) cleanupIdleWorkers() {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	activeWorkers := make([]*worker, 0, len(tp.workerPool))
	for _, w := range tp.workerPool {
		if time.Since(w.lastActiveTime) > tp.idleTimeout && len(tp.workerPool) > tp.minWorkers {
			w.stop()
		} else {
			activeWorkers = append(activeWorkers, w)
		}
	}

	tp.workerPool = activeWorkers
}

func (tp *ThreadPool) Wait() {
	tp.wg.Wait()
}

func (tp *ThreadPool) Stop() {
	tp.cancel()

	tp.mutex.Lock()
	for _, w := range tp.workerPool {
		w.stop()
	}

	close(tp.taskQueue)
}
