// Package spinner implements a minimal terminal spinner, written by hand so
// skycast doesn't need to pull in an external dependency for it.
package spinner

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const defaultInterval = 80 * time.Millisecond

// Spinner is a start/stop terminal spinner that writes to stderr.
type Spinner struct {
	mu       sync.Mutex
	out      io.Writer
	interval time.Duration
	text     string
	running  bool
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// New returns a stopped Spinner that renders to stderr.
func New() *Spinner {
	return &Spinner{out: os.Stderr, interval: defaultInterval}
}

// SetOutput sets the writer the spinner renders to. Call it before Start.
func (s *Spinner) SetOutput(w io.Writer) {
	s.out = w
}

// SetInterval sets the frame refresh interval. Call it before Start.
func (s *Spinner) SetInterval(d time.Duration) {
	s.interval = d
}

// Start begins rendering the spinner with the given text.
func (s *Spinner) Start(text string) {
	s.mu.Lock()
	s.text = text
	s.running = true
	s.mu.Unlock()

	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})

	go s.run()
}

// SetText updates the text shown next to the spinner.
func (s *Spinner) SetText(text string) {
	s.mu.Lock()
	s.text = text
	s.mu.Unlock()
}

// Stop halts the spinner and clears its line.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	<-s.doneCh

	fmt.Fprint(s.out, "\r\x1b[2K")
}

func (s *Spinner) run() {
	defer close(s.doneCh)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	frame := 0

	for {
		s.mu.Lock()
		text := s.text
		s.mu.Unlock()

		fmt.Fprintf(s.out, "\r\x1b[2K%s %s", frames[frame%len(frames)], text)
		frame++

		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
		}
	}
}
