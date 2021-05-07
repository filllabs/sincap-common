package logging

import (
	"sync"

	"go.uber.org/zap"
)

// LoggedWaitGroup is a wait group with loggging.
type LoggedWaitGroup struct {
	sync.WaitGroup
	Name  string
	count int
}

// Addl is basicly an Add function with log
func (l *LoggedWaitGroup) Addl(name string, delta int) {
	l.Add(delta)
	l.count += delta
	Logger.Named(l.Name).Debug("Add", zap.String("caller", name), zap.Int("delta", delta), zap.Int("count", l.count))
}

// Donel is basicly an Done function with log
func (l *LoggedWaitGroup) Donel(name string) {
	l.Done()
	l.count--
	Logger.Named(l.Name).Debug("Done", zap.String("caller", name), zap.Int("count", l.count))
}
