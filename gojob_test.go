package gojob

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

var (
	errCancelled = fmt.Errorf("cancelled")
	errDone      = fmt.Errorf("done")
)

func randDur(max uint16) time.Duration {
	b := make([]byte, 2)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	var v uint16
	err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &v)
	if err != nil {
		panic(err)
	}
	return time.Duration(v%max + 1)
}

func TestTask(t *testing.T) {
	manager := NewManager(4)

	f := func(ctx context.Context, id TaskID) error {
		running := randDur(1000)
		//t.Log(id, "running", running)
		select {
		case <-ctx.Done():
			//t.Log(id, "cancelled")
			return errCancelled
		case <-time.After(time.Millisecond * running):
			//t.Log(id, "done")
			return errDone
		}
	}

	go func() {
		for {
			delay := randDur(1000)
			//t.Log("delay", delay)
			time.Sleep(time.Millisecond * delay)
			manager.Close()
		}
	}()

	var done, cancelled atomic.Int32
	n := 16
	for i := 0; i < n; i++ {
		manager.Go(f, nil, func(err error) {
			if err == errDone {
				done.Inc()
			}

			if err == errCancelled {
				cancelled.Inc()
			}
		})
	}

	manager.Wait()
	require.Equal(t, int32(n), done.Load()+cancelled.Load())
	t.Log("done", done.Load(), "cancelled", cancelled.Load())
}
