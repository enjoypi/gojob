package gojob

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"testing"
	"time"
)

func init() {

}

func TestTask(t *testing.T) {
	b := make([]byte, 8)
	crand.Read(b)

	u := binary.LittleEndian.Uint64(b)
	t.Log("u", u)

	rand.Seed(int64(u))
	f := func(ctx context.Context, id TaskID) error {
		t.Log("start", id)
		select {
		case <-ctx.Done():
			t.Log("cancel", id)
			return nil
		case <-time.After(time.Millisecond * time.Duration(500+rand.Float32()*1000)):
			t.Log("end", id)
			return nil
		}
	}

	Go(f, nil)
	Go(f, nil)
	Go(f, nil)
	go func() {
		delay := rand.Float32() * 1000
		t.Log("delay", delay)
		time.Sleep(time.Millisecond * time.Duration(delay))
		Close()
	}()
	Wait()
}
