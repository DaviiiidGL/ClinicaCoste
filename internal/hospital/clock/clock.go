package clock

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

type SimulatedClock struct {
	mu          sync.Mutex
	currentTime time.Time
	multiplier  float64
}

func NewSimulatedClock(start time.Time, multiplier float64) *SimulatedClock {
	return &SimulatedClock{
		currentTime: start,
		multiplier:  multiplier,
	}
}
func (s *SimulatedClock) Now() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentTime
}

func (s *SimulatedClock) Advance(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentTime = s.currentTime.Add(time.Duration(float64(d) * s.multiplier))
}
