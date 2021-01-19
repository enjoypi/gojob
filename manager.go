package gojob

import (
	"context"
	"runtime"
	"sync"

	"go.uber.org/atomic"
	"golang.org/x/sync/semaphore"
)

type OnError func(error)

type taskDetail struct {
	id TaskID
	context.CancelFunc
}

type Manager struct {
	sync.WaitGroup
	context.Context
	sync.Map

	lastTaskID atomic.Int32
	running    atomic.Int32
	sem        *semaphore.Weighted
}

var (
	DefaultManager *Manager
)

func NewManager(maxWorkers int64) *Manager {
	return &Manager{
		Context: context.Background(),
		sem:     semaphore.NewWeighted(maxWorkers),
	}
}
func init() {
	DefaultManager = NewManager(int64(runtime.GOMAXPROCS(0) * 1024))
}

func (m *Manager) Close() {
	m.Map.Range(func(key interface{}, value interface{}) bool {
		cancel := value.(*taskDetail).CancelFunc
		cancel()
		return true
	})
}

func (m *Manager) Go(task Task, onError OnError) {
	if err := m.sem.Acquire(m.Context, 1); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(m.Context)
	taskID := m.lastTaskID.Inc()
	m.Map.Store(taskID, &taskDetail{id: taskID, CancelFunc: cancel})
	m.WaitGroup.Add(1)
	go func() {
		defer cancel()
		defer m.WaitGroup.Done()
		defer m.sem.Release(1)

		err := task(ctx, taskID)
		if err != nil && onError != nil {
			onError(err)
		}
	}()
}

func (m *Manager) Wait() {
	m.WaitGroup.Wait()
}

func Close() {
	DefaultManager.Close()
}

func Go(task Task, onError OnError) {
	DefaultManager.Go(task, onError)
}

func Wait() {
	DefaultManager.Wait()
}
