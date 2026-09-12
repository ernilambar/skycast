package spinner

import (
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuffer is a goroutine-safe buffer so tests can read what the spinner's
// rendering goroutine writes without racing it.
type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

func newTestSpinner() (*Spinner, *syncBuffer) {
	buf := &syncBuffer{}
	s := New()
	s.SetOutput(buf)
	s.SetInterval(time.Millisecond)

	return s, buf
}

func TestSpinnerRendersFrameWithText(t *testing.T) {
	s, buf := newTestSpinner()

	s.Start("Looking up Tokyo...")
	s.Stop()

	out := buf.String()
	if !strings.Contains(out, "Looking up Tokyo...") {
		t.Errorf("spinner output = %q, want it to contain the spinner text", out)
	}
	if !strings.Contains(out, frames[0]) {
		t.Errorf("spinner output = %q, want it to contain the first frame %q", out, frames[0])
	}
}

func TestSpinnerClearsLineOnStop(t *testing.T) {
	s, buf := newTestSpinner()

	s.Start("Loading...")
	s.Stop()

	if !strings.HasSuffix(buf.String(), "\r\x1b[2K") {
		t.Errorf("spinner output = %q, want it to end with the clear-line sequence", buf.String())
	}
}

func TestSpinnerRendersUpdatedText(t *testing.T) {
	s, buf := newTestSpinner()

	s.Start("first")
	s.SetText("second")

	deadline := time.After(time.Second)
	for !strings.Contains(buf.String(), "second") {
		select {
		case <-deadline:
			s.Stop()
			t.Fatalf("spinner output = %q, want it to contain the updated text", buf.String())
		default:
			time.Sleep(time.Millisecond)
		}
	}

	s.Stop()
}

func TestSpinnerStopIsIdempotent(t *testing.T) {
	s, _ := newTestSpinner()

	s.Start("Loading...")
	s.Stop()

	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("second Stop() did not return")
	}
}

func TestSpinnerStopWithoutStartDoesNothing(t *testing.T) {
	s, buf := newTestSpinner()

	s.Stop()

	if buf.String() != "" {
		t.Errorf("spinner output = %q, want no output when it was never started", buf.String())
	}
}
