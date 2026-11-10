package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testApp struct {
	t      *testing.T
	srv    *Server
	ts     *httptest.Server
	client *http.Client
}

func newTestApp(t *testing.T, ports []int) *testApp {
	t.Helper()
	st, err := openStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	tracker := newTracker()
	wakes := newManager(ctx, ports, &Prober{store: st}, tracker)
	srv := newServer(Config{Ports: ports}, st, wakes, tracker)
	ts := httptest.NewServer(srv.routes())
	t.Cleanup(ts.Close)
	jar, _ := cookiejar.New(nil)
	return &testApp{t: t, srv: srv, ts: ts, client: &http.Client{Jar: jar, Timeout: 10 * time.Second}}
}

func (a *testApp) do(method, path string, body any, headers map[string]string) (int, map[string]any, http.Header) {
	a.t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader([]byte("{}"))
	}
	req, err := http.NewRequest(method, a.ts.URL+path, rd)
	if err != nil {
		a.t.Fatal(err)
	}
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := a.client.Do(req)
	if err != nil {
		a.t.Fatal(err)
	}
	defer res.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res.StatusCode, out, res.Header
}

func (a *testApp) doList(path string) (int, []map[string]any) {
	a.t.Helper()
	res, err := a.client.Get(a.ts.URL + path)
	if err != nil {
		a.t.Fatal(err)
	}
	defer res.Body.Close()
	var out []map[string]any
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res.StatusCode, out
}

func (a *testApp) setup() {
	a.t.Helper()
	code, out, _ := a.do("POST", "/api/setup", map[string]string{"token": a.srv.SetupToken(), "username": "Admin", "password": "a-long-password"}, nil)
	if code != 200 {
		a.t.Fatalf("setup: %d %v", code, out)
	}
}

func TestFullFlow(t *testing.T) {
	pc, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("no UDP socket: %v", err)
	}
	defer pc.Close()
	udpPort := pc.LocalAddr().(*net.UDPAddr).Port

	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("no TCP socket: %v", err)
	}
	defer ln.Close()
	tcpPort := ln.Addr().(*net.TCPAddr).Port

	a := newTestApp(t, []int{udpPort})

	code, out, _ := a.do("GET", "/api/session", nil, nil)
	if code != 200 || out["setupRequired"] != true || out["authenticated"] != false {
		t.Fatalf("fresh session: %d %v", code, out)
	}

	if code, _ := a.doList("/api/devices"); code != 401 {
		t.Errorf("unauthenticated devices: %d, want 401", code)
	}

	if code, out, _ = a.do("POST", "/api/setup", map[string]string{"token": "wrong", "username": "admin", "password": "a-long-password"}, nil); code != 403 {
		t.Errorf("wrong setup token: %d %v", code, out)
	}
	a.setup()

	if code, out, _ = a.do("POST", "/api/setup", map[string]string{"token": a.srv.SetupToken(), "username": "other", "password": "another-long-pass"}, nil); code != 409 {
		t.Errorf("second setup: %d %v", code, out)
	}

	code, out, _ = a.do("GET", "/api/session", nil, nil)
	if out["authenticated"] != true || out["user"] != "admin" {
		t.Fatalf("session after setup: %d %v", code, out)
	}

	body := map[string]any{
		"name": "Desktop", "mac": "AA:BB:CC:DD:EE:FF", "host": "127.0.0.1",
		"methods":    map[string]bool{"global": false, "directed": false, "unicast": true},
		"checkPorts": []int{tcpPort}, "icmp": false, "timeoutSec": 10, "retrySec": 2,
	}
	code, out, _ = a.do("POST", "/api/devices", body, nil)
	if code != 201 {
		t.Fatalf("create device: %d %v", code, out)
	}
	id, _ := out["id"].(string)
	if id == "" || out["mac"] != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("created device: %v", out)
	}

	if code, out, _ = a.do("POST", "/api/devices", body, nil); code != 409 || out["error"] != "duplicate_mac" {
		t.Errorf("duplicate: %d %v", code, out)
	}

	bad := map[string]any{"name": "", "mac": "x"}
	if code, out, _ = a.do("POST", "/api/devices", bad, nil); code != 400 || out["error"] != "invalid_name" {
		t.Errorf("invalid: %d %v", code, out)
	}
	if code, out, _ = a.do("POST", "/api/devices", map[string]any{"name": "x", "mac": "aa:bb:cc:dd:ee:00", "surprise": 1}, nil); code != 400 {
		t.Errorf("unknown field accepted: %d %v", code, out)
	}

	code, out, _ = a.do("POST", "/api/devices/"+id+"/wake", nil, nil)
	if code != 202 {
		t.Fatalf("wake: %d %v", code, out)
	}
	_ = pc.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 256)
	n, _, err := pc.ReadFrom(buf)
	if err != nil || n != magicPacketLen {
		t.Fatalf("magic packet not received: n=%d err=%v", n, err)
	}

	var state string
	for i := 0; i < 40; i++ {
		_, list := a.doList("/api/devices")
		if len(list) != 1 {
			t.Fatalf("device list: %v", list)
		}
		wake := list[0]["wake"].(map[string]any)
		state, _ = wake["state"].(string)
		if state != "waking" {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if state != "online" {
		t.Errorf("wake state = %q, want online", state)
	}

	upd := map[string]any{"name": "Desktop 2", "mac": "aa:bb:cc:dd:ee:ff", "host": "127.0.0.1", "netmask": "255.255.255.0", "icmp": false, "checkPorts": []int{}}
	if code, out, _ = a.do("PUT", "/api/devices/"+id, upd, nil); code != 200 || out["name"] != "Desktop 2" {
		t.Errorf("update: %d %v", code, out)
	}

	if code, out, _ = a.do("POST", "/api/password", map[string]string{"current": "wrong-password", "new": "a-new-long-password"}, nil); code != 403 {
		t.Errorf("password change with wrong current: %d %v", code, out)
	}
	if code, out, _ = a.do("POST", "/api/password", map[string]string{"current": "a-long-password", "new": "a-new-long-password"}, nil); code != 200 {
		t.Errorf("password change: %d %v", code, out)
	}

	if code, out, _ = a.do("DELETE", "/api/devices/"+id, nil, nil); code != 200 {
		t.Errorf("delete: %d %v", code, out)
	}
	if _, list := a.doList("/api/devices"); len(list) != 0 {
		t.Errorf("device still listed after delete: %v", list)
	}

	if code, _, _ = a.do("POST", "/api/logout", nil, nil); code != 200 {
		t.Errorf("logout: %d", code)
	}
	if code, _ := a.doList("/api/devices"); code != 401 {
		t.Errorf("devices after logout: %d, want 401", code)
	}

	if code, out, _ = a.do("POST", "/api/login", map[string]string{"username": "admin", "password": "a-long-password"}, nil); code != 401 {
		t.Errorf("old password still works: %d %v", code, out)
	}
	if code, out, _ = a.do("POST", "/api/login", map[string]string{"username": "ADMIN", "password": "a-new-long-password"}, nil); code != 200 {
		t.Errorf("login with the new password: %d %v", code, out)
	}
}

func TestLoginRateLimit(t *testing.T) {
	a := newTestApp(t, []int{9})
	a.setup()
	a.do("POST", "/api/logout", nil, nil)

	var last int
	for i := 0; i < 5; i++ {
		last, _, _ = a.do("POST", "/api/login", map[string]string{"username": "admin", "password": "wrong-password"}, nil)
		if last != 401 {
			t.Fatalf("attempt %d: %d, want 401", i+1, last)
		}
	}
	code, out, hdr := a.do("POST", "/api/login", map[string]string{"username": "admin", "password": "a-long-password"}, nil)
	if code != 429 || out["error"] != "rate_limited" || hdr.Get("Retry-After") == "" {
		t.Errorf("after 5 failures: %d %v retry-after=%q", code, out, hdr.Get("Retry-After"))
	}
}

func TestRequestProtections(t *testing.T) {
	a := newTestApp(t, []int{9})
	a.setup()

	code, out, _ := a.do("POST", "/api/devices", map[string]any{"name": "x", "mac": "aa:bb:cc:dd:ee:11"}, map[string]string{"Sec-Fetch-Site": "cross-site"})
	if code != 403 || out["error"] != "forbidden_origin" {
		t.Errorf("cross-site request: %d %v", code, out)
	}
	code, out, _ = a.do("POST", "/api/devices", map[string]any{"name": "x", "mac": "aa:bb:cc:dd:ee:11"}, map[string]string{"Origin": "https://evil.example.com"})
	if code != 403 {
		t.Errorf("foreign origin: %d %v", code, out)
	}
	code, out, _ = a.do("POST", "/api/devices", map[string]any{"name": "x", "mac": "aa:bb:cc:dd:ee:11"}, map[string]string{"Content-Type": "text/plain"})
	if code != 415 {
		t.Errorf("non-JSON content type: %d %v", code, out)
	}
	big := strings.Repeat("a", 100<<10)
	code, _, _ = a.do("POST", "/api/devices", map[string]any{"name": big, "mac": "aa:bb:cc:dd:ee:11"}, nil)
	if code != 400 {
		t.Errorf("oversized body: %d, want 400", code)
	}

	res, err := a.client.Get(a.ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("index: %d", res.StatusCode)
	}
	for _, h := range []string{"Content-Security-Policy", "X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy"} {
		if res.Header.Get(h) == "" {
			t.Errorf("missing header %s", h)
		}
	}
	if csp := res.Header.Get("Content-Security-Policy"); strings.Contains(csp, "unsafe-inline") {
		t.Errorf("CSP allows inline code: %s", csp)
	}

	for _, p := range []string{"/app.js", "/app.css", "/healthz"} {
		r, err := a.client.Get(a.ts.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("%s: %d", p, r.StatusCode)
		}
	}
}

func TestSessionCookieFlags(t *testing.T) {
	a := newTestApp(t, []int{9})
	body, _ := json.Marshal(map[string]string{"token": a.srv.SetupToken(), "username": "admin", "password": "a-long-password"})
	req, _ := http.NewRequest("POST", a.ts.URL+"/api/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var found bool
	for _, c := range res.Cookies() {
		if c.Name == "wol_session" {
			found = true
			if !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
				t.Errorf("cookie flags: httpOnly=%v sameSite=%v", c.HttpOnly, c.SameSite)
			}
			if len(c.Value) < 40 {
				t.Errorf("session token too short: %d", len(c.Value))
			}
		}
	}
	if !found {
		t.Error("no session cookie set")
	}
}
