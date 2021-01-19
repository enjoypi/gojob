package gojob

import (
	"context"
	"sync"

	"go.uber.org/atomic"
)

type OnError func(error)

type taskDetail struct {
	id TaskID
	context.CancelFunc
}

type Manager struct {
	sync.WaitGroup
	context.Context
	lastTaskID atomic.Int32
	sync.Map
}

var (
	DefaultManager Manager
)

func init() {
	DefaultManager.Context = context.Background()
}

func (m *Manager) Close() {
	m.Map.Range(func(key interface{}, value interface{}) bool {
		cancel := value.(*taskDetail).CancelFunc
		cancel()
		return true
	})
}

func (m *Manager) Go(task Task, onError OnError) {
	ctx, cancel := context.WithCancel(m.Context)
	taskID := m.lastTaskID.Inc()
	m.Map.Store(taskID, &taskDetail{id: taskID, CancelFunc: cancel})
	m.WaitGroup.Add(1)
	go func() {
		defer cancel()
		defer m.WaitGroup.Done()

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
