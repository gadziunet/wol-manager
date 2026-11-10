package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type iface struct {
	Name string
	IP   net.IP
	Mask net.IPMask
}

func localInterfaces(only string) []iface {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []iface
	for _, ifc := range ifs {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 || ifc.Flags&net.FlagBroadcast == 0 {
			continue
		}
		if only != "" {
			if ifc.Name != only {
				continue
			}
		} else if strings.HasPrefix(ifc.Name, "veth") || strings.HasPrefix(ifc.Name, "docker") {
			continue
		}
		addrs, err := ifc.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			if ip4 := ipn.IP.To4(); ip4 != nil {
				out = append(out, iface{Name: ifc.Name, IP: ip4, Mask: ipn.Mask})
			}
		}
	}
	return out
}

func broadcastAddr(ip net.IP, mask net.IPMask) net.IP {
	ip4 := ip.To4()
	if ip4 == nil {
		return nil
	}
	if len(mask) == net.IPv6len {
		mask = mask[12:]
	}
	if len(mask) != net.IPv4len {
		return nil
	}
	out := make(net.IP, 4)
	for i := range out {
		out[i] = ip4[i] | ^mask[i]
	}
	return out
}

type Target struct {
	Method string
	From   net.IP
	To     net.IP
	Port   int
}

type Result struct {
	Method string `json:"method"`
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Port   int    `json:"port,omitempty"`
	OK     bool   `json:"ok"`
	Err    string `json:"err,omitempty"`
}

func planTargets(d Device, hostIP net.IP, ifaces []iface, ports []int) []Target {
	var out []Target
	seen := map[string]bool{}
	add := func(method string, from, to net.IP) {
		for _, p := range ports {
			key := fmt.Sprintf("%s|%v|%v|%d", method, from, to, p)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, Target{Method: method, From: from, To: to, Port: p})
		}
	}
	if d.Methods.Global {
		if len(ifaces) == 0 {
			add("global", nil, net.IPv4bcast)
		}
		for _, ifc := range ifaces {
			add("global", ifc.IP, net.IPv4bcast)
		}
	}
	if d.Methods.Directed {
		for _, ifc := range ifaces {
			if b := broadcastAddr(ifc.IP, ifc.Mask); b != nil {
				add("directed", ifc.IP, b)
			}
		}
		if hostIP != nil && d.Netmask != "" {
			if m := parseMask(d.Netmask); m != nil {
				if b := broadcastAddr(hostIP, m); b != nil {
					add("directed", nil, b)
				}
			}
		}
	}
	if d.Methods.Unicast && hostIP != nil {
		add("unicast", nil, hostIP)
	}
	return out
}

func shortErr(err error) string {
	s := err.Error()
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}

func sendTargets(mac net.HardwareAddr, targets []Target) []Result {
	res := make([]Result, 0, len(targets))
	for _, t := range targets {
		err := sendPacket(t.From, t.To, t.Port, mac)
		r := Result{Method: t.Method, To: t.To.String(), Port: t.Port, OK: err == nil}
		if t.From != nil {
			r.From = t.From.String()
		}
		if err != nil {
			r.Err = shortErr(err)
		}
		res = append(res, r)
	}
	return res
}

const (
	stateIdle      = "idle"
	stateWaking    = "waking"
	stateOnline    = "online"
	stateTimeout   = "timeout"
	stateSent      = "sent"
	stateCancelled = "cancelled"

	pollEvery        = 2 * time.Second
	unverifiedSeries = 3
)

type WakeState struct {
	State      string   `json:"state"`
	Attempt    int      `json:"attempt"`
	ElapsedSec int      `json:"elapsedSec"`
	TimeoutSec int      `json:"timeoutSec"`
	EndedAgo   int      `json:"endedAgoSec"`
	Sent       int      `json:"sent"`
	Failed     int      `json:"failed"`
	Last       []Result `json:"last"`
}

type job struct {
	cancel  context.CancelFunc
	state   string
	attempt int
	start   time.Time
	end     time.Time
	timeout time.Duration
	sent    int
	failed  int
	last    []Result
}

type hostProber interface {
	resolve(ctx context.Context, d Device) net.IP
	Check(ctx context.Context, d Device) bool
}

type Manager struct {
	mu      sync.Mutex
	jobs    map[string]*job
	ports   []int
	prober  hostProber
	tracker *Tracker
	root    context.Context
	wg      sync.WaitGroup
}

func newManager(root context.Context, ports []int, prober hostProber, tracker *Tracker) *Manager {
	return &Manager{jobs: map[string]*job{}, ports: ports, prober: prober, tracker: tracker, root: root}
}

func (m *Manager) Start(d Device) WakeState {
	mac, err := net.ParseMAC(d.MAC)
	m.mu.Lock()
	if j := m.jobs[d.ID]; j != nil && j.state == stateWaking {
		st := snapshot(j, time.Now())
		m.mu.Unlock()
		return st
	}
	ctx, cancel := context.WithCancel(m.root)
	j := &job{cancel: cancel, state: stateWaking, start: time.Now(), timeout: time.Duration(d.TimeoutSec) * time.Second}
	m.jobs[d.ID] = j
	if err != nil {
		j.state, j.end = stateSent, time.Now()
		j.last = []Result{{Method: "error", Err: "invalid mac"}}
		st := snapshot(j, time.Now())
		m.mu.Unlock()
		cancel()
		return st
	}
	st := snapshot(j, time.Now())
	m.wg.Add(1)
	m.mu.Unlock()
	go m.run(ctx, j, d, mac)
	return st
}

func (m *Manager) Cancel(id string) {
	m.mu.Lock()
	j := m.jobs[id]
	m.mu.Unlock()
	if j != nil {
		j.cancel()
	}
}

func (m *Manager) Forget(id string) {
	m.mu.Lock()
	j := m.jobs[id]
	delete(m.jobs, id)
	m.mu.Unlock()
	if j != nil {
		j.cancel()
	}
}

func (m *Manager) Wait() { m.wg.Wait() }

func (m *Manager) State(id string) WakeState {
	m.mu.Lock()
	defer m.mu.Unlock()
	j := m.jobs[id]
	if j == nil {
		return WakeState{State: stateIdle, Last: []Result{}}
	}
	return snapshot(j, time.Now())
}

func snapshot(j *job, now time.Time) WakeState {
	st := WakeState{
		State:      j.state,
		Attempt:    j.attempt,
		TimeoutSec: int(j.timeout.Seconds()),
		Sent:       j.sent,
		Failed:     j.failed,
		Last:       append([]Result{}, j.last...),
	}
	if j.state == stateWaking {
		st.ElapsedSec = int(now.Sub(j.start).Seconds())
	} else {
		st.ElapsedSec = int(j.end.Sub(j.start).Seconds())
		st.EndedAgo = int(now.Sub(j.end).Seconds())
	}
	return st
}

func (m *Manager) finish(j *job, state string) {
	m.mu.Lock()
	j.state, j.end = state, time.Now()
	m.mu.Unlock()
}

func (m *Manager) record(j *job, attempt int, res []Result) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j.attempt = attempt
	for _, r := range res {
		if r.OK {
			j.sent++
		} else {
			j.failed++
		}
	}
	if len(res) > 64 {
		res = res[:64]
	}
	j.last = res
}

func (m *Manager) run(ctx context.Context, j *job, d Device, mac net.HardwareAddr) {
	defer m.wg.Done()
	defer j.cancel()

	retry := time.Duration(d.RetrySec) * time.Second
	deadline := j.start.Add(j.timeout)
	canDetect := d.Host != "" && (d.ICMP || len(d.CheckPorts) > 0)
	attempts := 0
	next := time.Now()

	for {
		if ctx.Err() != nil {
			m.finish(j, stateCancelled)
			return
		}
		if !time.Now().Before(next) {
			if !canDetect && attempts >= unverifiedSeries {
				m.finish(j, stateSent)
				return
			}
			attempts++
			hostIP := m.prober.resolve(ctx, d)
			targets := planTargets(d, hostIP, localInterfaces(d.Interface), m.ports)
			res := sendTargets(mac, targets)
			if d.Methods.Unicast && hostIP == nil {
				res = append(res, Result{Method: "unicast", Err: "no IP known"})
			}
			m.record(j, attempts, res)
			next = time.Now().Add(retry)
		}
		if canDetect && m.prober.Check(ctx, d) {
			m.tracker.Set(d.ID, true)
			m.finish(j, stateOnline)
			return
		}
		if !time.Now().Before(deadline) {
			if canDetect {
				m.tracker.Set(d.ID, false)
				m.finish(j, stateTimeout)
			} else {
				m.finish(j, stateSent)
			}
			return
		}
		select {
		case <-ctx.Done():
		case <-time.After(pollEvery):
		}
	}
}
