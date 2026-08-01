// Package spinner implements a minimal terminal spinner, written by hand so
// skycast doesn't need to pull in an external dependency for it.
package spinner

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const interval = 80 * time.Millisecond

// Spinner is a start/stop terminal spinner that writes to stderr.
type Spinner struct {
	mu      sync.Mutex
	text    string
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// New returns a stopped Spinner.
func New() *Spinner {
	return &Spinner{}
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

	fmt.Fprint(os.Stderr, "\r\x1b[2K")
}

func (s *Spinner) run() {
	defer close(s.doneCh)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	frame := 0

	for {
		s.mu.Lock()
		text := s.text
		s.mu.Unlock()

		fmt.Fprintf(os.Stderr, "\r\x1b[2K%s %s", frames[frame%len(frames)], text)
		frame++

		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
		}
	}
}
