package main

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const probeTimeout = 1500 * time.Millisecond

type Prober struct {
	store *Store
}

func (p *Prober) resolve(ctx context.Context, d Device) net.IP {
	if d.Host == "" {
		return nil
	}
	if ip := net.ParseIP(d.Host); ip != nil {
		return ip.To4()
	}
	rctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIP(rctx, "ip4", d.Host)
	if err == nil && len(ips) > 0 {
		ip := ips[0].To4()
		if ip != nil && p.store != nil && ip.String() != d.LastIP {
			_ = p.store.SetLastIP(d.ID, ip.String())
		}
		return ip
	}
	if d.LastIP != "" {
		return net.ParseIP(d.LastIP).To4()
	}
	return nil
}

func tcpUp(ctx context.Context, ip net.IP, port int) bool {
	dialer := net.Dialer{Timeout: probeTimeout}
	conn, err := dialer.DialContext(ctx, "tcp4", net.JoinHostPort(ip.String(), strconv.Itoa(port)))
	if err == nil {
		conn.Close()
		return true
	}
	return errors.Is(err, syscall.ECONNREFUSED)
}

func (p *Prober) Check(ctx context.Context, d Device) bool {
	if !d.ICMP && len(d.CheckPorts) == 0 {
		return false
	}
	ip := p.resolve(ctx, d)
	if ip == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	n := len(d.CheckPorts)
	if d.ICMP {
		n++
	}
	res := make(chan bool, n)
	if d.ICMP {
		go func() { res <- pingICMP(ip, probeTimeout) }()
	}
	for _, port := range d.CheckPorts {
		go func() { res <- tcpUp(ctx, ip, port) }()
	}
	for i := 0; i < n; i++ {
		select {
		case ok := <-res:
			if ok {
				return true
			}
		case <-ctx.Done():
			return false
		}
	}
	return false
}

type statusEntry struct {
	online bool
	at     time.Time
}

type Tracker struct {
	mu sync.RWMutex
	m  map[string]statusEntry
}

func newTracker() *Tracker { return &Tracker{m: map[string]statusEntry{}} }

func (t *Tracker) Set(id string, online bool) {
	t.mu.Lock()
	t.m[id] = statusEntry{online: online, at: time.Now()}
	t.mu.Unlock()
}

type StatusView struct {
	Known  bool `json:"known"`
	Online bool `json:"online"`
	AgeSec int  `json:"ageSec"`
}

func (t *Tracker) View(id string, now time.Time) StatusView {
	t.mu.RLock()
	defer t.mu.RUnlock()
	e, ok := t.m[id]
	if !ok {
		return StatusView{}
	}
	return StatusView{Known: true, Online: e.online, AgeSec: int(now.Sub(e.at).Seconds())}
}

func (t *Tracker) Prune(keep map[string]bool) {
	t.mu.Lock()
	for id := range t.m {
		if !keep[id] {
			delete(t.m, id)
		}
	}
	t.mu.Unlock()
}

type Monitor struct {
	store   *Store
	prober  *Prober
	tracker *Tracker
	every   time.Duration
}

func (m *Monitor) Run(ctx context.Context) {
	m.cycle(ctx)
	ticker := time.NewTicker(m.every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.cycle(ctx)
		}
	}
}

func (m *Monitor) cycle(ctx context.Context) {
	devs := m.store.Devices()
	keep := make(map[string]bool, len(devs))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for _, d := range devs {
		keep[d.ID] = true
		if d.Host == "" || (!d.ICMP && len(d.CheckPorts) == 0) {
			continue
		}
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil {
				return
			}
			m.tracker.Set(d.ID, m.prober.Check(ctx, d))
		}()
	}
	wg.Wait()
	m.tracker.Prune(keep)
}
