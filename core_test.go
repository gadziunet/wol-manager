package main

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	pbkdf2Iter = 2000
	os.Exit(m.Run())
}

func TestPasswordHash(t *testing.T) {
	h, err := hashPassword("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if !verifyPassword(h, "correct horse battery") {
		t.Error("correct password rejected")
	}
	if verifyPassword(h, "correct horse batter") {
		t.Error("wrong password accepted")
	}
	for _, bad := range []string{"", "x", "pbkdf2-sha256$1$a$b", "md5$2000$YQ$YQ", "pbkdf2-sha256$2000$!!$!!"} {
		if verifyPassword(bad, "anything") {
			t.Errorf("malformed hash %q accepted", bad)
		}
	}
}

func TestLimiter(t *testing.T) {
	l := newLimiter(3, time.Minute, time.Minute)
	now := time.Now()
	for i := 0; i < 2; i++ {
		l.fail("a", now)
		if ok, _ := l.allowed("a", now); !ok {
			t.Fatalf("blocked after %d failures", i+1)
		}
	}
	l.fail("a", now)
	if ok, _ := l.allowed("a", now); ok {
		t.Error("not blocked after 3 failures")
	}
	if ok, _ := l.allowed("b", now); !ok {
		t.Error("other key blocked")
	}
	if ok, _ := l.allowed("a", now.Add(2*time.Minute)); !ok {
		t.Error("still blocked after the block time")
	}
	l.reset("a")
	if ok, _ := l.allowed("a", now); !ok {
		t.Error("blocked after reset")
	}
}

func TestLimiterBounded(t *testing.T) {
	l := newLimiter(3, time.Minute, time.Minute)
	l.maxEntries = 50
	now := time.Now()
	for i := 0; i < 500; i++ {
		l.fail(string(rune('A'+i%26))+time.Duration(i).String(), now.Add(time.Duration(i)*time.Millisecond))
	}
	if len(l.m) > 50 {
		t.Errorf("limiter map grew to %d entries", len(l.m))
	}
}

func TestStorePersistence(t *testing.T) {
	dir := t.TempDir()
	st, err := openStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.CreateFirstUser("admin", "hash"); err != nil {
		t.Fatal(err)
	}
	if err := st.CreateFirstUser("second", "hash"); err != errSetupDone {
		t.Errorf("second setup: got %v, want errSetupDone", err)
	}
	d, err := st.AddDevice(Device{Name: "PC", MAC: "aa:bb:cc:dd:ee:ff", Methods: Methods{Unicast: true}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.AddDevice(Device{Name: "PC2", MAC: "aa:bb:cc:dd:ee:ff"}); err != errDupMAC {
		t.Errorf("duplicate mac: got %v", err)
	}
	if err := st.AddSession(Session{TokenHash: "t1", User: "admin", Expires: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("state.json permissions = %o, want 600", perm)
	}

	st2, err := openStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st2.User("admin"); !ok {
		t.Error("user lost after reopen")
	}
	if got, ok := st2.Device(d.ID); !ok || got.Name != "PC" {
		t.Error("device lost after reopen")
	}
	if u, ok := st2.SessionUser("t1", time.Now()); !ok || u != "admin" {
		t.Error("session lost after reopen")
	}
	if _, ok := st2.SessionUser("t1", time.Now().Add(2*time.Hour)); ok {
		t.Error("expired session accepted")
	}
}

func TestStoreRefusesCorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := openStore(dir); err == nil {
		t.Fatal("corrupt state file was accepted")
	}
	b, _ := os.ReadFile(path)
	if string(b) != "{not json" {
		t.Error("corrupt file was overwritten")
	}
}

func TestDeviceValidation(t *testing.T) {
	yes := true
	good := deviceInput{Name: " PC ", MAC: "AA-BB-CC-DD-EE-FF", Host: "192.168.1.50", Netmask: "255.255.255.0", CheckPorts: []int{3389, 22, 22}, ICMP: &yes}
	d, err := good.device()
	if err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}
	if d.Name != "PC" || d.MAC != "aa:bb:cc:dd:ee:ff" || d.TimeoutSec != 120 || d.RetrySec != 10 {
		t.Errorf("unexpected normalisation: %+v", d)
	}
	if !reflect.DeepEqual(d.CheckPorts, []int{22, 3389}) {
		t.Errorf("ports = %v, want [22 3389]", d.CheckPorts)
	}
	if !d.Methods.Global || !d.Methods.Directed || !d.Methods.Unicast {
		t.Errorf("default methods should all be on: %+v", d.Methods)
	}

	named := deviceInput{Name: "x", MAC: "aa:bb:cc:dd:ee:ff", Host: "nas.example.com"}
	if _, err := named.device(); err != nil {
		t.Errorf("host name rejected: %v", err)
	}

	cases := map[string]deviceInput{
		"invalid_name":     {Name: "", MAC: "aa:bb:cc:dd:ee:ff"},
		"invalid_mac":      {Name: "x", MAC: "nope"},
		"invalid_host":     {Name: "x", MAC: "aa:bb:cc:dd:ee:ff", Host: "300.1.1.1"},
		"invalid_netmask":  {Name: "x", MAC: "aa:bb:cc:dd:ee:ff", Netmask: "255.0.255.0"},
		"invalid_ports":    {Name: "x", MAC: "aa:bb:cc:dd:ee:ff", CheckPorts: []int{0}},
		"invalid_timeout":  {Name: "x", MAC: "aa:bb:cc:dd:ee:ff", TimeoutSec: 5},
		"invalid_retry":    {Name: "x", MAC: "aa:bb:cc:dd:ee:ff", RetrySec: 1},
		"no_method":        {Name: "x", MAC: "aa:bb:cc:dd:ee:ff", Methods: &Methods{}},
		"invalid_interfac": {Name: "x", MAC: "aa:bb:cc:dd:ee:ff", Interface: "a b"},
	}
	for want, in := range cases {
		_, err := in.device()
		fe, ok := err.(*fieldError)
		if !ok {
			t.Errorf("%s: no error", want)
			continue
		}
		if want == "invalid_interfac" {
			want = "invalid_interface"
		}
		if fe.Code != want {
			t.Errorf("got %s, want %s", fe.Code, want)
		}
	}
	for _, h := range []string{"192.168.1.1; rm -rf /", "a b", "-bad.example.com", "::1"} {
		in := deviceInput{Name: "x", MAC: "aa:bb:cc:dd:ee:ff", Host: h}
		if _, err := in.device(); err == nil {
			t.Errorf("host %q accepted", h)
		}
	}
}

func TestPlanTargets(t *testing.T) {
	d := Device{Host: "192.168.2.50", Netmask: "255.255.255.0", Methods: Methods{Global: true, Directed: true, Unicast: true}}
	ifaces := []iface{{Name: "eth0", IP: net.IPv4(192, 168, 1, 22).To4(), Mask: net.CIDRMask(24, 32)}}
	host := net.ParseIP(d.Host).To4()

	got := map[string]bool{}
	for _, tg := range planTargets(d, host, ifaces, []int{9, 7}) {
		got[tg.Method+" "+tg.From.String()+" "+tg.To.String()+" "+strconv.Itoa(tg.Port)] = true
	}
	want := []string{
		"global 192.168.1.22 255.255.255.255 9",
		"global 192.168.1.22 255.255.255.255 7",
		"directed 192.168.1.22 192.168.1.255 9",
		"directed <nil> 192.168.2.255 9",
		"directed <nil> 192.168.2.255 7",
		"unicast <nil> 192.168.2.50 9",
		"unicast <nil> 192.168.2.50 7",
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing target %q", w)
		}
	}
	if len(got) != len(want)+1 {
		t.Errorf("got %d targets, want %d: %v", len(got), len(want)+1, got)
	}

	only := Device{Methods: Methods{Global: true}}
	for _, tg := range planTargets(only, nil, nil, []int{9}) {
		if tg.Method != "global" || !tg.To.Equal(net.IPv4bcast) {
			t.Errorf("unexpected target %+v", tg)
		}
	}
	if n := len(planTargets(Device{Methods: Methods{Unicast: true}}, nil, ifaces, []int{9})); n != 0 {
		t.Errorf("unicast without an IP produced %d targets", n)
	}
}

func TestBroadcastAddr(t *testing.T) {
	cases := []struct {
		ip   string
		bits int
		want string
	}{
		{"192.168.1.10", 24, "192.168.1.255"},
		{"192.168.2.50", 25, "192.168.2.127"},
		{"10.1.2.3", 8, "10.255.255.255"},
		{"192.168.1.10", 32, "192.168.1.10"},
	}
	for _, c := range cases {
		got := broadcastAddr(net.ParseIP(c.ip), net.CIDRMask(c.bits, 32))
		if got.String() != c.want {
			t.Errorf("%s/%d: got %s, want %s", c.ip, c.bits, got, c.want)
		}
	}
}

type fakeProber struct {
	ip     net.IP
	online atomic.Bool
}

func (f *fakeProber) resolve(context.Context, Device) net.IP { return f.ip }
func (f *fakeProber) Check(context.Context, Device) bool     { return f.online.Load() }

func udpSink(t *testing.T) (port int, count func() int) {
	t.Helper()
	pc, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("no UDP socket: %v", err)
	}
	t.Cleanup(func() { pc.Close() })
	var n atomic.Int32
	go func() {
		buf := make([]byte, 256)
		for {
			size, _, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			if size == magicPacketLen {
				n.Add(1)
			}
		}
	}()
	return pc.LocalAddr().(*net.UDPAddr).Port, func() int { return int(n.Load()) }
}

func waitFor(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestManagerWithoutCheckSendsFixedSeries(t *testing.T) {
	port, received := udpSink(t)
	d := Device{
		ID: "t1", Name: "PC", MAC: "aa:bb:cc:dd:ee:ff", Host: "127.0.0.1",
		Methods: Methods{Unicast: true}, TimeoutSec: 30, RetrySec: 1,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := newManager(ctx, []int{port}, &Prober{}, newTracker())
	m.Start(d)
	m.Wait()

	st := m.State("t1")
	if st.State != stateSent {
		t.Errorf("state = %q, want %q (nothing to verify with)", st.State, stateSent)
	}
	if st.Attempt != unverifiedSeries {
		t.Errorf("attempts = %d, want %d", st.Attempt, unverifiedSeries)
	}
	waitFor(t, "all packets", func() bool { return received() >= unverifiedSeries })
}

func TestManagerRetriesUntilOnline(t *testing.T) {
	port, received := udpSink(t)
	fake := &fakeProber{ip: net.IPv4(127, 0, 0, 1).To4()}
	tracker := newTracker()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := newManager(ctx, []int{port}, fake, tracker)

	d := Device{ID: "up", MAC: "aa:bb:cc:dd:ee:01", Host: "127.0.0.1", Methods: Methods{Unicast: true}, ICMP: true, TimeoutSec: 30, RetrySec: 2}
	m.Start(d)
	waitFor(t, "second attempt", func() bool { return m.State("up").Attempt >= 2 })
	if st := m.State("up"); st.State != stateWaking {
		t.Errorf("state = %q while the host is still down, want waking", st.State)
	}
	fake.online.Store(true)
	m.Wait()

	if st := m.State("up"); st.State != stateOnline {
		t.Errorf("state = %q, want online", st.State)
	}
	if v := tracker.View("up", time.Now()); !v.Known || !v.Online {
		t.Errorf("tracker not updated: %+v", v)
	}
	waitFor(t, "packets of two attempts", func() bool { return received() >= 2 })
}

func TestManagerTimesOutAndCancels(t *testing.T) {
	port, _ := udpSink(t)
	fake := &fakeProber{ip: net.IPv4(127, 0, 0, 1).To4()}
	tracker := newTracker()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := newManager(ctx, []int{port}, fake, tracker)

	down := Device{ID: "down", MAC: "aa:bb:cc:dd:ee:02", Host: "127.0.0.1", Methods: Methods{Unicast: true}, ICMP: true, TimeoutSec: 3, RetrySec: 1}
	m.Start(down)
	m.Wait()
	st := m.State("down")
	if st.State != stateTimeout {
		t.Errorf("state = %q, want timeout", st.State)
	}
	if st.Attempt < 2 {
		t.Errorf("attempts = %d, want at least 2 retries", st.Attempt)
	}
	if v := tracker.View("down", time.Now()); !v.Known || v.Online {
		t.Errorf("tracker should say offline: %+v", v)
	}

	long := Device{ID: "long", MAC: "aa:bb:cc:dd:ee:03", Host: "127.0.0.1", Methods: Methods{Unicast: true}, ICMP: true, TimeoutSec: 60, RetrySec: 5}
	m.Start(long)
	waitFor(t, "first attempt", func() bool { return m.State("long").Attempt >= 1 })
	m.Cancel("long")
	m.Wait()
	if st := m.State("long"); st.State != stateCancelled {
		t.Errorf("state = %q, want cancelled", st.State)
	}
}

func TestProberDetectsOpenAndRefusedPorts(t *testing.T) {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("no TCP socket: %v", err)
	}
	defer ln.Close()
	p := &Prober{}
	d := Device{Host: "127.0.0.1", CheckPorts: []int{ln.Addr().(*net.TCPAddr).Port}}
	if !p.Check(context.Background(), d) {
		t.Error("open port not detected")
	}
	if (&Prober{}).Check(context.Background(), Device{Host: "127.0.0.1"}) {
		t.Error("device without any check reported online")
	}
}
