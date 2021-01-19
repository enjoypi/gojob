package gojob

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

var (
	errCancelled = fmt.Errorf("cancelled")
	errDone      = fmt.Errorf("done")
)

func TestTask(t *testing.T) {
	b := make([]byte, 8)
	crand.Read(b)
	u := binary.LittleEndian.Uint64(b)
	rand.Seed(int64(u))

	manager := NewManager(int64(runtime.GOMAXPROCS(0)))

	f := func(ctx context.Context, id TaskID) error {
		select {
		case <-ctx.Done():
			return errDone
		case <-time.After(time.Millisecond * time.Duration(rand.Float32()*1000)):
			return errCancelled
		}
	}

	go func() {
		delay := 1000 + rand.Float32()*1000
		time.Sleep(time.Millisecond * time.Duration(delay))
		manager.Close()
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
