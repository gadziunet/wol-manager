package main

import (
	"sync"
	"time"
)

const maxEvents = 500

// Event is one line of the log. Wake events carry the device, account events
// (kind "auth" and "admin") are only shown to administrators.
type Event struct {
	Time     time.Time `json:"time"`
	Kind     string    `json:"kind"`
	Code     string    `json:"code"`
	User     string    `json:"user,omitempty"`
	IP       string    `json:"ip,omitempty"`
	DeviceID string    `json:"deviceId,omitempty"`
	Device   string    `json:"device,omitempty"`
	N        int       `json:"n,omitempty"`
	Results  []Result  `json:"results,omitempty"`
}

// EventLog keeps the most recent events in memory. It is not written to disk,
// so it starts empty after a restart.
type EventLog struct {
	mu   sync.Mutex
	ring []Event
}

func newEventLog() *EventLog { return &EventLog{} }

func (l *EventLog) Add(e Event) {
	if l == nil {
		return
	}
	if e.Time.IsZero() {
		e.Time = time.Now().UTC()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ring = append(l.ring, e)
	if len(l.ring) > maxEvents {
		l.ring = append([]Event(nil), l.ring[len(l.ring)-maxEvents:]...)
	}
}

// Snapshot returns the events newest first. keep decides what the caller may see.
func (l *EventLog) Snapshot(keep func(Event) bool) []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Event, 0, len(l.ring))
	for i := len(l.ring) - 1; i >= 0; i-- {
		if keep(l.ring[i]) {
			out = append(out, l.ring[i])
		}
	}
	return out
}
